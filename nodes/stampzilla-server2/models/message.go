package models

import (
	"encoding/json"
	"fmt"

	"github.com/olahol/melody"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/interfaces"
)

type Message struct {
	Type string          `json:"type"`
	Body json.RawMessage `json:"body"`
}

func NewMessage(t string, body interface{}) (*Message, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return &Message{
		Type: t,
		Body: json.RawMessage(b),
	}, nil
}

func ParseMessage(msg []byte) (*Message, error) {
	data := &Message{}
	err := json.Unmarshal(msg, data)
	return data, err
}

func (m *Message) Write(s interfaces.MelodyWriter) error {
	msg, err := m.Encode()
	if err != nil {
		return err
	}

	return s.Write(msg)
}

func (m *Message) WriteWithFilter(mel *melody.Melody, f func(s *melody.Session) bool) error {
	msg, err := m.Encode()
	if err != nil {
		return err
	}

	return mel.BroadcastFilter(msg, f)
}

func (m *Message) Encode() ([]byte, error) {
	msg, err := json.Marshal(m)

	return msg, err
}
func (m *Message) String() string {
	return fmt.Sprintf("Type: %s Body: %s", m.Type, string(m.Body))
}
