package hueemulator

import (
	"io/ioutil"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

type light struct {
	State struct {
		On        bool      `json:"on"`
		Bri       uint8     `json:"bri"`
		Hue       int       `json:"hue,omitempty"`
		Sat       int       `json:"sat,omitempty"`
		Effect    string    `json:"effect,omitempty"`
		Ct        int       `json:"ct,omitempty"`
		Alert     string    `json:"alert,omitempty"`
		Colormode string    `json:"colormode,omitempty"`
		Reachable bool      `json:"reachable"`
		XY        []float64 `json:"xy,omitempty"`
	} `json:"state"`
	Type             string `json:"type"`
	Name             string `json:"name"`
	ModelId          string `json:"modelid"`
	ManufacturerName string `json:"manufacturername"`
	UniqueId         string `json:"uniqueid"`
	SwVersion        string `json:"swversion"`
}

type lights struct {
	Lights map[string]light `json:"lights"`
}

func initLight(id int, name string) *light {
	l := &light{
		Type:             "Dimmable light",
		ModelId:          "LWB014",
		SwVersion:        "1.15.0_r18729",
		ManufacturerName: "Philips",
		Name:             name,
		UniqueId:         strconv.Itoa(id),
	}
	l.State.Reachable = true
	return l
}

func enumerateLights() map[string]*light {
	lightList := make(map[string]*light)
	handlerMapLock.Lock()
	for _, hstate := range handlerMap {
		lightList[strconv.Itoa(hstate.Id)] = hstate.Light
	}
	handlerMapLock.Unlock()
	return lightList
}

func getLightsList(c *gin.Context) {
	c.JSON(200, enumerateLights())
}

type request struct {
	On         bool   `json:"on"`
	Brightness uint8  `json:"bri"`
	Hue        uint16 `json:"hue"`
	Sat        uint8  `json:"sat"`
}

func setLightState(c *gin.Context) {
	req := &request{}
	err := c.BindJSON(req)
	if err != nil {
		log.Println(err)
		defer c.Request.Body.Close()
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Println(err)
		}
		log.Println("Request body was:", string(body))
		return
	}

	lightId := c.Param("lightId")

	log.Println("[DEVICE]", c.Param("userId"), "requested state:", req.On, "requested brightness: ", req.Brightness)

	// if hstate, ok := handlerMap[lightId]; ok {
	if hstate := getHueStateById(lightId); hstate != nil {
		hstate.Handler(Request{
			UserId: c.Param("userId"),
			// RequestedOnState: req.On,
			Request:    req,
			RemoteAddr: c.Request.RemoteAddr,
		})
		hstate.Light.State.On = req.On
		hstate.Light.State.Bri = req.Brightness
		log.Println("[DEVICE] handler replied with state:", req.On)
		// handlerMap[lightId] = hstate
		c.Writer.Write([]byte("[{\"success\":{\"/lights/" + lightId + "/state/on\":" + strconv.FormatBool(req.On) + "}}]"))
	}
}

func getHueStateById(id string) *huestate {
	for _, h := range handlerMap {
		if strconv.Itoa(h.Id) == id {
			return h
		}
	}
	return nil
}

func getLightInfo(c *gin.Context) {
	id := c.Param("lightId")
	hs := getHueStateById(id)
	c.JSON(200, hs.Light)
}
