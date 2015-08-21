package protocol

type Command struct {
	Ping   bool     `json:",omitempty"`
	Cmd    string   `json:"cmd"`
	Args   []string `json:"args"`
	Params []string `json:"params"`
}
