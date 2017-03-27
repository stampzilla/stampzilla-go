package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/pkg/hueemulator"
	"github.com/stampzilla/stampzilla-go/protocol"
	"github.com/stampzilla/stampzilla-go/protocol/devices"
)

var VERSION string = "dev"
var BUILD_DATE string = ""
var listenPort string
var ip string
var debug bool
var standalone bool

func init() {
	flag.StringVar(&listenPort, "listenport", "80", "Port to listen to. Must be 80 for Google Home to work")
	flag.StringVar(&ip, "ip", hueemulator.GetPrimaryIp(), "Ip to listen to.")
	flag.BoolVar(&debug, "debug", false, "Debug. Without this we dont print other than errors. Optimized not to wear on raspberry pi SD card.")
	flag.BoolVar(&standalone, "standalone", false, "Run standalone without communicating with stampzilla-server Host:Port configured in config.json.")
}

func main() {
	flag.Parse()
	hueemulator.SetLogger(os.Stdout)
	hueemulator.SetDebug(debug)

	config := basenode.NewConfig()

	basenode.SetConfig(config)

	node := protocol.NewNode("huebridge")
	node.Version = VERSION
	node.BuildDate = BUILD_DATE

	//devices := NewDevices()
	nodespecific := &nodeSpecific{}
	err := config.NodeSpecific(&nodespecific)
	if err != nil {
		log.Println(err)
	}

	if nodespecific.ListenPort != "" && listenPort == "80" {
		listenPort = nodespecific.ListenPort
	}

	if !standalone {
		log.Println("Syncing devices from server")
		syncDevicesFromServer(config, nodespecific)

		go func() {
			for range time.NewTicker(60 * time.Second).C {
				if debug {
					log.Println("Syncing devices from server")
				}
				if syncDevicesFromServer(config, nodespecific) {
					setupHueHandlers(nodespecific)
				}
			}
		}()
		connection := basenode.Connect()
		go monitorState(node, connection)
	}

	//spew.Dump(config)

	setupHueHandlers(nodespecific)
	// it is very important to use a full IP here or the UPNP does not work correctly.
	ip := hueemulator.GetPrimaryIp()
	panic(hueemulator.ListenAndServe(ip + ":" + listenPort))
}

func setupHueHandlers(ns *nodeSpecific) {
	for _, d := range ns.Devices {
		log.Println("Setting up handler for device: ", d.Name)
		dev := d
		hueemulator.Handle(d.Id, d.Name, func(req hueemulator.Request) error {
			fmt.Println("im handling from", req.RemoteAddr, req.Request.On)
			if req.Request.Brightness != 0 {
				if dev.Url.Level == "" {
					log.Println("No level url set. ignoring...")
					return nil
				}
				// hue protocol is 0-255. Stampzilla level is 0-100
				bri := float64(req.Request.Brightness) * 100.0 / 255.0
				url := fmt.Sprintf(dev.Url.Level, bri)
				log.Println("Request url: ", url)
				_, err := http.Get(url)
				if err != nil {
					log.Println(err)
				}

				return nil
			}
			if req.Request.On {
				log.Println("Request url: ", dev.Url.On)
				_, err := http.Get(dev.Url.On)
				if err != nil {
					log.Println(err)
				}

			} else {
				log.Println("Request url: ", dev.Url.Off)
				_, err := http.Get(dev.Url.Off)
				if err != nil {
					log.Println(err)
				}

			}
			return nil
		})

	}
}

func syncDevicesFromServer(config *basenode.Config, ns *nodeSpecific) bool {
	didChange := false

	serverDevs, err := fetchDevices(config, ns)
	if err != nil {
		log.Println(err)
		return false
	}

outer:
	for uuid, sdev := range serverDevs.All() {
		for _, v := range ns.Devices {
			if v.UUID == uuid {
				if debug {
					log.Printf("Already have device: %s. Do not add again.\n", sdev.Name)
				}
				continue outer
			}
		}

		//Skip non controllable devices
		if sdev.Type != "lamp" && sdev.Type != "dimmableLamp" {
			continue
		}

		//We dont have the device so we add it
		baseURL := fmt.Sprintf("http://%s:%s/api/nodes/", config.Host, ns.Port)
		dev := &Device{
			Name: sdev.Name,
			Id:   ns.Devices.maxId() + 1,
			Url: &URL{
				Level: baseURL + sdev.Node + "/cmd/level/" + sdev.Id + "/%f",
				On:    baseURL + sdev.Node + "/cmd/on/" + sdev.Id,
				Off:   baseURL + sdev.Node + "/cmd/off/" + sdev.Id,
			},
			UUID: uuid,
		}

		didChange = true
		ns.Devices = append(ns.Devices, dev)
	}

	//Dont save file if no new devices are found
	if !didChange {
		return false
	}

	data, err := json.Marshal(ns)
	if err != nil {
		log.Println(err)
		return false
	}
	raw := json.RawMessage(data)
	config.Node = &raw
	basenode.SaveConfigToFile(config)
	return true
}

func fetchDevices(config *basenode.Config, ns *nodeSpecific) (*devices.Map, error) {
	//TODO use nodespecific config
	url := fmt.Sprintf("http://%s:%s/api/devices", config.Host, ns.Port)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	devmap := devices.NewMap()
	err = json.NewDecoder(resp.Body).Decode(&devmap)
	if err != nil {
		return nil, err
	}
	return devmap, nil

}

// WORKER that monitors the current connection state
func monitorState(node *protocol.Node, connection basenode.Connection) {
	for s := range connection.State() {
		switch s {
		case basenode.ConnectionStateConnected:
			connection.Send(node.Node())
		case basenode.ConnectionStateDisconnected:
		}
	}
}
