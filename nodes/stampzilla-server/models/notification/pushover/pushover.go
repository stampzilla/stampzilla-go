package pushover

import (
	"encoding/json"
	"fmt"

	pover "github.com/gregdel/pushover"
	"github.com/sirupsen/logrus"
)

var ErrNotImplemented = fmt.Errorf("not implemented")

type PushOver struct {
	Token string `json:"token"`
}

func New(parameters json.RawMessage) *PushOver {
	pb := &PushOver{}

	err := json.Unmarshal(parameters, pb)
	if err != nil {
		logrus.Error(err)
	}
	return pb
}

func (po *PushOver) Release(dest []string, body string) error {
	return ErrNotImplemented
}

func (po *PushOver) Trigger(dest []string, body string) error {
	app := pover.New(po.Token)
	var err error

	for _, userKey := range dest {
		recipient := pover.NewRecipient(userKey)
		message := pover.NewMessage(body)
		message.Title = "Stampzilla"
		_, err = app.SendMessage(message, recipient)
		if err != nil {
			return err
		}
	}
	return err
}

func (po *PushOver) Destinations() (map[string]string, error) {
	return nil, ErrNotImplemented
}
