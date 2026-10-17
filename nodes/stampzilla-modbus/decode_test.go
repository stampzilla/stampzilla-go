package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
	data1 := []byte{0xff, 0xff} // -0.1
	data2 := []byte{0xff, 0xfa} // -0.6
	data3 := []byte{0x00, 0x3c} // 6
	data4 := []byte{0x00, 0x02} // 0.2

	assert.Equal(t, -0.1, decode(data1)/10)
	assert.Equal(t, -0.6, decode(data2)/10)
	assert.Equal(t, 6.0, decode(data3)/10)
	assert.Equal(t, 0.2, decode(data4)/10)
}

func TestEncode(t *testing.T) {
	v, err := encode(-0.1, 10)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0xff, 0xff}, []byte{byte(v >> 8), byte(v)})

	v, err = encode(6.0, 10)
	assert.NoError(t, err)
	assert.Equal(t, uint16(60), v)

	v, err = encode(true, 0)
	assert.NoError(t, err)
	assert.Equal(t, uint16(1), v)

	_, err = encode(40000.0, 0)
	assert.Error(t, err)

	_, err = encode("foo", 0)
	assert.Error(t, err)
}
