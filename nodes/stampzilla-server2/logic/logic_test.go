package logic

import (
	"encoding/json"
	"testing"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/store"
	"github.com/stretchr/testify/assert"
)

func TestLoadRulesFromFile(t *testing.T) {

	store := store.New()
	l := NewLogic(store)
	l.Load("rules.json")
	//spew.Dump(l.Rules)
	jsonData, err := json.MarshalIndent(l.Rules, "", "\t")
	assert.NoError(t, err)
	t.Log(string(jsonData))
}

func TestEvaluateRules(t *testing.T) {

	store := store.New()

	l := NewLogic(store)
	r := l.AddRule("test")
	r.Expression_ = `devices["node.id"].on == true`

	l.UpdateDevice(&models.Device{
		Node: "node",
		ID:   "id",
		State: models.DeviceState{
			"on": true,
		},
	})

	l.EvaluateRules()

	assert.Equal(t, true, l.Rules[r.Uuid()].Active())

	l.UpdateDevice(&models.Device{
		Node: "node",
		ID:   "id",
		State: models.DeviceState{
			"on": false,
		},
	})
	l.EvaluateRules()
	assert.Equal(t, false, l.Rules[r.Uuid()].Active())
}
