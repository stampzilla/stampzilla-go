package devices

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeviceStateBool(t *testing.T) {

	ds := make(DeviceState)
	ds["true"] = true
	ds["false"] = false

	assert.Equal(t, true, ds.Bool("true"))
	assert.Equal(t, false, ds.Bool("false"))
	assert.Equal(t, false, ds.Bool("on"))

}
