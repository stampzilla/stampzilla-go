package types

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMarshalJSOn(t *testing.T) {
	s := struct {
		D Duration
	}{
		D: Duration(time.Second),
	}

	j, err := json.Marshal(&s)
	assert.NoError(t, err)
	assert.Equal(t, `{"D":"1s"}`, string(j))
}

func TestUnmarshalJSON(t *testing.T) {
	s := struct {
		D Duration
	}{}

	err := json.Unmarshal([]byte(`{"D":"1s"}`), &s)
	assert.NoError(t, err)
	assert.Equal(t, Duration(time.Second), s.D)

	err = json.Unmarshal([]byte(`{"D":"200ms"}`), &s)
	assert.NoError(t, err)
	assert.Equal(t, Duration(time.Millisecond*200), s.D)
}
