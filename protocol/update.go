package protocol

import (
	"encoding/json"

	log "github.com/cihub/seelog"
)

// Type describes what type of Update package it is
type Type byte

const (
	// TypeUpdateNode is a package from the node to the server with node specific information
	TypeUpdateNode Type = iota
	// TypeUpdateState is sent from node to server when the state changes
	TypeUpdateState
	// TypeUpdateDevices is sent from node to server with a new list of available devices
	TypeUpdateDevices
	// TypeNotification is a message to the user from a node
	TypeNotification
	// TypePing is a keep alive message
	TypePing
	// TypePong is the answer to the keep alive message
	TypePong
	// TypeCommand is sent from the server to the node with a task for the node do. For example turn on a light.
	TypeCommand
	// TypeDeviceConfigSet is sent from server to the node when a user changes a parameter in a device
	TypeDeviceConfigSet
)

func (t Type) String() string {
	s, ok := map[Type]string{
		TypeUpdateNode:      "UpdateNode",
		TypeUpdateState:     "UpdateState",
		TypeUpdateDevices:   "UpdateDevices",
		TypeNotification:    "Notification",
		TypePing:            "Ping",
		TypePong:            "Pong",
		TypeCommand:         "Command",
		TypeDeviceConfigSet: "DeviceConfigSet",
	}[t]

	if !ok {
		return "invalid Type"
	}
	return s
}

// Update is the outer most layer in the communication between node and server. The purpose is to encapsulate the data to transfer with a type.
type Update struct {
	Type Type
	Data *json.RawMessage
}

// NewUpdate provides a new Update struct
func NewUpdate() *Update {
	return &Update{}
}

// NewUpdateWithData encodes data in to a new Update struct that can be used to communicate between node and server
func NewUpdateWithData(t Type, data interface{}) *Update {
	var jsonByte json.RawMessage

	switch j := data.(type) {
	case json.RawMessage:
		jsonByte = j
	case *json.RawMessage:
		jsonByte = *j
	default:
		jsonStr, err := json.Marshal(data)
		if err != nil {
			log.Error(err)
			return nil
		}
		jsonByte = json.RawMessage(jsonStr)
	}

	u := &Update{
		Type: t,
		Data: &jsonByte,
	}

	return u
}

// ToJSON encodes the update package to json bytes
func (u Update) ToJSON() ([]byte, error) {
	return json.Marshal(u)
}
