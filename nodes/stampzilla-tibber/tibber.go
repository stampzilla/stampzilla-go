package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/pkg/node"
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

	for _, p := range prices.Today {
		pricesStore.Add(p)
	}

	for _, p := range prices.Tomorrow {
		pricesStore.Add(p)
	}

	pricesStore.ClearOld()

	if pricesStore.HasTomorrowPricesYet() {
		cheapestStartTime := pricesStore.calculateBestChargeHours(config.carChargeDuration)
		pricesStore.SetCheapestChargeStart(cheapestStartTime)

		cheapestHour := pricesStore.calculateBestChargeHours(1 * time.Hour)
		pricesStore.SetCheapestHour(cheapestHour)
	}

	node.UpdateState("1", pricesStore.State())
}

func fetchPrices(token string) (*PriceInfo, error) {
	query := `
{
  viewer {
    homes {
      currentSubscription{
        priceInfo{
          today {
            total
            startsAt
            level
          }
          tomorrow {
            total
            startsAt
            level
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
	return response.Data.Viewer.Homes[0].CurrentSubscription.PriceInfo, nil
}
