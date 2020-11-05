package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
)

type Player struct {
	Config PlayerConfig
	Name   string

	playlist []string
	command  chan string
	ctx      context.Context
	cancel   func()
	wg       sync.WaitGroup
}

type PlayerConfig struct {
	Shuffle  bool     `json:"shuffle"`
	Mode     string   `json:"mode"`
	Dir      string   `json:"dir"`
	Playlist []string `json:"playlist"`
}

var (
	players = make(map[string]*Player, 0)
	sr      = beep.SampleRate(44100)
)

func init() {
	speaker.Init(sr, sr.N(time.Second/10))
}

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

func (player *Player) makePlaylist() []string {
	playlist := make([]string, 0)
	files, err := filepath.Glob(player.Config.Dir + "/*.mp3")
	if err != nil {
		log.Fatal(err)
	}

	if player.Config.Shuffle {
		r := rand.New(rand.NewSource(time.Now().Unix()))
		for _, i := range r.Perm(len(files)) {
			playlist = append(playlist, files[i])
		}
	} else {
		for _, f := range files {
			playlist = append(playlist, f)
		}
	}

	return playlist
}

func (player *Player) Worker() {
	logrus.Debugf("Worker %s START", player.Name)
	defer player.wg.Done()
	defer logrus.Debugf("Worker %s EXIT", player.Name)

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
		if dev.State["on"] != true {
			// Not playing... (wait for shudown or command)
			select {
			case <-player.ctx.Done():
				return
			case cmd := <-player.command:
				if cmd != "play" {
					continue
				}
			}
		}

		if len(player.playlist) == 0 {
			logrus.Info("Making new playlist")
			player.playlist = player.makePlaylist()
		}
		if len(player.playlist) == 0 {
			logrus.Warn("Playlist is empty, skipping")
			newState := make(devices.State)
			newState["on"] = false
			n.UpdateState(dev.ID.ID, newState)
			continue
		}

		nextSong := player.playlist[0]
		player.playlist = player.playlist[1:]

		done, streamer, err := player.playFile(nextSong)
		if err != nil {
			logrus.Errorf("Playback failed: %s", err.Error())
			continue
		}

		newState := make(devices.State)
		newState["on"] = true
		n.UpdateState(dev.ID.ID, newState)

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
				if streamer != nil {
					streamer.Close()
				}
			case <-done:
				if player.Config.Mode == "single" {
					newState := make(devices.State)
					newState["on"] = false
					n.UpdateState(dev.ID.ID, newState)
				}

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

	resampled := beep.Resample(4, format.SampleRate, sr, streamer)

	done := make(chan struct{})
	speaker.Play(beep.Seq(resampled, beep.Callback(func() {
		streamer.Close()
		done <- struct{}{}
	})))

	return done, streamer, nil
}
