package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/stampzilla/stampzilla-go/pkg/hueemulator"
	"github.com/stampzilla/stampzilla-go/protocol/devices"
)

var port string
var ip string
var config string
var debug bool

func init() {
	flag.StringVar(&port, "port", "80", "Port to listen to. Must be 80 for Google Home to work")
	flag.StringVar(&ip, "ip", hueemulator.GetPrimaryIp(), "Ip to listen to.")
	flag.StringVar(&config, "config", "config.json", "Path to config file.")
	flag.BoolVar(&debug, "debug", false, "Debug. Without this we dont print other than errors. Optimized not to wear on raspberry pi SD card.")
	flag.Parse()
}

func main() {
	hueemulator.SetLogger(os.Stdout)
	hueemulator.SetDebug(debug)

	devices := NewDevices()
	err := devices.ReadFromFile(config)
	if err != nil {
		log.Println(err)
		return
	}

	//TODO this works. But we need to save devices to json before uncommenting and enableing it!
	//serverDevices, err := fetchDevices()
	//if err != nil {
	//log.Println(err)
	//} else {
	//SyncDevicesFromServer(devices, serverDevices)
	//}

	//spew.Dump(devices)

	for _, d := range devices.Devices {
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

	// it is very important to use a full IP here or the UPNP does not work correctly.
	ip := hueemulator.GetPrimaryIp()
	panic(hueemulator.ListenAndServe(ip + ":80"))
	//panic(hueemulator.ListenAndServe("192.168.13.86:8080"))
}

func NewDevices() *Devices {
	return &Devices{
		Devices: make([]*Device, 0),
	}
}

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

type Devices struct {
	Devices []*Device
}

func (c *Devices) ReadFromFile(filepath string) error {
	configFile, err := os.Open(filepath)
	if err != nil {
		return err
	}

	devices := make([]*Device, 0)
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&devices); err != nil {
		return err
	}

	c.Devices = devices
	return nil
}

func SyncDevicesFromServer(localDevs *Devices, serverDevs devices.Map) {

outer:
	for uuid, sdev := range serverDevs {
		for _, v := range localDevs.Devices {
			if v.UUID == uuid {
				log.Printf("Already have device: %s. Do not add again.\n", sdev.Name)
				continue outer
			}
		}

		//Skip non controllable devices
		if sdev.Type != "lamp" {
			continue
		}

		//We dont have the device so we add it
		dev := &Device{
			Name: sdev.Name,
			Id:   len(localDevs.Devices) + 1,
			Url: &Url{
				Level: "http://bulan.lan:8080/api/nodes/" + sdev.Node + "/cmd/level/" + sdev.Id + "/%f",
				On:    "http://bulan.lan:8080/api/nodes/" + sdev.Node + "/cmd/on/" + sdev.Id,
				Off:   "http://bulan.lan:8080/api/nodes/" + sdev.Node + "/cmd/off/" + sdev.Id,
			},
			UUID: uuid,
		}

		localDevs.Devices = append(localDevs.Devices, dev)
	}

}

func fetchDevices() (devices.Map, error) {
	//TODO use nodespecific config
	resp, err := http.Get("http://192.168.13.1:8080/api/devices")
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
