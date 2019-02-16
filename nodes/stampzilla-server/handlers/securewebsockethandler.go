package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/olahol/melody"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/ca"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/interfaces"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/logic"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/store"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/websocket"
)

type secureWebsocketHandler struct {
	CA              *ca.CA
	Store           *store.Store
	Config          *models.Config
	WebsocketSender websocket.Sender
}

// NewSecureWebsockerHandler is the constructor
func NewSecureWebsockerHandler(store *store.Store, config *models.Config, ws websocket.Sender, ca *ca.CA) WebsocketHandler {
	return &secureWebsocketHandler{
		CA:              ca,
		Store:           store,
		Config:          config,
		WebsocketSender: ws,
	}
}

func BroadcastUpdate(sender websocket.Sender) func(string, *store.Store) error {
	send := func(area string, data interface{}) error {
		return sender.BroadcastWithFilter(area, data, func(s *melody.Session) bool {
			v, exists := s.Get("subscriptions")
			if !exists {
				return false
			}
			if v, ok := v.([]string); ok {
				for _, topic := range v {
					if topic == area {
						return true
					}
				}
			}
			return false
		})
	}

	return func(area string, store *store.Store) error {
		switch area {
		case "devices":
			return send(area, store.GetDevices())
		case "connections":
			return send(area, store.GetConnections())
		case "nodes":
			return send(area, store.GetNodes())
		case "certificates":
			return send(area, store.GetCertificates())
		case "requests":
			return send(area, store.GetRequests())
		case "rules":
			return send(area, store.GetRules())
		case "savedstates":
			return send(area, store.GetSavedStates())
		case "schedules":
			return send(area, store.GetScheduledTasks())
		case "server":
			return send(area, store.GetServerStateAsJson())
		}
		return nil
	}
}

func sliceHas(s []string, val string) bool {
	for _, v := range s {
		if v == val {
			return true
		}

	}
	return false
}

func (wsh *secureWebsocketHandler) Message(s interfaces.MelodySession, msg *models.Message) error {
	switch msg.Type {
	case "accept-request":
		connection := ""
		err := json.Unmarshal(msg.Body, &connection)
		if err != nil {
			return err
		}

		wsh.Store.AcceptRequest(connection)
	case "subscribe":
		subscribeTo := []string{}
		err := json.Unmarshal(msg.Body, &subscribeTo)
		if err != nil {
			return err
		}

		v, exists := s.Get("subscriptions")
		if !exists {
			s.Set("subscriptions", subscribeTo)
		} else {
			subs := []string{}
			if v, ok := v.([]string); ok {
				for _, sub := range subscribeTo {
					if !sliceHas(v, sub) {
						subs = append(subs, sub)
					}
				}
				s.Set("subscriptions", append(v, subs...))
			}
		}

		if v, exists := s.Get("subscriptions"); exists {
			logrus.Debug("Active subscriptions: ", v)
		}

		fn := BroadcastUpdate(wsh.WebsocketSender)
		for _, v := range subscribeTo {
			fn(v, wsh.Store)
		}

		wsh.Store.ConnectionChanged()

	case "update-device":
		device := devices.NewDevice()
		err := json.Unmarshal(msg.Body, device)
		if err != nil {
			return err
		}

		logrus.WithFields(logrus.Fields{
			"from":   msg.FromUUID,
			"device": device,
		}).Debug("Received device")

		if device != nil {
			wsh.Store.AddOrUpdateDevice(device)
		}
	case "update-devices":
		devs := make(devices.DeviceMap)
		err := json.Unmarshal(msg.Body, &devs)
		if err != nil {
			return err
		}

		logrus.WithFields(logrus.Fields{
			"from":    msg.FromUUID,
			"devices": devs,
		}).Debug("Received devices")

		for _, dev := range devs {
			if dev.State == nil {
				dev.State = make(devices.State)
			}
			wsh.Store.AddOrUpdateDevice(dev)
		}
	case "setup-node":
		node := &models.Node{}
		err := json.Unmarshal(msg.Body, node)
		if err != nil {
			return err
		}

		logrus.WithFields(logrus.Fields{
			"from":   msg.FromUUID,
			"config": node,
		}).Debug("Received new node configuration")

		wsh.Store.AddOrUpdateNode(node)
		err = wsh.Store.SaveNodes()
		if err != nil {
			return err
		}
		wsh.WebsocketSender.SendToID(node.UUID, "setup", node)
	case "state-change":
		devs := devices.NewList()
		err := json.Unmarshal(msg.Body, devs)
		if err != nil {
			return err
		}

		logrus.WithFields(logrus.Fields{
			"from":    msg.FromUUID,
			"devices": devs,
		}).Debug("Received state change request")

		for node, devices := range devs.StateGroupedByNode() {
			logrus.WithFields(logrus.Fields{
				"to": node,
			}).Debug("Send state change request to node")
			wsh.WebsocketSender.SendToID(node, "state-change", devices)
		}
	case "update-rules":
		rules := logic.Rules{}
		err := json.Unmarshal(msg.Body, &rules)
		if err != nil {
			return err
		}

		logrus.WithFields(logrus.Fields{
			"from":  msg.FromUUID,
			"rules": rules,
		}).Debug("Received new rules")

		wsh.Store.AddOrUpdateRules(rules)
	case "update-schedules":
		tasks := logic.Tasks{}
		err := json.Unmarshal(msg.Body, &tasks)
		if err != nil {
			return err
		}

		logrus.WithFields(logrus.Fields{
			"from":      msg.FromUUID,
			"schedules": tasks,
		}).Debug("Received new schedules")

		wsh.Store.AddOrUpdateScheduledTasks(tasks)
	case "update-savedstates":
		ss := logic.SavedStates{}
		err := json.Unmarshal(msg.Body, &ss)
		if err != nil {
			return err
		}

		logrus.WithFields(logrus.Fields{
			"from":        msg.FromUUID,
			"savedstates": ss,
		}).Debug("Received new savedstates")

		wsh.Store.AddOrUpdateSavedStates(ss)
	default:
		logrus.WithFields(logrus.Fields{
			"type": msg.Type,
		}).Warnf("Received unknown message")
	}
	return nil
}

func (wsh *secureWebsocketHandler) Connect(s interfaces.MelodySession, r *http.Request, keys map[string]interface{}) error {
	proto, _ := s.Get(websocket.KeyProtocol.String())
	id, _ := s.Get(websocket.KeyID.String())
	logrus.Debugf("ws handle secure connect with id %s (%s)", id, proto)

	// Send a list of all nodes if its a webgui
	switch proto {
	case "node":
		n := wsh.Store.GetNode(id.(string))
		if n == nil {
			// New node, register the new node
			t, _ := s.Get("type")
			n = &models.Node{
				UUID:       id.(string),
				Type:       t.(string),
				Connected_: true,
			}
			wsh.Store.AddOrUpdateNode(n)
		}

		msg, err := models.NewMessage("setup", n)
		if err != nil {
			return err
		}
		msg.WriteTo(s)
	}

	return nil
}

func (wsh *secureWebsocketHandler) Disconnect(s interfaces.MelodySession) error {
	id, _ := s.Get(websocket.KeyID.String())
	n := wsh.Store.GetNode(id.(string))
	if n != nil {
		n.SetConnected(false)
	}
	return nil
}
