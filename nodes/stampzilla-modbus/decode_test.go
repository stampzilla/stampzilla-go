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
