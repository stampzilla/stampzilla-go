package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models/devices"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/websocket"
)

func main() {
	fmt.Println("vim-go")
}

type Rules map[string]*Rule

// Logic is the main struct
type Logic struct {
	StateStore *SavedStateStore
	Rules      map[string]*Rule
	devices    *devices.List
	//ActionProgressChan chan ActionProgress
	sync.RWMutex
	sync.WaitGroup
	c               chan func()
	WebsocketSender websocket.Sender
}

/*
type ActionProgress struct {
	Address string `json:"address"`
	Uuid    string `json:"uuid"`
	Step    int    `json:"step"`
}
*/

// NewLogic returns a new logic that is ready to use
func New(sss *SavedStateStore, websocketSender websocket.Sender) *Logic {
	l := &Logic{
		devices: devices.NewList(),
		//ActionProgressChan: make(chan ActionProgress, 100),
		Rules:           make(map[string]*Rule),
		StateStore:      sss,
		c:               make(chan func()),
		WebsocketSender: websocketSender,
	}
	return l
}

// AddRule adds a new logic rule. Mainly used in tests
func (l *Logic) AddRule(name string) *Rule {
	r := &Rule{Name_: name, Uuid_: uuid.New().String()}
	l.Lock()
	defer l.Unlock()
	l.Rules[r.Uuid()] = r
	return r
}

func (l *Logic) GetRules() Rules {
	l.RLock()
	defer l.RUnlock()
	return l.Rules
}

// SetRules overwrites all rules in logic
func (l *Logic) SetRules(rules Rules) {
	l.Lock()
	l.Rules = rules
	l.Unlock()
}

// Start starts the logic worker
func (l *Logic) Start(ctx context.Context) {
	l.Add(1)
	logrus.Info("logic: starting worker")
	go l.worker(ctx)
}

func (l *Logic) worker(ctx context.Context) {
	defer l.Done()
	for {
		select {
		case f := <-l.c:
			f()
			l.EvaluateRules()
		case <-ctx.Done():
			logrus.Info("logic: stopping worker")
			return
		}
	}
}

// UpdateDevice update the state in the logic store with the new state from the device
func (l *Logic) UpdateDevice(dev *devices.Device) {
	l.c <- func() {
		l.updateDevice(dev)
	}
}
func (l *Logic) updateDevice(dev *devices.Device) {
	if oldDev := l.devices.Get(dev.ID); oldDev != nil {
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
			//TODO implement for 5m here. Do not run or set active until 5m passed.
			// go run sleep 5m. cancel go routine if we have new evaluation. After 5m run evaluateRule again before rule.Run

			rule.SetActive(evaluation)
			if evaluation {
				logrus.Info("Rule: ", rule.Name(), " (", rule.Uuid(), ") - running actions")
				l.Add(1)
				go func() {
					rule.Run(l.StateStore, l.WebsocketSender)
					l.Done()
				}()
			} else {
				rule.Cancel()
			}
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

func (l *Logic) Save(path string) error {
	configFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("logic: error saving rules: %s", err.Error())
	}
	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "\t")
	l.Lock()
	defer l.Unlock()
	err = encoder.Encode(l.Rules)
	if err != nil {
		return fmt.Errorf("logic: error saving rules: %s", err.Error())
	}
	return nil
}

func (l *Logic) Load(path string) error {
	logrus.Debug("logic: loading rules from ", path)
	configFile, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Warn(err)
			return nil // We dont want to error our if the file does not exist when we start the server
		}
		return fmt.Errorf("logic: error loading rules: %s", err.Error())
	}

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&l.Rules); err != nil {
		return fmt.Errorf("logic: error loading rules: %s", err.Error())
	}

	//TODO loop over rules and generate UUIDs if needed. If it was needed save the rules again

	return nil
}

/*

new way:
{
	"name": "All off",
	"enabled": true,
	"active": true|false,
	"uuid": "e8092b86-1261-44cd-ab64-38121df58a79",
	"expression": "devices["asdf.123"].on == true && devices["asdf.123"].temperature > 20.0",
	"conditions": { // this is implemented in expression instead. rules are available there
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
