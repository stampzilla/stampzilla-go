package protocol

type Action struct { /*{{{*/
	Id        string
	Name      string
	Arguments []string
} /*}}}*/

func NewAction(id, name string, args []string) *Action {
	return &Action{id, name, args}
}
