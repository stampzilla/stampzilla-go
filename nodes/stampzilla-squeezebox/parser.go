package main

import (
	"net"
	"net/url"
	"strconv"
	"strings"

	log "github.com/cihub/seelog"
)

type parser struct {
}

func (p *parser) ParseType(cmd string) (string, string) {
	d := strings.Fields(cmd)
	if len(d) > 0 {
		return d[0], strings.Join(d[1:], " ")
	}
	return "", ""
}

func (p *parser) parse(cmd string) map[string]string {
	d := strings.Fields(cmd)
	ret := make(map[string]string)
	for _, v := range d {
		key, value := p.split(v)
		ret[key] = value
	}
	return ret
}

func (p *parser) split(v string) (string, string) {
	unescaped, err := url.QueryUnescape(v)
	if err != nil {
		log.Error(err)
		return "", ""
	}
	splitted := strings.Split(unescaped, ":")
	if len(splitted) < 2 {
		return "value", splitted[0]
	}
	return splitted[0], strings.Join(splitted[1:], ":")
}

func (p *parser) Players(cmd string) []*Device {
	players := []*Device{}
	d := strings.Fields(cmd)
	var player *Device
	for _, v := range d {

		key, value := p.split(v)

		switch key {
		case "playerid":
			player = NewDevice(value, "")
			players = append(players, player)
			continue
		case "ip":
			addr, _ := net.ResolveTCPAddr("tcp", value)
			player.Ip = addr.IP
			continue
		case "name":
			player.Name = value
			continue
		}
	}

	return players
}

func (p *parser) MixerVolume(oldVolume int, cmd string) int {
	value := p.parse(cmd)
	v := value["value"]
	prefix := string(v[0])
	if prefix == "+" || prefix == "-" {
		v, err := strconv.Atoi(v)
		if err != nil {
			log.Error(err)
			return oldVolume
		}
		return oldVolume + v
	}
	vol, err := strconv.Atoi(v)
	if err != nil {
		log.Error(err)
		return oldVolume
	}
	return vol
}
func (p *parser) Song(cmd string) string {
	fields := strings.Fields(cmd)

	if len(fields) > 2 {
		value, err := url.QueryUnescape(fields[2])
		if err != nil {
			log.Error(err)
			return ""
		}
		return value
	}

	return ""
}
func (p *parser) Power(cmd string) bool {
	fields := strings.Fields(cmd)
	if len(fields) > 1 {
		if fields[1] == "1" {
			return true
		}
	}
	return false
}
