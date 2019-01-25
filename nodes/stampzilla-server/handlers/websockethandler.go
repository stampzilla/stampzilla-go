package handlers

import (
	"net/http"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/interfaces"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models"
)

type WebsocketHandler interface {
	Message(s interfaces.MelodySession, msg *models.Message) error
	Connect(s interfaces.MelodySession, r *http.Request, keys map[string]interface{}) error
	Disconnect(s interfaces.MelodySession) error
}
