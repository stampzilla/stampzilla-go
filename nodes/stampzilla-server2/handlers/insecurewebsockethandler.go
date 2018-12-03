package handlers

import (
	"encoding/json"
	"encoding/pem"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/ca"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/interfaces"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/store"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/websocket"
)

type insecureWebsocketHandler struct {
	Store           *store.Store
	Config          *models.Config
	WebsocketSender websocket.Sender
	ca              *ca.CA
}

func NewInSecureWebsockerHandler(store *store.Store, config *models.Config, ws websocket.Sender, ca *ca.CA) WebsocketHandler {
	return &insecureWebsocketHandler{
		Store:           store,
		Config:          config,
		WebsocketSender: ws,
		ca:              ca,
	}
}

func (wsh *insecureWebsocketHandler) Message(msg *models.Message) error {
	logrus.Warn("Unsecure ws sent data: ", msg)

	// client requested certificate. We must approve manually

	if msg.Type == "certificate-signing-request" {
		var body string
		json.Unmarshal(msg.Body, &body)

		//TODO we want to add this to for example store to make it statefull and so the admin can approve the request
		// for now we just approve automaticly and send it directly

		cert := &strings.Builder{}
		err := wsh.ca.CreateCertificateFromRequest(cert, "nodename", []byte(body))
		if err != nil {
			return err
		}

		// send certificate to node
		err = wsh.WebsocketSender.SendToID(msg.FromUUID, "approved-certificate-signing-request", cert.String())
		if err != nil {
			return err
		}

		// send ca to node
		ca := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: wsh.ca.CAX509.Raw})

		err = wsh.WebsocketSender.SendToID(msg.FromUUID, "certificate-authority", string(ca))
		if err != nil {
			return err
		}

	}

	return nil
}

func (wsh *insecureWebsocketHandler) Connect(s interfaces.MelodySession, r *http.Request, keys map[string]interface{}) error {

	logrus.Info("ws handle insecure connect")
	msg, err := models.NewMessage("server-info", models.ServerInfo{
		Name:    wsh.Config.Name,
		UUID:    wsh.Config.UUID,
		TLSPort: wsh.Config.TLSPort,
		Port:    wsh.Config.Port,
	})
	if err != nil {
		return err
	}
	msg.WriteTo(s)

	return nil
}

func (wsh *insecureWebsocketHandler) Disconnect(s interfaces.MelodySession) error {
	id, _ := s.Get("ID")
	wsh.Store.RemoveConnection(id.(string))
	return nil
}
