package logic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type actionStub struct {
	enterCallCount  int
	cancelCallCount int
}

func (a *actionStub) Name() string {
	return ""
}
func (a *actionStub) Uuid() string {
	return ""
}
func (a *actionStub) Run() {
	a.enterCallCount += 1
}
func (a *actionStub) Cancel() {
	a.cancelCallCount += 1
}

func TestRunCalls(t *testing.T) {

	rule := &rule{}

	enterStub := &actionStub{}
	exitStub := &actionStub{}

	rule.AddEnterAction(enterStub)
	rule.AddExitAction(exitStub)

	rule.RunEnter()

	assert.Equal(t, 1, enterStub.enterCallCount)
	assert.Equal(t, 0, enterStub.cancelCallCount)
	assert.Equal(t, 0, exitStub.enterCallCount)
	assert.Equal(t, 1, exitStub.cancelCallCount)

	rule.RunExit()

	assert.Equal(t, 1, enterStub.enterCallCount)
	assert.Equal(t, 1, enterStub.cancelCallCount)
	assert.Equal(t, 1, exitStub.enterCallCount)
	assert.Equal(t, 1, exitStub.cancelCallCount)

}
