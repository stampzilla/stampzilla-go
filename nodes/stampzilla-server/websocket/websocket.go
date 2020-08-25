package websocket

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/olahol/melody"
	"github.com/sirupsen/logrus"
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

	Request(to string, msgType string, data interface{}, timeout time.Duration) (json.RawMessage, error)
	Response(*models.Message)
}

type sender struct {
	Melody *melody.Melody

	requestID int
	requests  map[int]chan *models.Message

	sync.RWMutex
}

func NewWebsocketSender(m *melody.Melody) Sender {
	return &sender{
		Melody:   m,
		requests: make(map[int]chan *models.Message),
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

func (ws *sender) Request(to string, msgType string, data interface{}, timeout time.Duration) (json.RawMessage, error) {
	message, err := models.NewMessage(msgType, data)
	if err != nil {
		return nil, err
	}

	c := make(chan *models.Message)

	ws.Lock()
	ws.requestID++
	requestID := ws.requestID
	ws.Unlock()

	message.Request = json.RawMessage(strconv.Itoa(ws.requestID))
	err = ws.sendMessageTo(KeyID, to, message)
	if err != nil {
		return nil, err
	}

	ws.Lock()
	ws.requests[requestID] = c
	ws.Unlock()

	select {
	case resp := <-c:
		spew.Dump(resp)
		ws.requests[requestID] = nil
		return resp.Body, nil
	case <-time.After(timeout):
		close(c)
		ws.requests[requestID] = nil
		return nil, fmt.Errorf("timeout")
	}
}
func (ws *sender) Response(msg *models.Message) {
	requestID, err := strconv.Atoi(string(msg.Request))
	if err != nil {
		logrus.Error(err)
		return
	}

	ws.Lock()
	if ws.requests[requestID] != nil {
		select {
		case ws.requests[requestID] <- msg:
		default:
		}
		ws.requests[requestID] = nil
	}
	ws.Unlock()
}
