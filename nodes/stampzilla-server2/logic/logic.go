package logic

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
)

func main() {
	fmt.Println("vim-go")
}

type Logic struct {
	Rules              []*Rule
	devices            *models.Devices
	ActionProgressChan chan ActionProgress
	sync.RWMutex
}

func NewLogic() *Logic {
	l := &Logic{
		devices:            models.NewDevices(),
		ActionProgressChan: make(chan ActionProgress, 100),
	}
	return l
}

func (l *Logic) AddRule(name string) *Rule {
	r := &Rule{Name_: name, Uuid_: uuid.New().String()}
	l.Lock()
	defer l.Unlock()
	l.Rules = append(l.Rules, r)
	return r
}

func (l *Logic) UpdateDevice(dev *models.Device) {
	if oldDev := l.devices.Get(dev.Node, dev.ID); oldDev != nil {
		diff := oldDev.State.Diff(dev.State)
		if len(diff) > 0 {
			//oldDev.Lock() // TODO check if needed with -race
			for k, v := range diff {
				oldDev.State[k] = v
			}
			//oldDev.Unlock()
		}
		return
	}
	l.devices.Add(dev)
}

func (l *Logic) EvaluateRules() {
	for _, rule := range l.Rules {
		evaluation := l.evaluateRule(rule)
		if evaluation != rule.Active() {
			rule.SetActive(evaluation)
			if evaluation {
				logrus.Info("Rule: ", rule.Name(), " (", rule.Uuid(), ") - running enter actions")
				//rule.RunEnter(l.ActionProgressChan)
				rule.SetActive(true)
				continue
			}
			rule.SetActive(false)

		}
	}
}

func (l *Logic) evaluateRule(r *Rule) bool {
	return false
}

func (l *Logic) SaveRulesToFile(path string) {
	configFile, err := os.Create(path)
	if err != nil {
		logrus.Error("creating config file", err.Error())
		return
	}
	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "\t")
	err = encoder.Encode(l.Rules)
	if err != nil {
		logrus.Error("error marshal json", err)
	}
}

func (l *Logic) LoadRulesFromFile(path string) {
	configFile, err := os.Open(path)
	if err != nil {
		logrus.Warn("opening config file", err.Error())
		return
	}

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&l.Rules); err != nil {
		logrus.Error(err)
	}

}

/*

new way:
{
	"name": "All off",
	"enabled": true,
	"active": true|false,
	"uuid": "e8092b86-1261-44cd-ab64-38121df58a79",
	"expression": "devices["asdf.123"].on == true && devices["asdf.123"].temperature > 20.0",
	"conditions": {
		"id på annan regel": true,
	}
	"for": "5m", // after for we must evaluate expression again
	"actions": [ // save index on where we are in running actions and send to gui over websocket
		"1m", //sleep , if rule stops being active during actions run. Stopp running!
		"c7d352bb-23f4-468c-b476-f76599c09a0d"
	]
},

savedState{
	"c7d352bb-23f4-468c-b476-f76599c09a0d": {
			"name": "tänd allt",
			"uuid": "c7d352bb-23f4-468c-b476-f76599c09a0d",
			"state": {
				"a.1" :{
					"on":true
				},
				"a.2" :{
					"on":true
				}
			}
		}
	}

old way:

{
	"name": "All off",
	"enabled": true,
	"uuid": "e8092b86-1261-44cd-ab64-38121df58a79",
	"conditions": [
		{
			"statePath": "Devices.fefe749b.Status",
			"comparator": "==",
			"value": "R1B0",
			"uuid": "efd2bd24-ac50-4147-bdf9-da3dd12c8f8a"
		}
	],
	"enterActions": [
		"c7d352bb-23f4-468c-b476-f76599c09a0d"
	]
},

*/
