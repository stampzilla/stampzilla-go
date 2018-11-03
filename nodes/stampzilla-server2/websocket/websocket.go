package websocket

import (
	"github.com/olahol/melody"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
)

type Sender interface {
	SendMessageTo(to string, msg *models.Message) error
}

type sender struct {
	Melody *melody.Melody
}

func NewWebsocketSender(m *melody.Melody) Sender {
	return &sender{
		Melody: m,
	}
}

func (ws *sender) SendMessageTo(to string, msg *models.Message) error {
	return msg.WriteWithFilter(ws.Melody, func(s *melody.Session) bool {
		v, exists := s.Get("ID")
		return exists && v == to
	})
}
