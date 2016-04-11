package logic

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPauseCommandFromJson(t *testing.T) {

	mapper := NewActionsMapper()
	a := &ActionService{}
	mapper.Load(a)

	if a, ok := a.GetByUuid("actionuuid1").(*action); ok {
		if c, ok := a.Commands[1].(*command_pause); ok {
			t.Logf("%v", c.pause)
			assert.Equal(t, time.Duration(10000000), c.pause)
			return
		}
	}

	t.Error("Failed to assert duration of pause command")
}
