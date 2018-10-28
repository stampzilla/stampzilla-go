package main

import (
	"encoding/json"

	"github.com/olahol/melody"
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

func (m *Message) Write(s *melody.Session) error {
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
