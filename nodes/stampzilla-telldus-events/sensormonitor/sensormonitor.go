package sensormonitor

import (
	"fmt"
	"log"
	"time"

	"github.com/stampzilla/stampzilla-go/pkg/notifier"
)

type Monitor struct {
	notify  *notifier.Notify
	sensors map[int]time.Time
	alive   chan (int)
}

func New(notify *notifier.Notify) *Monitor {
	sm := &Monitor{
		sensors: make(map[int]time.Time),
		alive:   make(chan (int)),
		notify:  notify,
	}
	return sm
}

func (s Monitor) Start() {
	go s.worker()
}

func (s Monitor) worker() {
	ticker := time.NewTicker(5 * time.Minute)
	for {
		select {
		case id := <-s.alive:
			s.sensors[id] = time.Now()
		case <-ticker.C:
			s.CheckDead("1h")
		}
	}
}
func (s Monitor) CheckDead(dur string) {
	duration, err := time.ParseDuration(dur)
	if err != nil {
		log.Println(err)
		return
	}

	for id, t := range s.sensors {
		if t.Add(duration).Before(time.Now()) {
			err := fmt.Sprintf("Sensor %d has not been updated in %s", id, duration)
			if s.notify != nil {
				s.notify.Error(err)
			}
		}
	}
}

func (s Monitor) Alive(id int) {
	s.alive <- id
}
