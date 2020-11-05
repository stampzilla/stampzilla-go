package notification

import "github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models"

type Messages []Message

type Message struct {
	DestinationSelector models.Labels `json:"destinationSelector"`
	Head                string        `json:"head"`
	Body                string        `json:"body"`
}
