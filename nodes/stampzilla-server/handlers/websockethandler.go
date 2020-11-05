package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/interfaces"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models"
)

type WebsocketHandler interface {
	Message(s interfaces.MelodySession, msg *models.Message) (json.RawMessage, error)
	Connect(s interfaces.MelodySession, r *http.Request, keys map[string]interface{}) error
	Disconnect(s interfaces.MelodySession) error
}
