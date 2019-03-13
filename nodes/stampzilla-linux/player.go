package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
)

type Player struct {
	Config PlayerConfig
	Name   string

	command chan string
	ctx     context.Context
	cancel  func()
	wg      sync.WaitGroup
}

type PlayerConfig struct {
	Shuffle  bool     `json:"shuffle"`
	Mode     string   `json:"mode"`
	Dir      string   `json:"dir"`
	Playlist []string `json:"playlist"`
}

var players = make(map[string]*Player, 0)

func startPlayers() {
	for name, config := range config.Players {
		player := &Player{
			Config:  config,
			Name:    name,
			command: make(chan string),
		}

		players[name] = player
		player.start()
	}
}

func restartPlayers() {
	for _, player := range players {
		player.stop()
	}
	players = make(map[string]*Player, 0)

	startPlayers()
}

func commandPlayer(player string, state bool) error {
	if p, ok := players[player]; ok {
		if state {
			p.command <- "play"
		} else {
			p.command <- "stop"
		}
		return nil
	}

	return fmt.Errorf("Player was not found")
}

func (player *Player) start() {
	player.ctx, player.cancel = context.WithCancel(context.Background())

	player.wg.Add(1)
	go player.Worker()
}

func (player *Player) stop() {
	player.cancel()
	player.wg.Wait()
}

func (player *Player) Worker() {
	defer player.wg.Done()
	defer logrus.Warnf("Worker %s EXIT", player.Name)

	dev := &devices.Device{
		Name:   "Player " + player.Name,
		ID:     devices.ID{ID: "player:" + player.Name},
		Online: true,
		Traits: []string{"OnOff"},
		State: devices.State{
			"on":   false,
			"file": "",
		},
	}
	n.AddOrUpdate(dev)

	for {
		// Not playing... (wait for shudown or command)
		select {
		case <-player.ctx.Done():
			return
		case cmd := <-player.command:
			if cmd != "play" {
				continue
			}
		}

		done, streamer, err := player.playFile("songs/jingle.mp3")
		if err != nil {
			logrus.Errorf("Playback failed: %s", err.Error())
		}

		dev.State["on"] = true
		n.AddOrUpdate(dev)

		// Playing.. (wait for shutdown, command or playback done)
	L:
		for {
			select {
			case <-player.ctx.Done():
				streamer.Close()
				return
			case cmd := <-player.command:
				if cmd != "stop" {
					continue
				}
				streamer.Close()
			case <-done:
				dev.State["on"] = false
				n.AddOrUpdate(dev)

				break L
			}
		}
		logrus.Info("Playback done")
	}
}

func (player *Player) playFile(file string) (chan struct{}, beep.StreamSeekCloser, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, nil, err
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		return nil, nil, err
	}

	sr := beep.SampleRate(44100)
	speaker.Init(sr, sr.N(time.Second/10))

	resampled := beep.Resample(4, format.SampleRate, sr, streamer)

	done := make(chan struct{})
	speaker.Play(beep.Seq(resampled, beep.Callback(func() {
		streamer.Close()
		done <- struct{}{}
	})))

	return done, streamer, nil
}
