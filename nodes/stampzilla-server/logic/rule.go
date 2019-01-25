package logic

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/cel-go/checker"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common"
	"github.com/google/cel-go/common/packages"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/interpreter"
	"github.com/google/cel-go/parser"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/websocket"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

type Rule struct {
	Name_       string          `json:"name"`
	Uuid_       string          `json:"uuid"`
	Operator_   string          `json:"operator"`
	Active_     bool            `json:"active"`
	Enabled     bool            `json:"enabled"`
	Expression_ string          `json:"expression"`
	Conditions_ map[string]bool `json:"conditions"`
	Actions_    []string        `json:"actions"`
	Labels_     []string        `json:"labels"`
	checkedExp  *exprpb.CheckedExpr
	sync.RWMutex
	cancel context.CancelFunc
}

func (r *Rule) Operator() string {
	r.RLock()
	defer r.RUnlock()
	return r.Operator_
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

func (r *Rule) Cancel() {
	logrus.Debugf("Cancel actions on rule %s", r.Uuid())
	r.RLock()
	if r.cancel != nil {
		r.cancel()
	}
	r.RUnlock()
}

func (r *Rule) Run(store *SavedStateStore, sender websocket.Sender) {

	//TODO implement FOR 5m that spawns a new goroutine that sleeps for 5m and then checks if the rule is still true
	// if rule is run again during this time the goroutine should be canceled/aborted
	ctx, cancel := context.WithCancel(context.Background())
	r.Lock()
	r.cancel = cancel
	r.Unlock()
	defer cancel()
	for k, v := range r.Actions_ {
		//TODO notify progress here

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

}

//Eval evaluates the cel expression
func (r *Rule) Eval(devices *devices.List, rules map[string]bool) (bool, error) {
	devicesState := make(map[string]map[string]interface{})
	for devID, v := range devices.All() {
		devicesState[devID.String()] = make(map[string]interface{})
		for k, v := range v.State {
			devicesState[devID.String()][k] = v
		}
	}

	// lazy loading improved performance like this:
	//
	// before
	// BenchmarkEval-4   	     200	   6632144 ns/op
	// after
	// BenchmarkEval-4   	  100000	     15064 ns/op

	typeProvider := types.NewProvider()
	if r.checkedExp == nil {
		// Parse the expression and returns the accumulated errors.
		src := common.NewTextSource(r.Expression())
		expr, errors := parser.Parse(src)
		if len(errors.GetErrors()) != 0 {
			return false, fmt.Errorf(errors.ToDisplayString())
		}

		env := checker.NewStandardEnv(packages.DefaultPackage, typeProvider)
		env.Add(
			decls.NewIdent("devices", decls.NewMapType(decls.String, decls.Dyn), nil))
		env.Add(
			decls.NewIdent("rules", decls.NewMapType(decls.String, decls.Bool), nil))
		c, errors := checker.Check(expr, src, env)
		if len(errors.GetErrors()) != 0 {
			return false, fmt.Errorf(errors.ToDisplayString())
		}
		r.checkedExp = c
	}

	// Interpret the checked expression using the standard overloads.
	i := interpreter.NewStandardInterpreter(packages.DefaultPackage, typeProvider)
	eval := i.NewInterpretable(interpreter.NewCheckedProgram(r.checkedExp))
	result, _ := eval.Eval(
		interpreter.NewActivation(
			map[string]interface{}{
				"devices": devicesState,
				"rules":   rules,
			}))

	if result.Type() != types.BoolType {
		if result.Type() == types.ErrType {
			return false, result.Value().(error)
		}
		return false, ErrExpressionNotBool
	}

	return result == types.True, nil
}

var ErrExpressionNotBool = fmt.Errorf("Invalid result of expression. Only bool expressions are valid")

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
