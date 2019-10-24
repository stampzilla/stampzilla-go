package file

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrigger(t *testing.T) {
	f := New(json.RawMessage("{\"append\": false, \"timestamp\": false}"))

	file, err := ioutil.TempFile("", "testTrigger")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())

	err = f.Trigger([]string{file.Name()}, "test1")
	assert.NoError(t, err)

	err = f.Trigger([]string{file.Name()}, "test2")
	assert.NoError(t, err)

	h, err := os.Open(file.Name())
	assert.NoError(t, err)
	defer h.Close()

	b, err := ioutil.ReadAll(h)

	assert.Equal(t, b, []byte("test2\tTriggered\r\n"))
}

func TestTriggerAppend(t *testing.T) {
	f := New(json.RawMessage("{\"append\": true, \"timestamp\": false}"))

	file, err := ioutil.TempFile("", "testTrigger")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())

	err = f.Trigger([]string{file.Name()}, "test1")
	assert.NoError(t, err)

	err = f.Release([]string{file.Name()}, "test2")
	assert.NoError(t, err)

	h, err := os.Open(file.Name())
	assert.NoError(t, err)
	defer h.Close()

	b, err := ioutil.ReadAll(h)

	assert.Equal(t, b, []byte("test1\tTriggered\r\ntest2\tReleased\r\n"))
}

func TestTriggerTimestamp(t *testing.T) {
	f := New(json.RawMessage("{\"append\": false, \"timestamp\": true}"))

	file, err := ioutil.TempFile("", "testTrigger")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())

	err = f.Trigger([]string{file.Name()}, "test1")
	assert.NoError(t, err)

	h, err := os.Open(file.Name())
	assert.NoError(t, err)
	defer h.Close()

	b, err := ioutil.ReadAll(h)

	assert.Equal(t, len(b), 37)
}

func TestNoFile(t *testing.T) {
	f := New(json.RawMessage(""))

	err := f.Trigger([]string{"/"}, "test1")
	assert.Error(t, err)

	err = f.Release([]string{"/"}, "test1")
	assert.Error(t, err)
}

func TestDestinations(t *testing.T) {
	f := New(json.RawMessage(""))

	d, err := f.Destinations()

	assert.Error(t, err)
	assert.Nil(t, d)
}
