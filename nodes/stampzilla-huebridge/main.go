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

func init() {
	flag.StringVar(&listenPort, "listenport", "80", "Port to listen to. Must be 80 for Google Home to work")
	flag.StringVar(&ip, "ip", hueemulator.GetPrimaryIp(), "Ip to listen to.")
	flag.BoolVar(&debug, "debug", false, "Debug. Without this we dont print other than errors. Optimized not to wear on raspberry pi SD card.")
	flag.Parse()
}

type NodeSpecific struct {
	Port       string
	ListenPort string
	Devices    []*Device
}

func main() {
	hueemulator.SetLogger(os.Stdout)
	hueemulator.SetDebug(debug)

	config := basenode.NewConfig()

	basenode.SetConfig(config)

	node := protocol.NewNode("huebridge")
	node.Version = VERSION
	node.BuildDate = BUILD_DATE

	//devices := NewDevices()
	nodespecific := &NodeSpecific{}
	err := config.NodeSpecific(&nodespecific)
	if err != nil {
		log.Println(err)
		return
	}

	if nodespecific.ListenPort != "" {
		listenPort = nodespecific.ListenPort
	}

	//TODO this works. But we need to save devices to json before uncommenting and enableing it!
	log.Println("Syncing devices from server")
	SyncDevicesFromServer(config, nodespecific)

	go func() {
		for range time.NewTicker(60 * time.Second).C {
			if debug {
				log.Println("Syncing devices from server")
			}
			SyncDevicesFromServer(config, nodespecific)
		}
	}()

	//spew.Dump(config)

	for _, d := range nodespecific.Devices {
		log.Println(d)
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

	connection := basenode.Connect()

	go monitorState(node, connection)

	// it is very important to use a full IP here or the UPNP does not work correctly.
	ip := hueemulator.GetPrimaryIp()
	panic(hueemulator.ListenAndServe(ip + ":" + listenPort))
	//panic(hueemulator.ListenAndServe("192.168.13.86:8080"))
}

//func NewDevices() *Devices {
//return &Devices{
//Devices: make([]*Device, 0),
//}
//}

type Url struct {
	Level string
	On    string
	Off   string
}

type Device struct {
	Name string
	Id   int
	UUID string
	Url  *Url
}

func SyncDevicesFromServer(config *basenode.Config, ns *NodeSpecific) {

	serverDevs, err := fetchDevices(config, ns)
	if err != nil {
		log.Println(err)
		return
	}

outer:
	for uuid, sdev := range serverDevs {
		for _, v := range ns.Devices {
			if v.UUID == uuid {
				log.Printf("Already have device: %s. Do not add again.\n", sdev.Name)
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
			Id:   len(ns.Devices) + 1,
			Url: &Url{
				Level: baseURL + sdev.Node + "/cmd/level/" + sdev.Id + "/%f",
				On:    baseURL + sdev.Node + "/cmd/on/" + sdev.Id,
				Off:   baseURL + sdev.Node + "/cmd/off/" + sdev.Id,
			},
			UUID: uuid,
		}

		ns.Devices = append(ns.Devices, dev)
	}

	data, err := json.Marshal(ns)
	if err != nil {
		log.Println(err)
		return
	}
	raw := json.RawMessage(data)
	config.Node = &raw
	basenode.SaveConfigToFile(config)
}

func fetchDevices(config *basenode.Config, ns *NodeSpecific) (devices.Map, error) {
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
