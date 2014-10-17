package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"

	log "github.com/cihub/seelog"
	"github.com/stampzilla/stampzilla-go/protocol"
	"github.com/stampzilla/stampzilla-go/stampzilla-server/logic"
	serverprotocol "github.com/stampzilla/stampzilla-go/stampzilla-server/protocol"
)

func netStart(port string) {
	listen, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("listen error", err)
		return
	}

	l := logic.NewLogic()
	rule := l.AddRule("test rule 1")

	rule.AddEnterAction(logic.NewRuleAction(&protocol.Command{"testEnterAction", nil}, "uuid1", nil))
	rule.AddExitAction(logic.NewRuleAction(&protocol.Command{"testExitAction", nil}, "uuid2", nil))

	rule.AddCondition(logic.NewRuleCondition(`Devices.0186ff7d.On`, "==", true))
	rule.AddCondition(logic.NewRuleCondition(`Devices[2].State`, "!=", "OFF"))

	go func() {
		for {
			fd, err := listen.Accept()
			if err != nil {
				fmt.Println("accept error", err)
				return
			}

			go newClient(l, fd)
		}
	}()
}

func newClient(logic *logic.Logic, connection net.Conn) {
	// Recive data
	log.Info("New client connected")
	name := ""
	uuid := ""
	var logicChannel chan string
	for {
		reader := bufio.NewReader(connection)
		decoder := json.NewDecoder(reader)
		var info serverprotocol.Node
		err := decoder.Decode(&info)

		//err = json.Unmarshal(data, &cmd)
		if err != nil {
			if err.Error() == "EOF" {
				log.Info(name, " - Client disconnected")
				if uuid != "" {
					nodes.Delete(uuid)
					close(logicChannel)
				}
				//TODO be able to not send everything always. perhaps implement remove instead of all?
				clients.messageOtherClients(&Message{Type: "all", Data: nodes.All()})
				return
			}
			log.Warn("Not disconnect but error: ", err)
			//return here?
		} else {
			name = info.Name
			uuid = info.Uuid
			info.SetConn(connection)

			if logicChannel == nil {
				logicChannel = logic.ListenForChanges(uuid)
			}

			nodes.Add(&info)
			log.Info(info.Name, " - Got update on state")
			clients.messageOtherClients(&Message{Type: "singlenode", Data: info})
			//Send to logic for evaluation

			state, _ := json.Marshal(info.State)
			logicChannel <- string(state)
		}

	}

}

// Handle a client
//func newClientOld(c net.Conn) {
//log.Info("New client connected")
//id := ""
//for {
//buf := make([]byte, 51200)
//nr, err := c.Read(buf)
//if err != nil {
//log.Info(id, " - Client disconnected")
//if id != "" {
//delete(nodes, id)
//}
////TODO be able to not send everything always.
//clients.messageOtherClients(&Message{"all", nodes})
//return
//}

////TODO: Handle when multiple messages gets concated ex: msg}{msg2
////  TODO: see possible solution above in the new improved newClient function :)  (jonaz) <Fri 10 Oct 2014 10:04:43 AM CEST>

//data := buf[0:nr]

//var info Node
//err = json.Unmarshal(data, &info)
//if err != nil {
//log.Warn(err, " -->", string(data), "<--")
//} else {
//id = info.Id
//nodes[info.Id] = info
//nodesConnection[info.Id] = &NodeConnection{conn: c, wait: nil}

//log.Info(info.Id, " - Got update on state")

//if nodesConnection[info.Id].wait != nil {
//select {
//case nodesConnection[info.Id].wait <- false:
//close(nodesConnection[info.Id].wait)
//nodesConnection[info.Id].wait = nil
//default:
//}
//}

//clients.messageOtherClients(&Message{"singlenode", nodes[info.Id]})
//}
//}
//}
