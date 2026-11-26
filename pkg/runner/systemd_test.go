package runner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJobResult(t *testing.T) {
	assert.NoError(t, jobResult("stampzilla-tibber.service", "done"))
	for _, r := range []string{"failed", "timeout", "canceled", "dependency", "skipped"} {
		err := jobResult("stampzilla-tibber.service", r)
		assert.EqualError(t, err, "stampzilla-tibber.service: systemd job "+r)
	}
}

func TestCheckRunning(t *testing.T) {
	u := "stampzilla-tibber.service"
	assert.NoError(t, checkRunning(u, 100, unitState{"active", "running", 100}))

	assert.EqualError(t, checkRunning(u, 100, unitState{"activating", "auto-restart", 0}),
		u+" is activating (auto-restart) after restart")
	assert.EqualError(t, checkRunning(u, 100, unitState{"failed", "failed", 0}),
		u+" is failed (failed) after restart")
	assert.EqualError(t, checkRunning(u, 100, unitState{"active", "running", 200}),
		u+" crashed and was restarted by systemd (pid 100 -> 200)")
}
