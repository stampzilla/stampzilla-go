package logic

import (
	"bytes"
	"encoding/json"
	"os"

	log "github.com/cihub/seelog"
)

type ActionsMapper interface {
	Save(*Actions)
	Load(*Actions)
}

type actionsMapper struct {
}

func newActionsMapper() ActionsMapper {
	return &actionsMapper{}

}

func (am *actionsMapper) Save(actions *Actions) {
	configFile, err := os.Create("actions.json")
	if err != nil {
		log.Error("creating config file", err.Error())
		return
	}
	var out bytes.Buffer
	b, err := json.Marshal(actions)
	if err != nil {
		log.Error("error marshal json", err)
	}
	json.Indent(&out, b, "", "    ")
	out.WriteTo(configFile)
}
func (am *actionsMapper) Load(actions *Actions) {
	file, err := os.Open("actions.json")
	if err != nil {
		log.Warn("opening config file", err.Error())
		return
	}
	jsonParser := json.NewDecoder(file)
	if err = jsonParser.Decode(&actions); err != nil {
		log.Error(err)
	}
}
