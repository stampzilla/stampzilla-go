package logic

type Action interface {
	Run()
	Uuid() string
	Name() string
}

type action struct {
	Name_    string     `json:"name"`
	Uuid_    string     `json:"uuid"`
	Commands []*command `json:"commands"`
}

func (a *action) Uuid() string {
	return a.Uuid_
}
func (a *action) Name() string {
	return a.Name_
}
func (a *action) Run() {
	for _, v := range a.Commands {
		v.Run()
	}
}
