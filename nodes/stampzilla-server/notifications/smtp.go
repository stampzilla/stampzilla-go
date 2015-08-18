package notifications

import (
	"sync"

	log "github.com/cihub/seelog"
	"gopkg.in/gomail.v1"
)

type Smtp struct {
	wg       sync.WaitGroup
	Server   string `default:localhost`
	Username string
	Password string
	Port     int    `default:25`
	Sender   string `default:stampzilla`
	To       string `default:stampzilla`
}

func (self *Smtp) Start() {
}

func (self *Smtp) Dispatch(note Notification) {
	self.wg.Add(1)
	defer self.wg.Done()

	msg := gomail.NewMessage()
	msg.SetHeader("From", self.Sender)
	msg.SetHeader("To", self.To)
	msg.SetHeader("Subject", "Hello!")
	msg.SetBody("text/html", "Hello <b>Bob</b> and <i>Cora</i>!")

	mailer := gomail.NewMailer(self.Server, self.Username, self.Password, self.Port)
	if err := mailer.Send(msg); err != nil {
		log.Error("Failed to send mail - ", err)
	}
}
func (self *Smtp) Stop() {
}
