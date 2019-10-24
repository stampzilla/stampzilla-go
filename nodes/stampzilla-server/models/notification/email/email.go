package email

import (
	"encoding/json"
	"fmt"
	"net/smtp"

	"github.com/davecgh/go-spew/spew"
)

type EmailSender struct {
	Server   string `json:"server"`
	Port     int    `json:"port"`
	From     string `json:"from"`
	Password string `json:"password"`
	send     func(string, smtp.Auth, string, []string, []byte) error
}

func New(parameters json.RawMessage) *EmailSender {
	es := &EmailSender{send: smtp.SendMail}

	json.Unmarshal(parameters, es)

	return es
}

func (es *EmailSender) Trigger(dest []string, body string) error {
	return es.notify(true, dest, body)
}

func (es *EmailSender) Release(dest []string, body string) error {
	return es.notify(false, dest, body)
}

func (es *EmailSender) notify(trigger bool, dest []string, body string) error {
	event := "Triggered"
	if !trigger {
		event = "Released"
	}

	msg := "From: " + es.From + "\n" +
		"Subject: stampzilla - " + event + "\n\n" +
		body

	spew.Dump(smtp.PlainAuth("", es.From, es.Password, es.Server))

	return es.send(fmt.Sprintf("%s:%d", es.Server, es.Port),
		smtp.PlainAuth("", es.From, es.Password, es.Server),
		es.From, dest, []byte(msg))
}

func (es *EmailSender) Destinations() (map[string]string, error) {
	return nil, fmt.Errorf("not implemented")
}
