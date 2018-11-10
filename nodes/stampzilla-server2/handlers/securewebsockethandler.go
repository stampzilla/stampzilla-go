package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/interfaces"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/store"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/websocket"
)

type secureWebsocketHandler struct {
	Store           *store.Store
	Config          *models.Config
	WebsocketSender websocket.Sender
}

func NewSecureWebsockerHandler(store *store.Store, config *models.Config, ws websocket.Sender) WebsocketHandler {
	return &secureWebsocketHandler{
		Store:           store,
		Config:          config,
		WebsocketSender: ws,
	}
}

func (wsh *secureWebsocketHandler) Message(msg *models.Message) error {
	switch msg.Type {
	case "update-node":
		node := &models.Node{
			UUID: msg.FromUUID,
		}
		err := json.Unmarshal(msg.Body, node)
		if err != nil {
			return err
		}

		wsh.Store.AddOrUpdateNode(node)

		for _, dev := range node.Devices {
			wsh.Store.AddOrUpdateDevice(dev)
		}

	case "setup-node":
		node := &models.Node{}
		err := json.Unmarshal(msg.Body, node)
		if err != nil {
			return err
		}

		wsh.Store.AddOrUpdateNode(node)
		err = wsh.Store.WriteToDisk()
		if err != nil {
			return err
		}
		wsh.WebsocketSender.SendTo(node.UUID, "setup", node)
	}
	return nil
}

func (wsh *secureWebsocketHandler) Connect(s interfaces.MelodySession, r *http.Request, keys map[string]interface{}) error {
	proto, exists := s.Get("protocol")
	id, exists := s.Get("ID")
	t, _ := s.Get("type")
	logrus.Infof("ws handle secure connect with id %s (%s)", id, proto)

	// Send a list of all nodes if its a webgui
	if exists && proto == "gui" {
		msg, err := models.NewMessage("nodes", wsh.Store.GetNodes())
		if err != nil {
			return err
		}
		msg.Write(s)
	}

	// Send node setup if its a node
	if exists && proto == "node" {
		n := wsh.Store.GetNode(id.(string))
		if n == nil {
			// New node, register the new node
			n = &models.Node{
				UUID: id.(string),
				Type: t.(string),
			}
			wsh.Store.AddOrUpdateNode(n)
		}

		msg, err := models.NewMessage("setup", n)
		if err != nil {
			return err
		}
		msg.Write(s)
	}

	return nil
}

func (wsh *secureWebsocketHandler) Disconnect(s interfaces.MelodySession) error {
	id, _ := s.Get("ID")
	wsh.Store.RemoveConnection(id.(string))
	return nil
}
