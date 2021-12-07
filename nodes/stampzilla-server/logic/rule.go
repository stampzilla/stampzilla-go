package logic

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/interpreter/functions"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/websocket"
	stypes "github.com/stampzilla/stampzilla-go/v2/pkg/types"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

type Rule struct {
	Name_         string          `json:"name"`
	Uuid_         string          `json:"uuid"`
	Active_       bool            `json:"active"`
	Pending_      bool            `json:"pending"`
	Enabled       bool            `json:"enabled"`
	Expression_   string          `json:"expression"`
	Conditions_   map[string]bool `json:"conditions"`
	Actions_      []string        `json:"actions"`
	Labels_       models.Labels   `json:"labels"`
	For_          stypes.Duration `json:"for"`
	Type_         string          `json:"type"`
	Destinations_ []string        `json:"destinations"`
	ast           *cel.Ast
	sync.RWMutex
	cancel context.CancelFunc
	stop   chan struct{}
}

func (r *Rule) Expression() string {
	r.RLock()
	defer r.RUnlock()
	return r.Expression_
}

func (r *Rule) Active() bool {
	r.RLock()
	defer r.RUnlock()
	return r.Active_
}

func (r *Rule) Pending() bool {
	r.RLock()
	defer r.RUnlock()
	return r.Pending_
}

func (r *Rule) Uuid() string {
	r.RLock()
	defer r.RUnlock()
	return r.Uuid_
}

func (r *Rule) Name() string {
	r.RLock()
	defer r.RUnlock()
	return r.Name_
}

func (r *Rule) SetUuid(uuid string) {
	r.Lock()
	r.Uuid_ = uuid
	r.Unlock()
}

func (r *Rule) Conditions() map[string]bool {
	r.RLock()
	defer r.RUnlock()
	return r.Conditions_
}

func (r *Rule) SetActive(a bool) {
	r.Lock()
	r.Active_ = a
	r.Unlock()
}

func (r *Rule) SetPending(a bool) {
	r.Lock()
	r.Pending_ = a
	r.Unlock()
}

func (r *Rule) For() stypes.Duration {
	r.RLock()
	defer r.RUnlock()
	return r.For_
}

func (r *Rule) Type() string {
	r.RLock()
	defer r.RUnlock()
	return r.Type_
}

func (r *Rule) Stop() {
	select {
	case r.stop <- struct{}{}:
	default:
	}
}

func (r *Rule) Cancel() {
	logrus.Debugf("Cancel actions on rule %s", r.Uuid())
	r.RLock()
	if r.cancel != nil {
		r.cancel()
	}
	r.RUnlock()
}

func (r *Rule) Run(store *SavedStateStore, sender websocket.Sender, triggerDestination func(string, string) error) {
	ctx, cancel := context.WithCancel(context.Background())
	r.Lock()
	r.cancel = cancel
	r.Unlock()
	defer cancel()
	for k, v := range r.Actions_ {
		duration, err := time.ParseDuration(v)
		if err == nil { // its a duration. Do the sleep only
			logrus.Debugf("logic: sleep action: %s", duration)
			select {
			case <-time.After(duration):
			case <-ctx.Done():
				logrus.Debugf("logic: stopping action %d due to cancel", k)
				return
			}
			continue
		}
		// assume its a saved state
		if ctx.Err() == context.Canceled {
			logrus.Debugf("logic: stopping action %d due to cancel", k)
			return
		}

		stateList := store.Get(v)
		if stateList == nil {
			logrus.Errorf("SavedState %s does not exist", v)
			return
		}
		devicesByNode := make(map[string]map[devices.ID]devices.State)
		for id, state := range stateList.State {
			if devicesByNode[id.Node] == nil {
				devicesByNode[id.Node] = make(map[devices.ID]devices.State)
			}
			devicesByNode[id.Node][id] = state
		}
		for nodeID, devs := range devicesByNode {
			logrus.WithFields(logrus.Fields{
				"to": nodeID,
			}).Debug("Send state change request to node")
			err := sender.SendToID(nodeID, "state-change", devs)
			if err != nil {
				logrus.Error("logic: error sending state-change to node: ", err)
				continue
			}
		}
	}

	for _, dest := range r.Destinations_ {
		logrus.Warnf("Send notification to %s", dest)
		triggerDestination(dest, r.Name())
	}
}

