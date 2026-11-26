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
