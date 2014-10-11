package protocol

type Command struct {
	Cmd  string   `json:"cmd"`
	Args []string `json:"args"`
}
