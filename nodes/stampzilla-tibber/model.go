package main

import (
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
)

type Response struct {
	Data struct {
		Viewer struct {
			Homes []struct {
				CurrentSubscription CurrentSubscription `json:"currentSubscription"`
			} `json:"homes"`
		} `json:"viewer"`
	} `json:"data"`
}

type CurrentSubscription struct {
	PriceInfo *PriceInfo `json:"priceInfo"`
}

type PriceInfo struct {
	Current  Price   `json:"current"`
	Today    []Price `json:"today"`
	Tomorrow []Price `json:"tomorrow"`
}

type Price struct {
	Level    string    `json:"level"`
	Total    float64   `json:"total"`
	Energy   float64   `json:"energy"`
	Tax      float64   `json:"tax"`
	Currency string    `json:"currency"`
	StartsAt time.Time `json:"startsAt"`
}

type Prices struct {
	prices              map[time.Time]Price
	mutex               sync.RWMutex
	cheapestChargeStart time.Time
}

func NewPrices() *Prices {
	return &Prices{
		prices: make(map[time.Time]Price),
	}
}

func (p *Prices) Add(price Price) {
	p.mutex.Lock()
	p.prices[price.StartsAt] = price
	p.mutex.Unlock()
}

func (p *Prices) SetCheapestChargeStart(t time.Time) {
	p.mutex.Lock()
	p.cheapestChargeStart = t
	p.mutex.Unlock()
}

func (p *Prices) CheapestChargeStart() time.Time {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.cheapestChargeStart
}

func (p *Prices) Last() time.Time {
	ss := make([]time.Time, len(p.prices))
	i := 0
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for t := range p.prices {
		ss[i] = t
		i++
	}
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].After(ss[j])
	})

	return ss[0]
}

func (p *Prices) Current() *Price {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for t, price := range p.prices {
		if inTimeSpan(t, t.Add(60*time.Minute), time.Now()) {
			return &price
		}
	}

	return nil
}

func (p *Prices) HasTomorrowPricesYet() bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for t := range p.prices {
		if time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour).Equal(t.Truncate(24 * time.Hour)) {
			return true
		}
	}

	return false
}

func (p *Prices) calculateBestChargeHours(dur time.Duration) time.Time {
	dur = dur.Round(time.Hour)
	sumPriceForTimePlusDur := make(map[time.Time]float64)
	last := p.Last().Add(time.Hour * 1)
	p.mutex.Lock()
	for t := range p.prices {
		var d time.Duration
		priceSum := 0.0
		for d < dur {
			priceSum += p.prices[t.Add(d)].Total
			d += time.Hour
		}
		if t.Add(dur).Before(last) || t.Add(dur).Equal(last) {
			sumPriceForTimePlusDur[t] = priceSum
		}
	}
	p.mutex.Unlock()

	type kv struct {
		Key   time.Time
		Value float64
	}

	ss := make([]kv, len(sumPriceForTimePlusDur))
	i := 0
	for k, v := range sumPriceForTimePlusDur {
		ss[i] = kv{Key: k, Value: v}
		i++
	}
	sort.Slice(ss, func(i, j int) bool {
		if ss[i].Value == ss[j].Value {
			return ss[i].Key.Before(ss[j].Key)
		}
		return ss[i].Value < ss[j].Value
	})

	for _, v := range ss {
		logrus.Debug(v)
	}
	return ss[0].Key
}

func (p *Prices) ClearOld() {
	p.mutex.Lock()
	for t := range p.prices {
		if time.Now().Add(time.Hour * -1).After(t) {
			delete(p.prices, t)
		}
	}
	p.mutex.Unlock()
}

func inTimeSpan(start, end, check time.Time) bool {
	if start.Before(end) {
		return !check.Before(start) && !check.After(end)
	}
	if start.Equal(end) {
		return check.Equal(start)
	}
	return !start.After(check) || !end.Before(check)
}

func (p *Prices) State() devices.State {
	state := make(devices.State)
	current := p.Current()
	state["price"] = current.Total

	state["carChargeStart"] = false
	now := time.Now()
	if p.CheapestChargeStart().Before(now) || p.CheapestChargeStart().Equal(now) {
		state["carChargeStart"] = true
	}

	state["priceExpensive"] = false
	if current.Level == "EXPENSIVE" || current.Level == "VERY_EXPENSIVE" {
		state["priceExpensive"] = true
	}

	return state
}
