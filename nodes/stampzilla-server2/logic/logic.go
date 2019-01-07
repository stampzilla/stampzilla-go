package logic

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/interfaces"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models/devices"
)

func main() {
	fmt.Println("vim-go")
}

type Logic struct {
	StateSyncer interfaces.StateSyncer
	StateStore  *SavedStateStore
	// TODO MAJOR important! must move rules storage to store.Store and load them from there when we start logic so we dont get circular dependencies! :(
	Rules              map[string]*Rule
	devices            *devices.List
	ActionProgressChan chan ActionProgress
	sync.RWMutex
}

type ActionProgress struct {
	Address string `json:"address"`
	Uuid    string `json:"uuid"`
	Step    int    `json:"step"`
}

func NewLogic(s interfaces.StateSyncer) *Logic {
	l := &Logic{
		devices:            devices.NewList(),
		ActionProgressChan: make(chan ActionProgress, 100),
		Rules:              make(map[string]*Rule),
		StateSyncer:        s,
		StateStore:         NewSavedStateStore(),
	}
	return l
}

func (l *Logic) AddRule(name string) *Rule {
	r := &Rule{Name_: name, Uuid_: uuid.New().String()}
	l.Lock()
	defer l.Unlock()
	l.Rules[r.Uuid()] = r
	return r
}

func (l *Logic) UpdateDevice(dev *devices.Device) {
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
				rule.Run(l.StateStore, l.StateSyncer)
				continue
			}
			rule.SetActive(false)

		}
	}
}

func (l *Logic) evaluateRule(r *Rule) bool {
	rules := make(map[string]bool)
	for _, v := range l.Rules {
		rules[v.Uuid()] = v.Active()
	}
	result, err := r.Eval(l.devices, rules)
	if err != nil {
		logrus.Errorf("Error evaluating rule %s: %s", r.Uuid(), err.Error())
		return false
	}
	return result
}

func (l *Logic) Save(path string) {
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

func (l *Logic) Load(path string) {
	configFile, err := os.Open(path)
	if err != nil {
		logrus.Warn("opening config file", err.Error())
		return
	}

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&l.Rules); err != nil {
		logrus.Error(err)
	}

	//TODO loop over rules and generate UUIDs if needed. If it was needed save the rules again

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
