package main

type nodeSpecific struct {
	Port       string
	ListenPort string
	Devices    deviceList
}

type deviceList []*Device

func (d deviceList) maxId() int {
	i := 0
	for _, v := range d {
		id := v.Id
		if id > i {
			i = id
		}
	}
	return i
}

// URL containing urls for dim and on and off
type URL struct {
	Level string
	On    string
	Off   string
}

type Device struct {
	Name string
	Id   int
	UUID string
	Url  *URL
}
