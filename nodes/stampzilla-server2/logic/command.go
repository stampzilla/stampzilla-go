package logic

type Command interface {
	Run(abort <-chan struct{})
}
type command struct {
	//Command *protocol.Command `json:"command"`
	//Uuid_   string            `json:"uuid"`
	//nodes   serverprotocol.Searchable
}

func NewCommand() *command {
	return &command{}
}

//func (c *command) Uuid() string {
//return c.Uuid_
//}

func (c *command) Run(abort <-chan struct{}) {
	//TODO implement
}
