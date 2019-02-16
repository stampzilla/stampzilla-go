package main

import (
	"encoding/json"
	"net/http"

	"github.com/RangelReale/osin"
	"github.com/davecgh/go-spew/spew"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-google-assistant/googleassistant"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/pkg/node"
)

// SmartHomeHandler contains the logic to answer Google Actions API requests and authorize them usnig oauth2.
type SmartHomeHandler struct {
	node       *node.Node
	deviceList *devices.List
}

// NewSmartHomeHandler returns a new instance of SmartHomeHandler.
func NewSmartHomeHandler(node *node.Node, deviceList *devices.List) *SmartHomeHandler {
	return &SmartHomeHandler{
		node:       node,
		deviceList: deviceList,
	}

}

func (shh *SmartHomeHandler) smartHomeActionHandler(oauth2server *osin.Server) func(c *gin.Context) {
	return func(c *gin.Context) {
		auth := osin.CheckBearerAuth(c.Request)
		if auth == nil {
			logrus.Error("CheckBearerAuth error")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		accessToken, err := oauth2server.Storage.LoadAccess(auth.Code)
		if err != nil || accessToken == nil {
			logrus.Errorf("LoadAccess error: %#v", err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if accessToken.IsExpired() {
			logrus.Errorf("Accesstoken %s expired at: %s", accessToken.AccessToken, accessToken.ExpireAt())
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
		logrus.Debug("Request:", spew.Sdump(r))
		switch r.Inputs.Intent() {
		case googleassistant.SyncIntent:
			c.JSON(http.StatusOK, shh.syncHandler(r))
		case googleassistant.ExecuteIntent:
			c.JSON(http.StatusOK, shh.executeHandler(r))

		}
	}
}

func (shh *SmartHomeHandler) executeHandler(req *googleassistant.Request) *googleassistant.Response {
	resp := &googleassistant.Response{}
	resp.RequestID = req.RequestID

	levelCommands := make(map[int]googleassistant.ResponseCommand)

	onCommand := googleassistant.NewResponseCommand()
	onCommand.States.On = true
	onCommand.States.Online = true
	offCommand := googleassistant.NewResponseCommand()
	offCommand.States.Online = true

	deviceNotFound := googleassistant.NewResponseCommand()
	deviceNotFound.Status = "ERROR"
	deviceNotFound.ErrorCode = "deviceNotFound"

	deviceOffline := googleassistant.NewResponseCommand()
	deviceOffline.Status = "OFFLINE"

	affectedDevs := make(devices.DeviceMap)

	for _, command := range req.Inputs.Payload().Commands {

		for _, v := range command.Execution {
			for _, googleDev := range command.Devices {
				devID, err := devices.NewIDFromString(googleDev.ID)
				if err != nil {
					logrus.Error(err)
					continue
				}

				dev := shh.deviceList.Get(devID)
				if dev == nil || dev.Type != "light" {
					deviceNotFound.IDs = append(deviceNotFound.IDs, googleDev.ID)
					continue
				}

				if !dev.Online {
					deviceOffline.IDs = append(deviceOffline.IDs, googleDev.ID)
					continue
				}
				if v.Command == googleassistant.CommandOnOff {
					if v.Params.On {
						logrus.Infof("Turning device %s (%s) on ", dev.Name, dev.ID)
						dev.State["on"] = true
						onCommand.IDs = append(onCommand.IDs, googleDev.ID)
					} else {
						offCommand.IDs = append(onCommand.IDs, googleDev.ID)
						logrus.Infof("Turning device %s (%s) off", dev.Name, dev.ID)
						dev.State["on"] = false
					}
					affectedDevs[devID] = dev
				}
				if v.Command == googleassistant.CommandBrightnessAbsolute {
					bri := v.Params.Brightness
					logrus.Infof("Dimming device %s (%s) to %d", dev.Name, dev.ID, bri)
					dev.State["brightness"] = float64(bri) / 100.0
					affectedDevs[devID] = dev
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

	shh.node.WriteMessage("state-change", affectedDevs)

	for _, v := range levelCommands {
		resp.Payload.Commands = append(resp.Payload.Commands, v)
	}

	for _, v := range []googleassistant.ResponseCommand{onCommand, offCommand, deviceNotFound, deviceOffline} {
		if v.IDs != nil {
			resp.Payload.Commands = append(resp.Payload.Commands, v)
		}
	}

	if logrus.GetLevel() >= logrus.DebugLevel {
		jResp, err := json.Marshal(resp)
		logrus.Debugf("Execute Error: %s Response: %s", err, string(jResp))
	}
	return resp
}

func (shh *SmartHomeHandler) syncHandler(req *googleassistant.Request) *googleassistant.Response {

	resp := &googleassistant.Response{}
	resp.RequestID = req.RequestID
	resp.Payload.AgentUserID = "agentuserid"

	for _, dev := range shh.deviceList.All() {

		rdev := googleassistant.Device{
			ID:   dev.ID.String(),
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

	logrus.Debug("Sync Response: ", resp)
	return resp
}
