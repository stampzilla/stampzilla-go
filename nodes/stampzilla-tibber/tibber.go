package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/pkg/node"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func fetchAndCalculate(config *Config, node *node.Node) {
	prices, err := fetchPrices(config.Token)
	if err != nil {
		logrus.Error(err)
		curr := pricesStore.Current()
		if curr != nil { // update state if we have cached hour and API call fails.
			logrus.Errorf("using cached value for Price: %#v", curr)
			node.UpdateState("1", pricesStore.State())
		}
		return
	}

	now := time.Now()
	for _, p := range prices {
		if now.Add(time.Hour * -24).After(p.Time) {
			continue
		}
		pricesStore.Add(p)
	}

	pricesStore.ClearOld()

	if pricesStore.HasTomorrowPricesYet() && !pricesStore.LastCalculated().Truncate(24*time.Hour).Equal(now.Truncate(24*time.Hour)) {
		c := pricesStore.Current()

		cheapestStartTime := pricesStore.calculateBestChargeHours(c.Time, config.carChargeDuration)
		pricesStore.SetCheapestChargeStart(cheapestStartTime)

		cheapestHour := pricesStore.calculateCheapestHour(
			now.Truncate(24*time.Hour).Add(time.Hour*18),                  // today plus 18 is 19:00 CET
			now.Add(24*time.Hour).Truncate(24*time.Hour).Add(time.Hour*8), // tomorrow 00 plus 8 hours is 09:00 CET
		)
		pricesStore.SetCheapestHour(cheapestHour)
		logrus.Infof("cheapestHour is: %s", cheapestHour)

		pricesStore.SetLastCalculated(time.Now())
	}

	node.UpdateState("1", pricesStore.State())
}

func fetchPrices(token string) ([]Price, error) {
	query := `
{
  viewer {
    websocketSubscriptionUrl
    homes {
      currentSubscription{
        priceRating{
          hourly {
            entries {
              total
              time
              level
              energy
              difference
              tax
            }
          }
        }
      }
    }
  }
}`

	out, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	reqBody := fmt.Sprintf(`{"query":%s}`, out)

	req, err := http.NewRequest("POST", "https://api.tibber.com/v1-beta/gql", strings.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error fetching from tibber api status: %d", resp.StatusCode)
	}

	response := &Response{}
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return nil, err
	}
	if len(response.Data.Viewer.Homes) == 0 {
		return nil, fmt.Errorf("no homes found in response")
	}
	return response.Data.Viewer.Homes[0].CurrentSubscription.PriceRating.Hourly.Entries, nil
}
func getWsURL(token, homeID string) (string, error) {
	query := fmt.Sprintf(`
{
  viewer {
    websocketSubscriptionUrl
    home(id: "%s") {
      id
      features {
        realTimeConsumptionEnabled
      }
    }
  }
}`, homeID)

	out, err := json.Marshal(query)
	if err != nil {
		return "", err
	}
	reqBody := fmt.Sprintf(`{"query":%s}`, out)

	req, err := http.NewRequest("POST", "https://api.tibber.com/v1-beta/gql", strings.NewReader(reqBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("error fetching from tibber api status: %d", resp.StatusCode)
	}

	response := &Response{}
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return "", err
	}
	if !response.Data.Viewer.Home.Features.RealTimeConsumptionEnabled {
		return "", fmt.Errorf("RealTimeConsumptionEnabled not enabled for home %s", homeID)
	}

	return response.Data.Viewer.WebsocketSubscriptionUrl, nil
}

type updateStateFunc func(data *DataPayload)

func reconnectWS(ctx context.Context, u, token, homeID string, cb updateStateFunc) error {
	i := 0
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		connectTime := time.Now()
		err := connectWS(ctx, u, token, homeID, cb)
		connectedDuration := time.Since(connectTime)

		if connectedDuration.Seconds() < 1.0 && i < 30 { // we was connected for less than 1 sec. Lets backoff.
			i++
		} else {
			i = 0
		}

		sleepDuration := time.Second*10 + time.Duration(i)*time.Second
		if err != nil {
			logrus.Errorf("connectWs err %s reconnecting in %s", err, sleepDuration)
		}

		time.Sleep(sleepDuration)
	}
}

func connectWS(ctx context.Context, u, token, homeID string, cb updateStateFunc) error {

	logrus.Infof("connecting to tibber websocket: %s", u)
	c, _, err := websocket.Dial(ctx, u, &websocket.DialOptions{Subprotocols: []string{"graphql-transport-ws"}, HTTPHeader: http.Header{
		"User-Agent": {"stampzilla/2 stampzilla-tibber/2"},
	}})
	if err != nil {
		return err
	}
	defer c.Close(websocket.StatusNormalClosure, "shutting down")

	err = c.Write(ctx, websocket.MessageText, []byte(fmt.Sprintf(`{"type":"connection_init","payload":{"token":"%s"}}`, token)))
	if err != nil {
		return fmt.Errorf("error writing: %w", err)
	}

	_, msg, err := c.Read(ctx)
	if err != nil {
		return fmt.Errorf("error reading: %w", err)
	}

	resp := &WsMsg{}
	err = json.Unmarshal(msg, resp)
	if err != nil {
		return err
	}

	if resp.Type != "connection_ack" {
		return fmt.Errorf("no connection_ack")
	}

	payload, err := json.Marshal(WsPayloadQuery{
		Query: fmt.Sprintf("subscription {\n  liveMeasurement(homeId: \"%s\") {\n    timestamp\n    power\n    currentL1\n    currentL2\n    currentL3\n    maxPower\n   accumulatedConsumption\n accumulatedCost\n  }\n}", homeID),
	})
	if err != nil {
		return err
	}

	req := &WsMsg{
		ID:      uuid.New().String(),
		Type:    "subscribe",
		Payload: payload,
	}

	err = wsjson.Write(ctx, c, req)
	if err != nil {
		return fmt.Errorf("error writing: %w", err)
	}

	lastData := time.Now().Unix()
	go func() {
		defer c.Close(websocket.StatusNormalClosure, "no data timeout")
		for {
			t := time.Unix(atomic.LoadInt64(&lastData), 0)
			if time.Since(t) > time.Minute {
				return
			}
			time.Sleep(time.Second)
		}
	}()

	for {
		err = wsjson.Read(ctx, c, resp)
		if err != nil {
			return fmt.Errorf("error reading json: %w", err)
		}

		atomic.StoreInt64(&lastData, time.Now().Unix())
		data := &DataPayload{}
		err := json.Unmarshal(resp.Payload, data)
		if err != nil {
			return fmt.Errorf("error unmarshal json: %w", err)
		}
		if resp.Type != "next" {
			logrus.Warnf("expected response type 'next' got: %s payload: %s", resp.Type, string(resp.Payload))
			continue
		}
		cb(data)
	}
}
