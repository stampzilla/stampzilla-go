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
	Area     int
	ID       string
	Name     string
	UserName string
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

	t, err := time.Parse("15040502012006", string(tmp[1]))
	if err != nil {
		return nil, err
	}

	extraInfo := bytes.Split(tmp[4], []byte{0xa6}) // this is ¦ in Kök IR¦ZONE¦1¦Larm

	pkt := &Packet{
		ID:       string(tmp[3]),
		Name:     toUtf8(extraInfo[0]),
		Class:    string(tmp[2]),
		SystemID: systemid,
		Time:     t,
	}

	if len(extraInfo) < 3 {
		return pkt, nil
	}

	if len(extraInfo) == 4 && string(extraInfo[1]) == "ZONE" {
		// [#1000|07442202062022|FA|1|Brandvarnare¦ZONE¦2¦Brandlarm||0]
		i, err := strconv.Atoi(string(extraInfo[2]))
		if err != nil {
			return nil, err
		}
		pkt.Area = i
	}
	if len(extraInfo) == 3 {
		// E2[#1000|19531104112020|OG|1|Larm¦Jonas¦1||0] disarm area
		pkt.UserName = string(extraInfo[1])
	}

	return pkt, nil
}

func toUtf8(data []byte) string {
	buf := make([]rune, len(data))
	for i, b := range data {
		buf[i] = rune(b)
	}

	return string(buf)
}

// GenerateDevice returns stampzilla device and area id if the packet had that.
func GenerateDevice(pkg *Packet) (*devices.Device, *devices.Device) {
	state := devices.State{}
	var prefix string
	var name string
	var updateAreaDev bool

	switch pkg.Class {
	case "ZO": // open zone
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
	case "FA": // fire alarm
		state["fire"] = true
		prefix = "zone"
		updateAreaDev = true
	case "FR": // fire restoral
		state["fire"] = false
		prefix = "zone"
		updateAreaDev = true
	case "BA": // Burglary Alarm
		state["alarm"] = true
		prefix = "zone"
		updateAreaDev = true
	case "BV": // Burglary Verified
		state["verified"] = true
		prefix = "area"
	case "BR": // Burglary Restoral
		state["alarm"] = false
		state["verified"] = false
		prefix = "zone"
		updateAreaDev = true
	case "BT": // Burglary Trouble
		state["error"] = true
		prefix = "zone"
	case "BJ": // Burglary Trouble Restore
		state["error"] = false
		prefix = "zone"
	case "YS": // Communications Trouble
		// [#1000|19480006062022|YS|1|Telelinjefel\xa61||0]
		state["error"] = true
		prefix = "modem"
		name = "Modem"
	case "YK": // Communications Restoral
		// [#1000|21023306062022|YK|1|Telelinjefel \xe5terst\xe4llt\xa61||0]
		state["error"] = false
		prefix = "modem"
		name = "Modem"
	default:
		return nil, nil
	}

	var areaDev *devices.Device

	if updateAreaDev {
		areaDev = &devices.Device{
			Type: "sensor",
			ID: devices.ID{
				ID: "area." + strconv.Itoa(pkg.Area),
			},
			Online: true,
			State:  state,
		}
	}

	if name == "" {
		name = fmt.Sprintf("%s %s", strings.Title(prefix), pkg.Name)
	}

	return &devices.Device{
		Type: "sensor",
		ID: devices.ID{
			ID: prefix + "." + pkg.ID,
		},
		Name:   name,
		Online: true,
		State:  state,
	}, areaDev
}
