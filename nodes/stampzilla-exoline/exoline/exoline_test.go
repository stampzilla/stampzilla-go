package exoline

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFloat64bytes(t *testing.T) {
	f := 12.34
	b := FloatTobytes(f)
	assert.Equal(t, []byte{164, 112, 69, 65}, b)
}

func TestAsRoundedFloat(t *testing.T) {
	b := []byte{164, 112, 69, 65}
	f, err := AsRoundedFloat(b)
	assert.NoError(t, err)
	assert.Equal(t, 12.34, f)
}