var celEnv *cel.Env

func init() {
	var err error
	celEnv, err = cel.NewEnv(cel.Declarations(
		decls.NewVar("devices", decls.NewMapType(decls.String, decls.Dyn)),
		decls.NewVar("rules", decls.NewMapType(decls.String, decls.Bool)),
		decls.NewFunction("daily",
			decls.NewOverload("daily_string_string",
				[]*exprpb.Type{decls.String, decls.String},
				decls.Bool)),
	))
	if err != nil {
		logrus.Fatal(err)
	}
}

// Eval evaluates the cel expression.
func (r *Rule) Eval(devices *devices.List, rules map[string]bool) (bool, error) {
	return eval(r.Expression(), devices, rules, r.ast)
}

func getDailyCelFunc() *functions.Overload {
	return &functions.Overload{
		Operator: "daily_string_string",
		Binary: func(from ref.Val, to ref.Val) ref.Val {
			now := time.Now().UTC()
			z, _ := now.Zone()
			loc, err := time.LoadLocation(z)
			if err != nil {
				return types.NewErr(err.Error())
			}
			start, err := time.ParseInLocation("15:04", from.Value().(string), loc)
			if err != nil {
				return types.NewErr(err.Error())
			}
			end, err := time.ParseInLocation("15:04", to.Value().(string), loc)
			if err != nil {
				return types.NewErr(err.Error())
			}

			today := now.Truncate(24 * time.Hour)
			n := start.Truncate(24 * time.Hour).Add(now.Sub(today))

			if inTimeSpan(start, end, n) {
				return types.Bool(true)
			}
			return types.Bool(false)
		},
	}
}

func eval(exp string, devices *devices.List, rules map[string]bool, ast *cel.Ast) (bool, error) {
	devicesState := make(map[string]map[string]interface{})
	for devID, v := range devices.All() {
		devicesState[devID.String()] = make(map[string]interface{})
		v.Lock()
		for k, v := range v.State {
			devicesState[devID.String()][k] = v
		}
		v.Unlock()
	}

	if ast == nil {
		var iss *cel.Issues
		ast, iss = celEnv.Parse(exp)
		if iss.Err() != nil {
			return false, iss.Err()
		}

		c, iss := celEnv.Check(ast)
		if iss.Err() != nil {
			return false, iss.Err()
		}
		ast = c
	}

	prg, err := celEnv.Program(ast, cel.EvalOptions(cel.OptOptimize), cel.Functions(getDailyCelFunc()))
	if err != nil {
		return false, err
	}
	result, _, err := prg.Eval(map[string]interface{}{
		"devices": devicesState,
		"rules":   rules,
	})
	if err != nil {
		return false, err
	}

	if result.Type() != types.BoolType {
		if result.Type() == types.ErrType {
			return false, result.Value().(error)
		}
		return false, ErrExpressionNotBool
	}

	return result == types.True, nil
}

var ErrExpressionNotBool = fmt.Errorf("invalid result of expression. Only bool expressions are valid")

func inTimeSpan(start, end, check time.Time) bool {
	_end := end
	_check := check
	if end.Before(start) {
		_end = end.Add(24 * time.Hour)
		if check.Before(start) {
			_check = check.Add(24 * time.Hour)
		}
	}
	return _check.After(start) && _check.Before(_end)
}

/*
func (r *Rule) RunActions(progressChan chan ActionProgress) {
	logrus.Debugf("Rule action: %s", r.Uuid())
	for _, a := range r.actions_ {
		//a.Cancel()
		a.Run(progressChan)
	}
}

// SyncActions syncronizes the action store with our actions
func (r *Rule) SyncActions(actions ActionStore) {
	r.Lock()
	r.actions_ = make([]Action, len(r.Actions_))
	for _, v := range r.Actions_ {
		r.actions_ = append(r.actions_, actions.Get(v))
	}
	r.Unlock()
}
*/

//func (r *Rule) MarshalJSON() ([]byte, error) {
//r.RLock()
//defer r.RUnlock()
//type LocalRule Rule
////TODO find a way to solve call of LocalRule copies lock value: logic.Rule (vet)
//return json.Marshal(LocalRule(*r))
//}
