package main

import (
	"net/url"
	"strings"

	"github.com/stampzilla/stampzilla-go/nodes/basenode"
	"github.com/stampzilla/stampzilla-go/protocol"
)

type Processor struct {
	node           *protocol.Node
	connection     basenode.Connection
	state          *State
	squeezeboxConn *squeezebox
}

func NewProcessor(node *protocol.Node, connection basenode.Connection, state *State, sq *squeezebox) *Processor {
	return &Processor{
		node:           node,
		connection:     connection,
		state:          state,
		squeezeboxConn: sq,
	}
}

func (p *Processor) ProcessSqueezeboxCommand(cmd string) {
	parser := &parser{}

	//Global stuff
	t, rest := parser.ParseType(cmd)
	switch t {
	case "players":
		players := parser.Players(cmd)
		for _, player := range players {
			p.state.AddDevice(player)
			p.squeezeboxConn.Send(player.Id + " mixer volume ?")
			p.squeezeboxConn.Send(player.Id + " power ?")
			//p.squeezeboxConn.Send(player.Id.String() + " status 0 999 subscribe:0")
		}
		p.connection.Send(p.node.Node())

	default:
		// Player specific stuff
		id, _ := url.QueryUnescape(t)
		if player := p.state.Device(id); player != nil {

			switch {
			case strings.Contains(rest, "mixer volume"):
				volume := parser.MixerVolume(player.Volume, rest)
				player.Volume = volume
				p.connection.Send(p.node.Node())
			case strings.Contains(rest, "playlist newsong"):
				player.Title = parser.Song(rest)
				player.Playing = true
				p.connection.Send(p.node.Node())
			case rest == "pause" || rest == "stop":
				player.Playing = false
				p.connection.Send(p.node.Node())
			case rest == "play":
				player.Playing = true
				p.connection.Send(p.node.Node())
			case strings.Contains(rest, "power"):
				player.Power = parser.Power(rest)
				p.connection.Send(p.node.Node())
			}

		}
	}

	// 00%3A04%3A20%3A17%3A7f%3A2b client disconnect
	// 00%3A04%3A20%3A17%3A7f%3A2b client reconnect
	// 00%3A04%3A20%3A1f%3A06%3Ab2 pause
	// 00%3A04%3A20%3A1f%3A06%3Ab2 play
	// 00%3A04%3A20%3A1f%3A06%3Ab2 mixer volume %2B2
	// 00%3A04%3A20%3A1f%3A06%3Ab2 mixer volume -2
	// 00%3A04%3A20%3A1f%3A06%3Ab2 status 0 999 subscribe player_name%3AK%C3%B6ket player_connected%3A1 player_ip%3A192.168.13.50%3A46550 power%3A1 signalstrength%3A88 mode%3Aplay remote%3A1 current_title%3AFM03 time%3A6800.16706204224 rate%3A1 mixer%20volume%3A20 mixer%20bass%3A0 mixer%20treble%3A0 playlist%20repeat%3A0 playlist%20shuffle%3A0 playlist%20mode%3Aoff seq_no%3A0 playlist_cur_index%3A0 playlist_timestamp%3A1443945330.83546 playlist_tracks%3A1 remoteMeta%3AHASH(0x10324aa8) playlist%20index%3A0 id%3A-208529976 title%3ABudapest artist%3AGeorge%20Ezra duration%3A0)
}
func (p *Processor) ProcessServerCommand(cmd protocol.Command) {

	player := p.state.Device(cmd.Args[0])
	if player == nil {
		return
	}
	switch cmd.Cmd {
	case "on":
		p.squeezeboxConn.SendTo(player, "power 1")
	case "off":
		p.squeezeboxConn.SendTo(player, "power 0")
	case "pause":
		p.squeezeboxConn.SendTo(player, "pause")
	case "stop":
		p.squeezeboxConn.SendTo(player, "stop")
	case "play":
		if len(cmd.Args) > 1 {
			p.squeezeboxConn.SendTo(player, "play "+url.QueryEscape(cmd.Args[1]))
			return
		}
		p.squeezeboxConn.SendTo(player, "play")
	case "volume":
		if len(cmd.Args) > 1 {
			p.squeezeboxConn.SendTo(player, "mixer volume "+cmd.Args[1])
		}

	}
}
