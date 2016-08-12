package protocol

import (
	"encoding/json"

	log "github.com/cihub/seelog"
)

type Type byte

const (
	UpdateNode Type = iota
	UpdateState
	UpdateDevices
	Notification
	Ping
	Pong
)

func (t Type) String() string {
	s, ok := map[Type]string{UpdateNode: "UpdateNode", UpdateState: "UpdateState", UpdateDevices: "UpdateDevices"}[t]
	if !ok {
		return "invalid Type"
	}
	return s
}

type Update struct {
	Type Type
	Data *json.RawMessage
}

func NewUpdate() *Update {
	return &Update{}
}
func NewUpdateWithData(t Type, data interface{}) *Update {

	jsonStr, err := json.Marshal(data)
	if err != nil {
		log.Error(err)
		return nil
	}

	jsonByte := json.RawMessage(jsonStr)
	u := &Update{
		Type: t,
		Data: &jsonByte,
	}

	return u
}
