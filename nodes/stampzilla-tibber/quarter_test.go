package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNextDelay(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Stockholm")
	now := time.Date(2026, 11, 25, 17, 7, 30, 0, loc)
	assert.Equal(t, 7*time.Minute+30*time.Second, nextDelay(now, 15*time.Minute))
	assert.Equal(t, 52*time.Minute+30*time.Second, nextDelay(now, time.Hour))
	assert.Equal(t, 52*time.Minute+30*time.Second, nextDelay(now, 0))
}

func TestCurrentQuarterHourly(t *testing.T) {
	prices := NewPrices()
	prices.SetSlot(15 * time.Minute)
	start := time.Now().Truncate(15 * time.Minute)
	for i := -4; i < 4; i++ {
		prices.Add(Price{Time: start.Add(time.Duration(i) * 15 * time.Minute), Total: float64(i)})
	}
	c := prices.Current()
	assert.NotNil(t, c)
	assert.Equal(t, start, c.Time)
	assert.Equal(t, 0.0, c.Total)
}

func TestSetSlotClearsOnChange(t *testing.T) {
	prices := NewPrices()
	prices.Add(Price{Time: time.Now().Truncate(time.Hour)})
	prices.SetSlot(time.Hour)
	assert.Len(t, prices.prices, 1)
	prices.SetSlot(15 * time.Minute)
	assert.Len(t, prices.prices, 0)
}

func TestCalculateBestChargeQuarterHourly(t *testing.T) {
	prices := NewPrices()
	prices.SetSlot(15 * time.Minute)
	start := time.Date(2026, 11, 25, 0, 0, 0, 0, time.UTC)
	// 4h of 15m prices, all 10 except a cheap 30 min window at 02:15-02:45
	for i := 0; i < 16; i++ {
		total := 10.0
		if i == 9 || i == 10 {
			total = 1.0
		}
		prices.Add(Price{Time: start.Add(time.Duration(i) * 15 * time.Minute), Total: total})
	}
	got := prices.calculateBestChargeHours(start, 30*time.Minute)
	assert.Equal(t, start.Add(2*time.Hour+15*time.Minute), got)
}

func TestFetchPricesQuarterHourly(t *testing.T) {
	var query string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, 4096)
		n, _ := r.Body.Read(b)
		query = string(b[:n])
		w.Write([]byte(`{"data":{"viewer":{"homes":[{"currentSubscription":{"priceInfo":{
			"today":[{"total":1.5,"energy":1,"tax":0.5,"startsAt":"2026-11-25T00:00:00.000+01:00","level":"NORMAL","currency":"SEK"},
			         {"total":2.5,"energy":2,"tax":0.5,"startsAt":"2026-11-25T00:15:00.000+01:00","level":"HIGH","currency":"SEK"}],
			"tomorrow":[{"total":0.5,"energy":0.4,"tax":0.1,"startsAt":"2026-11-26T00:00:00.000+01:00","level":"CHEAP","currency":"SEK"}]}}}]}}}`))
	}))
	defer srv.Close()
	old := httpClient
	httpClient = &http.Client{Transport: rewriteTransport{srv.URL}}
	defer func() { httpClient = old }()

	prices, err := fetchPrices("token", ResolutionQuarterHourly)
	assert.NoError(t, err)
	assert.Contains(t, query, "priceInfo(resolution: QUARTER_HOURLY)")
	assert.Len(t, prices, 3)
	assert.Equal(t, 2.5, prices[1].Total)
	assert.Equal(t, "SEK", prices[1].Currency)
	assert.True(t, prices[1].Time.Equal(time.Date(2026, 11, 24, 23, 15, 0, 0, time.UTC)))
}

func TestUpdatedConfigResolution(t *testing.T) {
	c := NewConfig()
	assert.NoError(t, updatedConfig(c)([]byte(`{"carChargeDuration":"6h"}`)))
	assert.Equal(t, ResolutionQuarterHourly, c.Resolution)
	assert.Equal(t, 15*time.Minute, c.slot)

	assert.NoError(t, updatedConfig(c)([]byte(`{"carChargeDuration":"6h","resolution":"HOURLY"}`)))
	assert.Equal(t, time.Hour, c.slot)

	assert.Error(t, updatedConfig(c)([]byte(`{"carChargeDuration":"6h","resolution":"5m"}`)))
}
