package websocket

import (
	"github.com/olahol/melody"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
)

type Sender interface {
	SendTo(to string, msgType string, data interface{}) error
}

type sender struct {
	Melody *melody.Melody
}

func NewWebsocketSender(m *melody.Melody) Sender {
	return &sender{
		Melody: m,
	}
}

func (ws *sender) sendMessageTo(to string, msg *models.Message) error {
	return msg.WriteWithFilter(ws.Melody, func(s *melody.Session) bool {
		v, exists := s.Get("ID")
		return exists && v == to
	})
}
func (ws *sender) SendTo(to string, msgType string, data interface{}) error {
	message, err := models.NewMessage(msgType, data)
	if err != nil {
		return err
	}
	return ws.sendMessageTo(to, message)
}
