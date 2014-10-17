package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"

	log "github.com/cihub/seelog"
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
	l.SetNodes(nodes)
	//rule := l.AddRule("test rule 1")
	//rule.AddEnterAction(logic.NewRuleAction(&protocol.Command{"testar", []string{"0186ff7d"}}, "enocean"))
	//rule.AddExitAction(logic.NewRuleAction(&protocol.Command{"testar", []string{"0186ff7d"}}, "enocean"))
	//rule.AddCondition(logic.NewRuleCondition(`Devices.0186ff7d.On`, "==", true))
	//rule.AddCondition(logic.NewRuleCondition(`Devices[2].State`, "!=", "OFF"))

	//TODO see logic.go
	l.RestoreRulesFromFile("rules.json")

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
