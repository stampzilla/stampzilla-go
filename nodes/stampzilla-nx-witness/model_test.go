package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStampzillaDeviceID(t *testing.T) {
	r := Rule{
		Comment: "test stampzilla:123a",
	}
	assert.Equal(t, "123a", r.StampzillaDeviceID())
}
