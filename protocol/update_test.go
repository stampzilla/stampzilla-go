package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateString(t *testing.T) {
	u := Type(TypeUpdateNode)
	assert.Equal(t, "UpdateNode", u.String())
	u = Type(TypeUpdateState)
	assert.Equal(t, "UpdateState", u.String())
	u = Type(TypeUpdateDevices)
	assert.Equal(t, "UpdateDevices", u.String())
	u = Type(TypeNotification)
	assert.Equal(t, "Notification", u.String())
	u = Type(TypePing)
	assert.Equal(t, "Ping", u.String())
	u = Type(TypePong)
	assert.Equal(t, "Pong", u.String())
	u = Type(TypeCommand)
	assert.Equal(t, "Command", u.String())
	u = Type(TypeDeviceConfigSet)
	assert.Equal(t, "DeviceConfigSet", u.String())

	u = Type(100)
	assert.Equal(t, "invalid Type", u.String())
}

func TestNewUpdateWithData(t *testing.T) {

	u := NewUpdateWithData(TypeDeviceConfigSet, 123)
	assert.Equal(t, TypeDeviceConfigSet, u.Type)
	data, err := u.ToJSON()
	assert.NoError(t, err)
	assert.Equal(t, "{\"Type\":7,\"Data\":123}", string(data))
}
