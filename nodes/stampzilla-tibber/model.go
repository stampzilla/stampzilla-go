package main

import (
	"encoding/json"
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
)

type Response struct {
	Data struct {
		Viewer struct {
			Home struct {
				Features struct {
					RealTimeConsumptionEnabled bool
				}
			}
			WebsocketSubscriptionUrl string
			Homes                    []struct {
				CurrentSubscription CurrentSubscription `json:"currentSubscription"`
			} `json:"homes"`
		} `json:"viewer"`
	} `json:"data"`
}

type WsMsg struct {
	Type    string          `json:"type"`
	ID      string          `json:"id,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type WsPayloadQuery struct {
	Query string `json:"query"`
}

type DataPayload struct {
	Data struct {
		LiveMeasurement struct {
			Timestamp              time.Time `json:"timestamp"`
			Power                  float64   `json:"power"`
			CurrentL1              float64   `json:"currentL1"`
			CurrentL2              float64   `json:"currentL2"`
			CurrentL3              float64   `json:"currentL3"`
			MaxPower               float64   `json:"maxPower"`
			AccumulatedConsumption float64   `json:"accumulatedConsumption"`
			AccumulatedCost        float64   `json:"accumulatedCost"`
		} `json:"liveMeasurement"`
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
	tot := 0.0
	totCnt := 0.0
	p.mutex.RLock()
	for _, v := range p.prices {
		if t.Add(time.Hour*-24).After(v.Time) || v.Time.After(t) {
			continue
		}
		tot += v.Total
		totCnt += 1.0
	}
	p.mutex.RUnlock()
	average := tot / totCnt

	switch {
	case total >= average*1.20:
		lvl = 3 // HIGH
	case total <= average*0.90:
		lvl = 1 // LOW
	default:
		lvl = 2 // NORMAL
	}

	diff = total / average

	//fmt.Println("avg: ", average)
	//fmt.Println("price: ", total)
	//fmt.Println("diff: ", diff)
	//fmt.Println("lvl: ", lvl)
	return
}
