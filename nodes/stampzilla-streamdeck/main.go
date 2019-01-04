package main

import (
	"encoding/json"
	"fmt"
	"sync"

	"./streamdeck"
	"github.com/davecgh/go-spew/spew"
	"github.com/llgcode/draw2d"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/pkg/node"
	"github.com/stampzilla/stampzilla-go/pkg/websocket"
)

var ICON_SIZE = 72

type state struct {
	decks  []*streamdeck.StreamDeck
	page   []string
	render chan struct{}
	config *config
	sync.Mutex
}

func main() {
	client := websocket.New()
	node := node.New(client)
	node.Type = "streamdeck"

	config := &config{}
	state := &state{
		render: make(chan struct{}),
	}

	node.OnConfig(updatedConfig(node, state, config))
	//node.OnRequestStateChange(func(state models.DeviceState, device *models.Device) error {
	//spew.Dump(config)
	//return nil
	//})

	err := node.Connect()
	if err != nil {
		logrus.Error(err)
		return
	}

	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
	logrus.SetReportCaller(false)

	//node.OnShutdown(func() {
	//})

	// ---------------------------

	draw2d.SetFontFolder("./fonts")
	go state.worker()

	node.Wait()
	node.Client.Wait()
}

func updatedConfig(node *node.Node, state *state, config *config) func(json.RawMessage) error {
	return func(data json.RawMessage) error {
		var configString string
		err := json.Unmarshal(data, &configString)
		if err != nil {
			return err
		}

		config.Lock()
		defer config.Unlock()
		err = json.Unmarshal([]byte(configString), config)
		if err != nil {
			return err
		}

		state.Lock()
		state.config = config
		state.Unlock()

		logrus.Warn("Received new config")
		select {
		case state.render <- struct{}{}:
		default:
		}

		return nil
	}
}

func (state *state) worker() {
	state.decks = streamdeck.FindDecks()

	for index, deck := range state.decks {
		err := deck.Open()
		if err != nil {
			logrus.Error(err)
		}

		deck.Reset()
		deck.SetBrightness(30)

		deck.OnKeyDown(func(key int, s bool) {
			fmt.Printf("Key %d pressed\n", key)

			state.Lock()
			page := "default"
			if index < len(state.page) {
				page = state.page[index]
			}

			pageConfig, ok := state.config.Pages[page]
			state.Unlock()

			if ok && key >= len(pageConfig.Keys) {
				return
			}

			spew.Dump(pageConfig.Keys[key])
		})
	}

	for {
		for index, deck := range state.decks {
			state.Lock()
			page := "default"
			if index < len(state.page) {
				page = state.page[index]
			}

			pageConfig, ok := state.config.Pages[page]
			state.Unlock()
			if !ok {
				continue
			}

			logrus.Infof("Draw page %s to deck %d", page, index)
			for key, keyConfig := range pageConfig.Keys {
				logrus.Infof("Draw key %d to deck %d (%s)", key, index, keyConfig.Name)
				drawTempToKey(deck, keyConfig.Name, float32(key), key)
			}
		}

		//drawFactionToKey(decks[0], factionImage, 0)

		//decks[0].OnKeyDown(func(key int, state bool) {
		//fmt.Printf("Key %d pressed\n", key)
		//})
		//decks[0].OnKeyUp(func(key int, state bool) {
		//fmt.Printf("Key %d released\n", key)
		//})
		//decks[0].OnKeyChange(func(key int, state bool) {
		//fmt.Printf("Key %d changed\n", key)
		//})

		//for i := 0; i < 15; i++ {
		//drawTempToKey(decks[0], "Key "+strconv.Itoa(i), float32(i), i)
		//}

		<-state.render
	}
}
