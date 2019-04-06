package notification

type Sender interface {
	Send(Message) error
}

type MailSender struct {
	Type     Type   `json:"type"`
	Server   string `json:"server"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type SmsSender struct {
	Type   Type   `json:"type"`
	ApiKey string `json:"apiKey"`
}
