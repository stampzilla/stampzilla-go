package notification

import "sync"

type Sender interface {
	ID() string
	Send(Message) error
}

type Senders struct {
	senders map[string]Sender
	sync.RWMutex
}

func NewSenders() *Senders {
	return &Senders{
		senders: make(map[string]Sender),
	}
}

type MailSender struct {
	ID_      string `json:"id"`
	Type     Type   `json:"type"`
	Server   string `json:"server"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type SmsSender struct {
	ID_    string `json:"id"`
	Type   Type   `json:"type"`
	ApiKey string `json:"apiKey"`
}

func (s *Senders) Add(sender Sender) {
	s.Lock()
	s.senders[sender.ID()] = sender
	s.Unlock()
}

func (s *Senders) Get(id string) Sender {
	s.RLock()
	defer s.RUnlock()
	return s.senders[id]
}

func (s *Senders) All() map[string]Sender {
	s.RLock()
	defer s.RUnlock()
	return s.senders
}

func (s *Senders) Remove(id string) {
	s.Lock()
	delete(s.senders, id)
	s.Unlock()
}
