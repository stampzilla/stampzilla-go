package logic

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadRulesFromFile(t *testing.T) {

	l := NewLogic()
	l.LoadRulesFromFile("rules.json")
	//spew.Dump(l.Rules)
	jsonData, err := json.MarshalIndent(l.Rules, "", "\t")
	assert.NoError(t, err)
	t.Log(string(jsonData))
}
