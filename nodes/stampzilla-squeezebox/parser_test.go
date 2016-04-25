package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParserPower(t *testing.T) {
	p := &parser{}
	assert.Equal(t, true, p.Power("power 1 1"))
}
