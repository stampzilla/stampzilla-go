package notification

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateFileSender(t *testing.T) {
	s := &Sender{
		Type: "file",
	}

	d1 := &Destination{
		Name: "name",
		UUID: "uuid",
	}

	err := s.Trigger(d1, "test")
	assert.NoError(t, err)

	err = s.Release(d1, "test")
	assert.NoError(t, err)

	d, err := s.Destinations()
	assert.Error(t, err)
	assert.Nil(t, d)
}

func TestCreateUnknownSender(t *testing.T) {
	s := &Sender{
		Type: "unknown",
	}

	d1 := &Destination{
		Name: "name",
		UUID: "uuid",
	}

	err := s.Trigger(d1, "test")
	assert.Error(t, err)

	err = s.Release(d1, "test")
	assert.Error(t, err)

	d, err := s.Destinations()
	assert.Error(t, err)
	assert.Nil(t, d)
}

func TestReadWriteSenders(t *testing.T) {
	file, err := ioutil.TempFile("", "readwritesenders")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())

	s1 := NewSenders()
	s2 := NewSenders()

	sender1 := Sender{
		Name:       "name1",
		UUID:       "uuid1",
		Type:       "type1",
		Parameters: json.RawMessage("null"),
	}
	s1.Add(sender1)

	err = s1.Save(file.Name())
	assert.NoError(t, err)

	err = s2.Load(file.Name())
	assert.NoError(t, err)

	sender2, ok := s2.Get("uuid1")
	assert.True(t, ok)
	assert.Equal(t, sender1, sender2)

	assert.Len(t, s2.All(), 1)

	s2.Remove("uuid1")
	assert.Len(t, s2.All(), 0)
}
