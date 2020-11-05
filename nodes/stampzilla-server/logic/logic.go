package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/websocket"
)

/* rules.json example:
{
    "e8092b86-1261-44cd-ab64-38121df58a79": {
        "name": "All off",
        "enabled": true,
        "active": false,
        "uuid": "e8092b86-1261-44cd-ab64-38121df58a79",
        "expression": "devices['fd230f30-6d84-4507-8ace-c1ec715be51e.1'].on == true",
        "for": "5m",
        "actions": [
            "1s",
            "c7d352bb-23f4-468c-b476-f76599c09a0d"
        ]
    },
    "1fd25327-f43c-4a00-aa67-3969dfed06b5": {
        "name": "chromecast p\u00e5",
        "enabled": true,
        "active": false,
        "uuid": "1fd25327-f43c-4a00-aa67-3969dfed06b5",
        "expression": "devices['asdf.123'].on == true ",
        "for": "5m",
        "actions": [
            "1m",
            "c7d352bb-23f4-468c-b476-f76599c09a0d"
        ],
		"labels": [
			"livingroom",
			"kitchen"
		]
    }
}
*/

// Rules is a list of rules.
type Rules map[string]*Rule

// Logic is the main struct.
type Logic struct {
	StateStore           *SavedStateStore
	Rules                map[string]*Rule
	devices              *devices.List
	onReportState        func(string, devices.State)
	onTriggerDestination func(string, string) error
	// ActionProgressChan chan ActionProgress
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

// New returns a new logic that is ready to use.
func New(sss *SavedStateStore, websocketSender websocket.Sender) *Logic {
	l := &Logic{
		devices: devices.NewList(),
		// ActionProgressChan: make(chan ActionProgress, 100),
		Rules:                make(map[string]*Rule),
		StateStore:           sss,
		onReportState:        func(string, devices.State) {},
		onTriggerDestination: func(string, string) error { return nil },
		c:                    make(chan func()),
		WebsocketSender:      websocketSender,
	}
	return l
}

// AddRule adds a new logic rule. Mainly used in tests.
func (l *Logic) AddRule(name string) *Rule {
	r := &Rule{Name_: name, Uuid_: uuid.New().String()}
	l.Lock()
	defer l.Unlock()
	l.Rules[r.Uuid()] = r
	return r
}

//GetRules returns a list of Rules.
func (l *Logic) GetRules() Rules {
	l.RLock()
	defer l.RUnlock()
	return l.Rules
}

// SetRules overwrites all rules in logic.
func (l *Logic) SetRules(rules Rules) {
	l.Lock()
	l.Rules = rules
	l.Unlock()

	// Trigger an evaluation of the new rules
	l.c <- func() {}
}

func (l *Logic) OnReportState(callback func(string, devices.State)) {
	l.onReportState = callback
}

func (l *Logic) OnTriggerDestination(callback func(string, string) error) {
	l.onTriggerDestination = callback
}

// Start starts the logic worker.
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
			l.EvaluateRules(ctx)
		case <-ctx.Done():
			logrus.Info("logic: stopping worker")
			return
		}
	}
}

// UpdateDevice update the state in the logic store with the new state from the device.
func (l *Logic) UpdateDevice(dev *devices.Device) {
	l.c <- func() {
		l.updateDevice(dev)
	}
}

func (l *Logic) updateDevice(dev *devices.Device) {
	if oldDev := l.devices.Get(dev.ID); oldDev != nil {
		if diff := oldDev.State.Diff(dev.State); len(diff) > 0 {
			oldDev.Lock()
			oldDev.State.MergeWith(diff)
			oldDev.Unlock()
		}
		return
	}
	l.devices.Add(dev)
}

// EvaluateRules loops over each rule and run evaluation on them.
func (l *Logic) EvaluateRules(ctx context.Context) {
	for _, rule := range l.Rules {
		if !rule.Enabled {
			continue
		}
		evaluation := l.evaluateRule(rule)
		if rule.For_ == 0 {
			l.runNow(rule, evaluation)
			continue
		}
		l.runFor(ctx, rule, evaluation)
	}
}

func (l *Logic) runFor(ctx context.Context, rule *Rule, evaluation bool) {
	// TODO implement for 5m here. Do not run or set active until 5m passed.
	// go run sleep 5m. cancel go routine if we have new evaluation. After 5m run evaluateRule again before rule.Run

	if rule.stop == nil { // lazy initialized
		rule.stop = make(chan struct{})
	}

	if evaluation == rule.Pending() {
		return
	}

	rule.SetPending(evaluation)

	if evaluation {
		rule.Stop()
		l.Add(1)
		go func() {
			defer l.Done()

			l.onReportState(rule.Uuid(), map[string]interface{}{
				"pending": true,
			})
			logrus.Debug("Rule: ", rule.Name(), " (", rule.Uuid(), ") - sleeping for: ", rule.For())
			select {
			case <-time.After(time.Duration(rule.For())):
			case <-ctx.Done():
				return
			case <-rule.stop:
				return
			}

			if !l.evaluateRule(rule) {
				return
			}

			logrus.Info("Rule: ", rule.Name(), " (", rule.Uuid(), ") - running actions after for: ", rule.For())

			l.onReportState(rule.Uuid(), map[string]interface{}{
				"pending": false,
				"active":  true,
			})
			rule.SetActive(true)
			rule.Run(l.StateStore, l.WebsocketSender, l.onTriggerDestination)
		}()
		return
	}

	l.onReportState(rule.Uuid(), map[string]interface{}{
		"pending": false,
		"active":  false,
	})
	rule.SetActive(false)
}

func (l *Logic) runNow(rule *Rule, evaluation bool) {
	if evaluation != rule.Active() {
		l.onReportState(rule.Uuid(), map[string]interface{}{
			"pending": false,
			"active":  evaluation,
		})
		rule.SetActive(evaluation)
		if evaluation {
			l.Add(1)
			go func() {
				logrus.Info("Rule: ", rule.Name(), " (", rule.Uuid(), ") - running actions")
				rule.Run(l.StateStore, l.WebsocketSender, l.onTriggerDestination)
				l.Done()
			}()
		} else {
			rule.Cancel()
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
		l.onReportState(r.Uuid(), map[string]interface{}{
			"error": err.Error(),
		})
		logrus.Errorf("Error evaluating rule %s: %s", r.Uuid(), err.Error())
		return false
	}
	l.onReportState(r.Uuid(), map[string]interface{}{
		"error": "",
	})
	return result
}

// Save saves the rules to rules.json.
func (l *Logic) Save() error {
	configFile, err := os.Create("rules.json")
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

//Load loads the rules from rules.json.
func (l *Logic) Load() error {
	logrus.Debug("logic: loading rules from rules.json")
	configFile, err := os.Open("rules.json")
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Warn(err)
			return nil // We dont want to error our if the file does not exist when we start the server
		}
		return fmt.Errorf("logic: error loading rules.json: %s", err.Error())
	}

	l.Lock()
	defer l.Unlock()
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&l.Rules); err != nil {
		return fmt.Errorf("logic: error loading rules.json: %s", err.Error())
	}

	// TODO loop over rules and generate UUIDs if needed. If it was needed save the rules again

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
