package main

import (
	"bytes"
	"strconv"
	"time"
)

type Packet struct {
	Zone     int
	ID       string
	Name     string
	Action   string
	Time     time.Time
	SystemID int
}

func decode(data []byte) (*Packet, error) {
	tmp := bytes.Split(data[23:], []byte("|"))

	systemid, err := strconv.Atoi(string(bytes.Replace(tmp[0], []byte("[#"), []byte{}, 1)))
	if err != nil {
		return nil, err
	}

	name := bytes.Split(tmp[4], []byte{0xa6}) // this is ¦ in Kök IR¦ZONE¦1¦Larm

	zone, err := strconv.Atoi(string(name[2]))
	if err != nil {
		return nil, err
	}

	t, err := time.Parse("15040502012006", string(tmp[1]))
	if err != nil {
		return nil, err
	}

	return &Packet{
		ID:       string(tmp[3]),
		Name:     toUtf8(name[0]),
		Zone:     zone,
		Action:   getAction(tmp[2]),
		SystemID: systemid,
		Time:     t,
	}, nil
}

func getAction(action []byte) string {
	switch string(action) {
	case "ZO":
		return "open"
	case "ZC":
		return "close"
	}

	return "unknown"
}

func toUtf8(data []byte) string {
	buf := make([]rune, len(data))
	for i, b := range data {
		buf[i] = rune(b)
	}

	return string(buf)
}
