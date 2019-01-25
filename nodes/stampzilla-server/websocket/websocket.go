package websocket

import (
	"github.com/olahol/melody"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models"
)

type SessionKey string

const (
	KeyProtocol SessionKey = "protocol"
	KeyID       SessionKey = "ID"
)

func (sk SessionKey) String() string {
	return string(sk)
}

type Sender interface {
	SendToID(to string, msgType string, data interface{}) error
	SendToProtocol(to string, msgType string, data interface{}) error
	BroadcastWithFilter(msgType string, data interface{}, fn func(*melody.Session) bool) error
}

type sender struct {
	Melody *melody.Melody
}

func NewWebsocketSender(m *melody.Melody) Sender {
	return &sender{
		Melody: m,
	}
}

func (ws *sender) sendMessageTo(key SessionKey, to string, msg *models.Message) error {
	return msg.WriteWithFilter(ws.Melody, func(s *melody.Session) bool {
		v, exists := s.Get(key.String())
		return exists && v == to
	})
}
func (ws *sender) SendToID(to string, msgType string, data interface{}) error {
	message, err := models.NewMessage(msgType, data)
	if err != nil {
		return err
	}
	return ws.sendMessageTo(KeyID, to, message)
}

func (ws *sender) SendToProtocol(to string, msgType string, data interface{}) error {
	message, err := models.NewMessage(msgType, data)
	if err != nil {
		return err
	}
	return ws.sendMessageTo(KeyProtocol, to, message)
}
func (ws *sender) BroadcastWithFilter(msgType string, data interface{}, fn func(*melody.Session) bool) error {
	message, err := models.NewMessage(msgType, data)
	if err != nil {
		return err
	}
	return message.WriteWithFilter(ws.Melody, fn)
}
