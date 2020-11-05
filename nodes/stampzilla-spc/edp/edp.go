package edp

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
)

type Packet struct {
	Zone     int
	ID       string
	Name     string
	Class    string
	Time     time.Time
	SystemID int
}

func Decode(data []byte) (*Packet, error) {
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
		Class:    string(tmp[2]),
		SystemID: systemid,
		Time:     t,
	}, nil
}

func toUtf8(data []byte) string {
	buf := make([]rune, len(data))
	for i, b := range data {
		buf[i] = rune(b)
	}

	return string(buf)
}

func GenerateDevice(pkg *Packet) *devices.Device {
	state := devices.State{}
	var prefix string

	switch pkg.Class {
	case "ZO": // open zone
		// TODO call it on? or open? or triggered? or active? :)
		state["on"] = true
		prefix = "zone"
	case "ZC": // close zone
		state["on"] = false
		prefix = "zone"
	case "NL": // perimiter arm area
		state["armed"] = true
		state["full"] = false
		prefix = "area"
	case "CG": // close area
		state["armed"] = true
		state["full"] = true
		prefix = "area"
	case "OG": // open area
		state["armed"] = false
		prefix = "area"
	default:
		return nil
	}

	return &devices.Device{
		Type: "sensor",
		ID: devices.ID{
			ID: prefix + "." + pkg.ID,
		},
		Name:   fmt.Sprintf("%s %s", strings.Title(prefix), pkg.Name),
		Online: true,
		State:  state,
	}
}
