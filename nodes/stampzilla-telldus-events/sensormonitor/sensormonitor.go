package sensormonitor

import (
	"fmt"
	"log"
	"time"

	"github.com/stampzilla/stampzilla-go/pkg/notifier"
)

type sensor struct {
	alive    time.Time
	notified bool
}

type Monitor struct {
	notify  *notifier.Notify
	sensors map[int]*sensor
	alive   chan (int)
	timeout string
}

func New(notify *notifier.Notify) *Monitor {
	sm := &Monitor{
		sensors: make(map[int]*sensor),
		alive:   make(chan (int)),
		notify:  notify,
		timeout: "1h",
	}
	return sm
}

func (s Monitor) SetTimeout(timeout string) {
	s.timeout = timeout
}

func (s *Monitor) Start() {
	go s.worker()
}

func (s *Monitor) worker() {
	ticker := time.NewTicker(5 * time.Minute)
	for {
		select {
		case id := <-s.alive:
			s.updateAlive(id)
		case <-ticker.C:
			s.CheckDead("1h")
		}
	}
}

func (s *Monitor) updateAlive(id int) {
	if val, ok := s.sensors[id]; ok {
		val.alive = time.Now()
		val.notified = false
	}
	s.sensors[id] = &sensor{alive: time.Now()}
}

// CheckDead loops through list of sensors and sending notifications if duration has passed since last Alive.
func (s *Monitor) CheckDead(dur string) {
	duration, err := time.ParseDuration(dur)
	if err != nil {
		log.Println(err)
		return
	}

	for id, sensor := range s.sensors {
		if sensor.alive.Add(duration).Before(time.Now()) {
			if sensor.notified {
				continue
			}
			err := fmt.Sprintf("Sensor %d has not been updated in %s", id, duration)
			if s.notify != nil {
				s.notify.Error(err)
				sensor.notified = true
			}
		}
	}
}

func (s *Monitor) Alive(id int) {
	s.alive <- id
}
