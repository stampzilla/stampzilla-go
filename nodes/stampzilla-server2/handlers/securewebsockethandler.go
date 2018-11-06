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
		node := &models.Node{}
		err := json.Unmarshal(msg.Body, node)
		if err != nil {
			return err
		}

		wsh.Store.AddOrUpdateNode(node)
	}
	return nil
}

func (wsh *secureWebsocketHandler) Connect(s interfaces.MelodySession, r *http.Request, keys map[string]interface{}) error {
	t, exists := s.Get("protocol")
	id, exists := s.Get("ID")
	logrus.Infof("ws handle secure connect with id %s", id)

	if exists && t == "gui" {
		msg, err := models.NewMessage("nodes", wsh.Store.GetNodes())
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
