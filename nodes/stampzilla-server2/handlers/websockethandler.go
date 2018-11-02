package handlers

import (
	"net/http"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/interfaces"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
)

type WebsocketHandler interface {
	Message(msg *models.Message) error
	Connect(s interfaces.MelodySession, r *http.Request, keys map[string]interface{}) error
	Disconnect(s interfaces.MelodySession) error
}
