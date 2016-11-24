package hueemulator

import (
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

type light struct {
	State struct {
		On        bool      `json:"on"`
		Bri       int       `json:"bri"`
		Hue       int       `json:"hue"`
		Sat       int       `json:"sat"`
		Effect    string    `json:"effect"`
		Ct        int       `json:"ct"`
		Alert     string    `json:"alert"`
		Colormode string    `json:"colormode"`
		Reachable bool      `json:"reachable"`
		XY        []float64 `json:"xy"`
	} `json:"state"`
	Type             string `json:"type"`
	Name             string `json:"name"`
	ModelId          string `json:"modelid"`
	ManufacturerName string `json:"manufacturername"`
	UniqueId         string `json:"uniqueid"`
	SwVersion        string `json:"swversion"`
	PointSymbol      struct {
		One   string `json:"1"`
		Two   string `json:"2"`
		Three string `json:"3"`
		Four  string `json:"4"`
		Five  string `json:"5"`
		Six   string `json:"6"`
		Seven string `json:"7"`
		Eight string `json:"8"`
	} `json:"pointsymbol"`
}

type lights struct {
	Lights map[string]light `json:"lights"`
}

func initLight(name string) light {
	l := light{
		Type:             "Extended color light",
		ModelId:          "LCT001",
		SwVersion:        "66009461",
		ManufacturerName: "Philips",
		Name:             name,
		UniqueId:         name,
	}
	l.State.Reachable = true
	l.State.XY = []float64{0.4255, 0.3998} // this seems to be voodoo, if it is nil the echo says it could not turn on/off the device, useful...
	return l
}

func enumerateLights() map[string]light {
	//lightList := lights{}
	lightList := make(map[string]light)
	for name, hstate := range handlerMap {
		l := initLight(name)
		l.State.On = hstate.OnState
		lightList[l.UniqueId] = l
	}
	return lightList
}

func getLightsList(c *gin.Context) {
	c.JSON(200, enumerateLights())
}

func setLightState(c *gin.Context) {
	//defer r.Body.Close()
	//w.Header().Set("Content-Type", "application/json")
	req := make(map[string]bool)
	//json.NewDecoder(r.Body).Decode(&req)
	err := c.BindJSON(req)
	if err != nil {
		log.Println(err)
		return
	}

	lightId := c.Param("lightId")

	log.Println("[DEVICE]", c.Param("userId"), "requested state:", req["on"])
	response := Response{}
	if hstate, ok := handlerMap[lightId]; ok {
		hstate.Handler(Request{
			UserId:           c.Param("userId"),
			RequestedOnState: req["on"],
			RemoteAddr:       c.Request.RemoteAddr,
		}, &response)
		log.Println("[DEVICE] handler replied with state:", response.OnState)
		hstate.OnState = response.OnState
		handlerMap[lightId] = hstate
	}
	if !response.ErrorState {
		c.Writer.Write([]byte("[{\"success\":{\"/lights/" + lightId + "/state/on\":" + strconv.FormatBool(response.OnState) + "}}]"))
	}
}

func getLightInfo(c *gin.Context) {

	l := initLight(c.Param("lightId"))

	if hstate, ok := handlerMap[c.Param("lightId")]; ok {
		if hstate.OnState {
			l.State.On = true
		}
	}
	c.JSON(200, l)
}
