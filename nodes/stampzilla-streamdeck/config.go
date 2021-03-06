package main

import "sync"

type config struct {
	Pages map[string]page `json:"pages"`
	sync.Mutex
}

type page struct {
	Keys [15]key `json:"keys"`
}

type key struct {
	Name   string `json:"name"`
	Node   string `json:"node"`
	Device string `json:"device"`
	Action string `json:"action"`
	Field  string `json:"field"`
}
