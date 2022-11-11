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
	PriceRating *PriceRating `json:"priceRating"`
}

type PriceRating struct {
	Hourly Hourly `json:"hourly"`
}

type Hourly struct {
	Entries []Price `json:"entries"`
}

type Price struct {
	Level    string    `json:"level"`
	Total    float64   `json:"total"`
	Energy   float64   `json:"energy"`
	Tax      float64   `json:"tax"`
	Currency string    `json:"currency"`
	Time     time.Time `json:"time"`
}

type Prices struct {
	prices              map[time.Time]Price
	mutex               sync.RWMutex
	cheapestChargeStart time.Time
	cheapestHour        time.Time
	lastCalculated      time.Time
}

func NewPrices() *Prices {
	return &Prices{
		prices: make(map[time.Time]Price),
	}
}

func (p *Prices) Add(price Price) {
	p.mutex.Lock()
	p.prices[price.Time] = price
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
func (p *Prices) SetCheapestHour(t time.Time) {
	p.mutex.Lock()
	p.cheapestHour = t
	p.mutex.Unlock()
}

func (p *Prices) CheapestHour() time.Time {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.cheapestHour
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

func (p *Prices) calculateCheapestHour(from, to time.Time) time.Time {
	ss := []Price{}
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for _, t := range p.prices {
		if t.Time.Before(from) || t.Time.After(to) {
			continue
		}
		ss = append(ss, t)
	}
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Total < ss[j].Total
	})

	return ss[0].Time
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

func (p *Prices) SetLastCalculated(t time.Time) {
	p.mutex.Lock()
	p.lastCalculated = t
	p.mutex.Unlock()
}
func (p *Prices) LastCalculated() time.Time {
	p.mutex.Lock()
	t := p.lastCalculated
	p.mutex.Unlock()
	return t
}

// calculateCheapestTimes tries to find the cheapest <count> times with the constraint that it needs to be at least <hoursBetween> the X times.
func (p *Prices) calculateCheapestTimes(start time.Time, count int, hoursBetween int) []time.Time {
	// TODO use this function. perhaps new object in config that we can configure count and hoursBetween Cheapest<count>Hours<between>
	ret := []time.Time{}
	findCheapest := func(from, to time.Time) time.Time {
		ss := []Price{}
		p.mutex.Lock()
		for _, t := range p.prices {
			if t.Time.Before(from) || t.Time.After(to) {
				continue
			}
			ss = append(ss, t)
		}
		p.mutex.Unlock()

		if len(ss) == 0 {
			return time.Time{}
		}

		sort.Slice(ss, func(i, j int) bool {
			if ss[i].Total == ss[j].Total {
				return ss[j].Time.After(ss[i].Time)
			}
			return ss[i].Total < ss[j].Total
		})
		return ss[0].Time
	}

	var next = findCheapest(start, start.Add(time.Hour*time.Duration(hoursBetween)))
	ret = append(ret, next)
	for i := 0; i < count-1; i++ {
		next = findCheapest(
			next.Add(time.Hour*time.Duration(hoursBetween)),
			next.Add(time.Hour*time.Duration(hoursBetween*2)),
		)
		ret = append(ret, next)
	}

	return ret
}

func (p *Prices) calculateBestChargeHours(start time.Time, dur time.Duration) time.Time {
	dur = dur.Round(time.Hour)
	sumPriceForTimePlusDur := make(map[time.Time]float64)
	last := p.Last().Add(time.Hour * 1)
	p.mutex.Lock()
	for t := range p.prices {
		if start.After(t) {
			continue // only calculate on future prices not the past 24h
		}
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
		if time.Now().Add(time.Hour * -24).After(t) {
			delete(p.prices, t)
		}
	}
	p.mutex.Unlock()
}
func (p *Prices) SortedByTime() []Price {
	p.mutex.Lock()
	prices := make([]Price, len(p.prices))
	i := 0
	for _, v := range p.prices {
		prices[i] = v
		i++
	}
	p.mutex.Unlock()

	sort.Slice(prices, func(i, j int) bool {
		return prices[i].Time.Before(prices[j].Time)
	})

	return prices
}

func isSameHourAndDay(t1, t2 time.Time) bool {
	return t1.Truncate(1 * time.Hour).Equal(t2.Truncate(1 * time.Hour))
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

	state["cheapestHour"] = false
	if isSameHourAndDay(now, p.CheapestHour()) {
		state["cheapestHour"] = true
	}

	diff, lvl := p.calculateLevel(current.Time, current.Total)
	state["priceLevel"] = lvl
	state["priceDiff"] = diff

	if lvl == 3 {
		state["priceExpensive"] = true
	} else {
		state["priceExpensive"] = false
	}

	return state
}

func (p *Prices) calculateLevel(t time.Time, total float64) (diff float64, lvl int) {
	findMinMax := func(s time.Time, min bool) float64 {
		ss := []Price{}
		p.mutex.Lock()
		for _, t := range p.prices {
			if t.Time.Before(s) { // Only calculate min/max against future prices.
				continue
			}
			ss = append(ss, t)
		}
		p.mutex.Unlock()

		if len(ss) == 0 {
			return -1.0
		}

		sort.Slice(ss, func(i, j int) bool {
			if ss[i].Total == ss[j].Total {
				return ss[j].Time.After(ss[i].Time)
			}
			if min {
				return ss[i].Total < ss[j].Total
			} else {
				return ss[i].Total > ss[j].Total
			}
		})
		return ss[0].Total
	}

	min := findMinMax(t, true)
	max := findMinMax(t, false)

	if min == -1.0 || max == -1.0 {
		lvl = 2
		return
	}

	diff = max - min

	switch {
	case total >= max-diff*0.25:
		lvl = 3 // HIGH
	case total <= min+diff*0.25:
		lvl = 1 // LOW
	default:
		lvl = 2 // NORMAL
	}

	return
}
