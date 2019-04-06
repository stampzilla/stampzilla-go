package notification

import "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models"

type Destination struct {
	Type         string        `json:"type"`
	Labels       models.Labels `json:"labels"`
	Sender       Sender        `json:"sender"`
	Destinations []string      `json:"destinations"`
}
