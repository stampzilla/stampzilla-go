package main

import (
	"encoding/json"
	"math"

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
				if dev == nil {
					deviceNotFound.IDs = append(deviceNotFound.IDs, googleDev.ID)
					continue
				}

				if !dev.Online {
					deviceOffline.IDs = append(deviceOffline.IDs, googleDev.ID)
					continue
				}
				switch v.Command {
				case googleassistant.CommandOnOff:
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

				case googleassistant.CommandBrightnessAbsolute:
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
				case googleassistant.CommandColorAbsolute:
					colortemp := v.Params.Color.Temperature
					logrus.Infof("Setting device colortemperature %s (%s) to %d", dev.Name, dev.ID, colortemp)
					dev.State["temperature"] = float64(colortemp)
					affectedDevs[devID] = dev
					if _, ok := levelCommands[colortemp]; !ok {
						levelCommands[colortemp] = googleassistant.ResponseCommand{
							States: googleassistant.ResponseStates{
								//Color: colortemp,
							},
							Status: "SUCCESS",
						}
					}
					lvlCmd := levelCommands[colortemp]
					lvlCmd.IDs = append(lvlCmd.IDs, googleDev.ID)
					levelCommands[colortemp] = lvlCmd

				default:
					logrus.Warnf("Unkown command '%s'", v.Command)
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

func (shh *SmartHomeHandler) syncHandler(nodeID string, req *googleassistant.Request) *googleassistant.Response {

	resp := &googleassistant.Response{}
	resp.RequestID = req.RequestID
	resp.Payload.AgentUserID = nodeID

	for _, dev := range shh.deviceList.All() {
		traits := []string{}
		attributes := googleassistant.DeviceAttributes{}
		for _, v := range dev.Traits {
			switch v {
			case "OnOff":
				traits = append(traits, "action.devices.traits.OnOff")
			case "Brightness":
				traits = append(traits, "action.devices.traits.Brightness")
			case "ColorSetting":
				traits = append(traits, "action.devices.traits.ColorTemperature")
				attributes.TemperatureMinK = 2000
				attributes.TemperatureMaxK = 6500
			}
		}

		if len(traits) == 0 {
			continue
		}

		rdev := googleassistant.Device{
			ID:   dev.ID.String(),
			Type: "action.devices.types.LIGHT",
			Name: googleassistant.DeviceName{
				Name: dev.Name,
			},
			WillReportState: true,
			Traits:          traits,
			Attributes:      attributes,
		}
		if dev.Alias != "" {
			rdev.Name.Name = dev.Alias
		}
		resp.Payload.Devices = append(resp.Payload.Devices, rdev)

	}

	logrus.Debug("Sync Response: ", resp)
	return resp
}

func (shh *SmartHomeHandler) queryHandler(req *googleassistant.Request) *googleassistant.QueryResponse {
	resp := &googleassistant.QueryResponse{}
	resp.RequestID = req.RequestID
	resp.Payload.Devices = make(map[string]map[string]interface{})

	for _, v := range req.Inputs.Payload().Devices {
		devID, err := devices.NewIDFromString(v.ID)
		if err != nil {
			logrus.Error(err)
			continue
		}
		dev := shh.deviceList.Get(devID)
		if dev == nil {
			continue
		}
		resp.Payload.Devices[devID.String()] = map[string]interface{}{
			"on":     dev.State["on"],
			"online": dev.Online,
		}
		dev.State.Float("brightness", func(bri float64) {
			resp.Payload.Devices[devID.String()]["brightness"] = int(math.Round(bri * 100.0))
		})
	}
	return resp
}
