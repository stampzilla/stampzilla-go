package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/RangelReale/osin"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-google-assistant/googleassistant"
)

func smartHomeActionHandler(oauth2server *osin.Server) func(c *gin.Context) {
	return func(c *gin.Context) {
		auth := osin.CheckBearerAuth(c.Request)
		if auth == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		accessToken, err := oauth2server.Storage.LoadAccess(auth.Code)
		if err != nil || accessToken == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if accessToken.IsExpired() {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		dec := json.NewDecoder(c.Request.Body)
		defer c.Request.Body.Close()

		//body, err := ioutil.ReadAll(c.Request.Body)
		//if err != nil {
		//logrus.Error(err)
		//return
		//}
		//logrus.Info(string(body))

		r := &googleassistant.Request{}

		err = dec.Decode(r)
		if err != nil {
			logrus.Error(err)
			return
		}

		logrus.Info("Intent: ", r.Inputs.Intent())
		switch r.Inputs.Intent() {
		case googleassistant.SyncIntent:
			c.JSON(http.StatusOK, syncHandler(r))
		case googleassistant.ExecuteIntent:
			c.JSON(http.StatusOK, executeHandler(r))

		}
		if r.Inputs.Intent() == googleassistant.SyncIntent {
		}

	}
}

func executeHandler(req *googleassistant.Request) *googleassistant.Response {
	resp := &googleassistant.Response{}
	resp.RequestID = req.RequestID

	levelCommands := make(map[int]googleassistant.ResponseCommand)

	onCommand := googleassistant.NewResponseCommand()
	onCommand.States.On = true
	onCommand.States.Online = true
	offCommand := googleassistant.NewResponseCommand()
	onCommand.States.Online = true

	for _, command := range req.Inputs.Payload().Commands {

		for _, v := range command.Execution {
			for _, googleDev := range command.Devices {
				dev := nodespecific.Device(googleDev.ID)
				if dev == nil {
					continue
				}
				if v.Command == googleassistant.CommandOnOff {
					if v.Params.On {
						log.Println("Calling ", dev.Url.On)
						http.Get(dev.Url.On)
						onCommand.IDs = append(onCommand.IDs, googleDev.ID)
					} else {
						offCommand.IDs = append(onCommand.IDs, googleDev.ID)
						log.Println("Calling ", dev.Url.Off)
						http.Get(dev.Url.Off)
					}
				}
				if v.Command == googleassistant.CommandBrightnessAbsolute {
					bri := v.Params.Brightness
					u := fmt.Sprintf(dev.Url.Level, bri)
					log.Println("Calling ", u)
					http.Get(u)

					if _, ok := levelCommands[v.Params.Brightness]; !ok {
						levelCommands[bri] = googleassistant.ResponseCommand{
							States: googleassistant.ResponseStates{
								Brightness: bri,
							},
							Status: "SUCCESS",
						}
					}
					lvlCmd := levelCommands[bri]
					lvlCmd.IDs = append(lvlCmd.IDs, googleDev.ID)
					levelCommands[bri] = lvlCmd
				}
			}
		}

	}

	for _, v := range levelCommands {
		resp.Payload.Commands = append(resp.Payload.Commands, v)
	}
	if onCommand.IDs != nil {
		resp.Payload.Commands = append(resp.Payload.Commands, onCommand)
	}
	if offCommand.IDs != nil {
		resp.Payload.Commands = append(resp.Payload.Commands, offCommand)
	}

	return resp
}

func syncHandler(req *googleassistant.Request) *googleassistant.Response {

	resp := &googleassistant.Response{}
	resp.RequestID = req.RequestID
	resp.Payload.AgentUserID = "agentuserid"

	for _, dev := range nodespecific.Devices() {

		rdev := googleassistant.Device{
			ID:   dev.ID,
			Type: "action.devices.types.LIGHT",
			Name: googleassistant.DeviceName{
				Name: dev.Name,
			},
			WillReportState: false,
			Traits: []string{
				"action.devices.traits.OnOff",
				"action.devices.traits.Brightness",
				//"action.devices.traits.ColorTemperature",
				//"action.devices.traits.ColorSpectrum",
			},
			//Attributes: googleassistant.DeviceAttributes{
			//ColorModel:      "RGB",
			//TemperatureMinK: 2000,
			//TemperatureMaxK: 6500,
			//},
		}
		resp.Payload.Devices = append(resp.Payload.Devices, rdev)

	}

	return resp
}
