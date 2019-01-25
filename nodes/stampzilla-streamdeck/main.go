package main

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/davecgh/go-spew/spew"
	"github.com/llgcode/draw2d"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-streamdeck/streamdeck"
	"github.com/stampzilla/stampzilla-go/pkg/node"
)

var ICON_SIZE = 72

type appState struct {
	decks  []*streamdeck.StreamDeck
	page   []string
	render chan struct{}
	node   *node.Node
	config *config
	sync.Mutex
}

var previous = devices.NewList()
var app = &appState{
	render: make(chan struct{}),
}

func main() {
	node := node.New("streamdeck")
	app.node = node

	config := &config{}

	node.OnConfig(updatedConfig(node, config))
	node.On("devices", onDevices)
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

	node.Subscribe("devices")

	//node.OnShutdown(func() {
	//})

	// ---------------------------

	draw2d.SetFontFolder("./fonts")
	go app.worker()

	node.Wait()
	node.Client.Wait()
}

func updatedConfig(node *node.Node, config *config) func(json.RawMessage) error {
	return func(data json.RawMessage) error {
		config.Lock()
		defer config.Unlock()

		err := json.Unmarshal(data, config)
		if err != nil {
			return err
		}

		app.Lock()
		app.config = config
		app.Unlock()

		logrus.Warn("Received new config")
		select {
		case app.render <- struct{}{}:
		default:
		}

		return nil
	}
}

func onDevices(data json.RawMessage) error {
	devs := devices.NewList()
	err := json.Unmarshal(data, devs)
	if err != nil {
		return err
	}

	triggerRender := false

	for _, device := range devs.All() {
		//check if state is different
		//logrus.Info("state", device.State)
		state := make(devices.State)
		if prevDev := previous.Get(device.ID); prevDev != nil {
			//logrus.Info("prevState", prevDev.State)
			state = prevDev.State.Diff(device.State)
		} else {
			state = device.State
		}

		previous.Add(device)

		if len(state) > 0 {
			// Check if we should re-render
			for index, _ := range app.decks {
				_, pageConfig, ok := getPageConfig(index)

				if !ok {
					continue
				}

				for _, keyConfig := range pageConfig.Keys {
					if keyConfig.Node == device.ID.Node && keyConfig.Device == device.ID.ID {
						triggerRender = true
					}
				}
			}
		}
	}

	if triggerRender {
		select {
		case app.render <- struct{}{}:
		default:
		}
	}

	return nil
}

func (app *appState) worker() {
	app.decks = streamdeck.FindDecks()

	for index, deck := range app.decks {
		err := deck.Open()
		if err != nil {
			logrus.Error(err)
		}

		deck.Reset()
		deck.SetBrightness(30)

		deck.OnKeyDown(func(key int, s bool) {
			fmt.Printf("Key %d pressed\n", key)

			_, pageConfig, ok := getPageConfig(index)
			if !ok {
				return
			}

			keyConfig := pageConfig.Keys[key]

			// Find the linked device
			d := previous.Get(devices.ID{Node: keyConfig.Node, ID: keyConfig.Device})
			if d != nil {
				current, ok := d.State[keyConfig.Field]
				if ok {
					targetState := true
					if current != nil {
						targetState = !current.(bool)
					}

					devs := devices.NewList()
					newState := make(devices.State)
					newState[keyConfig.Field] = targetState
					devs.Add(d.Copy())
					devs.SetState(devices.ID{Node: keyConfig.Node, ID: keyConfig.Device}, newState)
					app.node.WriteMessage("state-change", devs)
				}
			}

			spew.Dump(keyConfig)
		})
	}

	for {
		for index, deck := range app.decks {
			page, pageConfig, ok := getPageConfig(index)

			if !ok {
				continue
			}

			logrus.Infof("Draw page %s to deck %d", page, index)
			for key, keyConfig := range pageConfig.Keys {
				// Try to find out if we have a state
				d := previous.Get(devices.ID{Node: keyConfig.Node, ID: keyConfig.Device})
				if d != nil {
					drawStateToKey(deck, keyConfig.Name, d.State[keyConfig.Field], key)
					continue
				}
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

		<-app.render
	}
}

func getPageConfig(index int) (string, *page, bool) {
	app.Lock()
	page := "default"
	if index < len(app.page) {
		page = app.page[index]
	}

	if app.config == nil {
		return "", nil, false
	}

	pageConfig, ok := app.config.Pages[page]
	app.Unlock()
	if !ok {
		return "", nil, false
	}

	return page, &pageConfig, true
}
