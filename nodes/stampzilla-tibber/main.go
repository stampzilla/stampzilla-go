package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/v2/pkg/node"
)

var httpClient = &http.Client{
	Timeout: time.Second * 20,
}

var pricesStore = NewPrices()

func main() {
	start()
}

func start() {
	config := NewConfig()
	node := setupNode(config)
	wait := node.WaitForFirstConfig()

	err := node.Connect()
	if err != nil {
		logrus.Error(err)
		return
	}
	logrus.Info("Waiting for config from server")
	err = wait()
	if err != nil {
		logrus.Error(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-node.Stopped()
		cancel()
	}()
	go func() {
		err := reconnectWS(ctx, config.Token, config.HomeID, func(data *DataPayload) {
			node.UpdateState("1", devices.State{
				"current_W":        data.Data.LiveMeasurement.Power,
				"L1_A":             data.Data.LiveMeasurement.CurrentL1,
				"L2_A":             data.Data.LiveMeasurement.CurrentL2,
				"L3_A":             data.Data.LiveMeasurement.CurrentL3,
				"consumptionToday": data.Data.LiveMeasurement.AccumulatedConsumption,
				"costToday":        data.Data.LiveMeasurement.AccumulatedCost,
			})
		})
		if err != nil && !errors.Is(err, context.Canceled) {
			logrus.Warn("will not start websocket connection to tibber: ", err)
		}
	}()

	tickerLoop(config, node)
}

func setupNode(config *Config) *node.Node {
	node := node.New("tibber")
	node.OnConfig(updatedConfig(config))

	return node
}

func tickerLoop(config *Config, node *node.Node) {
	dev := &devices.Device{
		Name:   "electricity",
		Type:   "sensor",
		ID:     devices.ID{ID: "1"},
		Online: true,
		// Traits: []string{"SensorState"}, // TODO add SensorState to frontend
		State: make(devices.State),
	}

	node.AddOrUpdate(dev)

	failures := 0
	run := func() time.Duration {
		if err := fetchAndCalculate(config, node); err != nil {
			failures++
		} else {
			failures = 0
		}
		return scheduleDelay(failures, nextDelay(time.Now(), config.slot))
	}

	delay := run()
	timer := time.NewTimer(delay)
	logrus.Info("scheduling next run in ", delay)
	for {
		select {
		case <-timer.C:
			timer.Reset(run())

		case <-node.Stopped():
			timer.Stop()
			logrus.Info("stopping tibber node")
			return
		}
	}
}

func updatedConfig(config *Config) func(data json.RawMessage) error {
	return func(data json.RawMessage) error {
		err := json.Unmarshal(data, config)
		if err != nil {
			return err
		}

		dur, err := time.ParseDuration(config.CarChargeDuration)
		if err != nil {
			return fmt.Errorf("wrong duration format %s: %w", config.carChargeDuration, err)
		}
		config.carChargeDuration = dur

		switch config.Resolution {
		case "", ResolutionQuarterHourly:
			config.Resolution = ResolutionQuarterHourly
			config.slot = 15 * time.Minute
		case ResolutionHourly:
			config.slot = time.Hour
		default:
			return fmt.Errorf("unknown resolution %q, must be %s or %s", config.Resolution, ResolutionQuarterHourly, ResolutionHourly)
		}
		return nil
	}
}

const fetchRetryDelay = time.Minute

// scheduleDelay returns when to run the next fetch. After failed fetches we
// retry sooner than the next price period, with exponential backoff (1m, 2m, 4m...)
// capped at the regular schedule.
func scheduleDelay(failures int, regular time.Duration) time.Duration {
	if failures == 0 {
		return regular
	}
	retry := fetchRetryDelay << min(failures-1, 6)
	if retry >= regular {
		return regular
	}
	logrus.Infof("fetch failed %d time(s), retrying in %s", failures, retry)
	return retry
}

// nextDelay returns the duration until the start of the next price period.
func nextDelay(now time.Time, slot time.Duration) time.Duration {
	if slot <= 0 {
		slot = time.Hour
	}
	return now.Truncate(slot).Add(slot).Sub(now)
}
