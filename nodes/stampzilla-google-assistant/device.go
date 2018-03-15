package main

type deviceList map[string]*Device

// URL containing urls for dim and on and off
type URL struct {
	Level string
	On    string
	Off   string
}

type Device struct {
	Name string
	ID   string
	Url  *URL
}
