package handlers

import (
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/interfaces"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/store"
)

type insecureWebsocketHandler struct {
	Store  *store.Store
	Config *models.Config
}

func NewInSecureWebsockerHandler(store *store.Store, config *models.Config) WebsocketHandler {
	return &insecureWebsocketHandler{
		Store:  store,
		Config: config,
	}
}

func (wsh *insecureWebsocketHandler) Message(msg *models.Message) error {
	logrus.Warn("Unsecure ws sent data: ", msg)
	return nil
}

func (wsh *insecureWebsocketHandler) Connect(s interfaces.MelodySession, r *http.Request, keys map[string]interface{}) error {
	id, _ := s.Get("ID")
	t, _ := s.Get("protocol")
	logrus.Info("ws handle insecure connect")

	wsh.Store.AddOrUpdateConnection(id.(string), &models.Connection{
		Type:       t.(string),
		RemoteAddr: r.RemoteAddr,
		Attributes: keys,
	})

	msg, err := models.NewMessage("server-info", models.ServerInfo{
		Name:    wsh.Config.Name,
		UUID:    wsh.Config.UUID,
		TLSPort: wsh.Config.TLSPort,
		Port:    wsh.Config.Port,
	})
	if err != nil {
		return err
	}
	msg.Write(s)

	return nil
}

func (wsh *insecureWebsocketHandler) Disconnect(s interfaces.MelodySession) error {
	id, _ := s.Get("ID")
	wsh.Store.RemoveConnection(id.(string))
	return nil
}
