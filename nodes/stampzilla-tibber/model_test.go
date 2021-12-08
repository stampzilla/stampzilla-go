package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func newPrice(t1 time.Time, level string) Price {
	return Price{StartsAt: t1, Level: level}
}

func newPriceTotal(hour int, total float64) Price {
	t1 := time.Date(2020, 10, 10, hour, 0, 0, 0, time.UTC)
	return Price{StartsAt: t1.Truncate(time.Hour), Total: total}
}

func TestClearOld(t *testing.T) {
	t.Parallel()
	price1 := newPrice(time.Now(), "")
	price2 := newPrice(time.Now().Add(-1*time.Hour), "")

	prices := NewPrices()
	prices.Add(price1)
	prices.Add(price2)

	assert.Len(t, prices.prices, 2)
	prices.ClearOld()
	assert.Len(t, prices.prices, 1)
}

func TestCurrent(t *testing.T) {
	t.Parallel()
	price1 := newPrice(time.Now(), "now")
	price2 := newPrice(time.Now().Add(-1*time.Hour), "lasthour")
	price3 := newPrice(time.Now().Add(1*time.Hour), "nexthour")

	prices := NewPrices()
	prices.Add(price1)
	prices.Add(price2)
	prices.Add(price3)

	cur := prices.Current()
	assert.Equal(t, "now", cur.Level)
}

func TestHasTomorrowPrices(t *testing.T) {
	t.Parallel()
	price1 := newPrice(time.Now().Truncate(24*time.Hour).Add(24*time.Hour), "")

	t.Log(price1)

	prices := NewPrices()
	prices.Add(price1)

	assert.True(t, prices.HasTomorrowPricesYet())
}

func TestCalculateBestChargeHours(t *testing.T) {
	t.Parallel()
	tests := []struct {
		hoursPrice []struct {
			hour  int
			price float64
		}
		exp int
	}{
		{
			hoursPrice: []struct {
				hour  int
				price float64
			}{
				{0, 0},
				{1, 10},
				{2, 0},
				{3, 1},
				{4, 10},
				{5, 0},
				{6, 0},
				{7, 0},
				{8, 0},
				{9, 5},
				{10, 5},
				{11, 5},
				{12, 0},
				{13, 0},
				{14, 0},
				{15, 0},
				{16, 0},
			},
			exp: 11,
		},
		{
			hoursPrice: []struct {
				hour  int
				price float64
			}{
				{0, 0},
				{1, 0},
				{2, 0},
				{3, 0},
				{4, 0},
				{5, 1},
				{6, 1},
				{7, 1},
				{8, 1},
				{9, 1},
				{10, 1},
				{11, 1},
				{12, 1},
				{13, 1},
				{14, 1},
				{15, 1},
				{16, 1},
			},
			exp: 0,
		},
		{
			hoursPrice: []struct {
				hour  int
				price float64
			}{
				{0, 10},
				{1, 10},
				{2, 10},
				{3, 0},
				{4, 0},
				{5, 1},
				{6, 1},
				{7, 10},
				{8, 1},
				{9, 1},
				{10, 1},
				{11, 1},
				{12, 1},
				{13, 10},
				{14, 10},
				{15, 10},
				{16, 10},
			},
			exp: 3,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("expected %d", tt.exp), func(t *testing.T) {
			t.Parallel()
			prices := NewPrices()
			for _, v := range tt.hoursPrice {
				prices.Add(newPriceTotal(v.hour, v.price))
			}
			cheapestStartTime := prices.calculateBestChargeHours(6 * time.Hour)

			t.Log("cheapestStartTime", cheapestStartTime)
			expected := time.Date(2020, 10, 10, tt.exp, 0, 0, 0, time.UTC)
			assert.Equal(t, expected, cheapestStartTime)
		})
	}
}

func TestCheapestSingleHour(t *testing.T) {
	t.Parallel()
	tests := []struct {
		hoursPrice []struct {
			hour  int
			price float64
		}
		exp int
	}{

		{
			hoursPrice: []struct {
				hour  int
				price float64
			}{
				{0, 30},
				{1, 29},
				{2, 28},
				{3, 27},
				{4, 25},
				{5, 15},
				{6, 2},
				{7, 11},
				{8, 12},
				{9, 13},
				{10, 14},
				{11, 1},
				{12, 16},
				{13, 2},
				{14, 18},
				{15, 19},
				{16, 20},
			},
			exp: 11,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("expected %d", tt.exp), func(t *testing.T) {
			t.Parallel()
			prices := NewPrices()
			for _, v := range tt.hoursPrice {
				prices.Add(newPriceTotal(v.hour, v.price))
			}
			cheapestStartTime := prices.calculateBestChargeHours(1 * time.Hour)
			expected := time.Date(2020, 10, 10, tt.exp, 0, 0, 0, time.UTC)
			assert.Equal(t, expected, cheapestStartTime)
		})
	}
}
