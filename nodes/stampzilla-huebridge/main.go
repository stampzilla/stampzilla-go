package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/stampzilla/stampzilla-go/pkg/hueemulator"
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

	//TODO fetch config from stampzilla server when devices feature is merged.

	for _, d := range devices.Devices {
		log.Println(d)
		dev := d
		hueemulator.Handle(d.Id, d.Name, func(req hueemulator.Request) error {
			fmt.Println("im handling test from", req.RemoteAddr, req.Request.On)
			if req.Request.Brightness != 0 {
				log.Println("Request url: ", dev.Url.Dim)
				_, err := http.Get(dev.Url.Dim)
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
	Dim string
	On  string
	Off string
}

type Device struct {
	Name string
	Id   int
	Url  Url
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
