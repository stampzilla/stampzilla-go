package notification

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEqual(t *testing.T) {
	d1 := &Destination{
		UUID:   "uuid1",
		Type:   "type1",
		Name:   "name1",
		Sender: "sender1",
		Destinations: []string{
			"a", "b",
		},
	}
	d2 := &Destination{
		UUID:   "uuid1",
		Type:   "type1",
		Name:   "name1",
		Sender: "sender1",
		Destinations: []string{
			"a", "b",
		},
	}
	assert.True(t, d1.Equal(d2))

	d3 := &Destination{
		UUID:   "uuid1",
		Type:   "type1",
		Name:   "name1",
		Sender: "sender2",
	}
	assert.False(t, d1.Equal(d3))

	d3.Sender = d1.Sender
	d3.Name = "name3"

	assert.False(t, d1.Equal(d3))

	d3.Name = d1.Name
	d3.Type = "type3"

	assert.False(t, d1.Equal(d3))

	d3.Type = d1.Type
	d3.UUID = "uuid3"

	assert.False(t, d1.Equal(d3))

	d3.UUID = d1.UUID
	d3.Destinations = []string{
		"a",
	}

	assert.False(t, d1.Equal(d3))

	d3.Destinations = []string{
		"a", "c",
	}

	assert.False(t, d1.Equal(d3))

}

func TestAddDestination(t *testing.T) {
	dests := NewDestinations()
	d1 := &Destination{
		Name: "name",
		UUID: "uuid",
	}
	dests.Add(d1)
	assert.Equal(t, "name", dests.Get("uuid").Name)
}
func TestRemoveDestination(t *testing.T) {
	dests := NewDestinations()
	d1 := &Destination{
		Name: "name",
		UUID: "uuid",
	}
	dests.Add(d1)
	dests.Remove("uuid")
	assert.Len(t, dests.All(), 0)
}

func TestReadWrite(t *testing.T) {
	file, err := ioutil.TempFile("", "readwritedestinations")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())

	ds1 := NewDestinations()
	ds2 := NewDestinations()

	d1 := &Destination{
		Name: "name",
		UUID: "uuid",
	}
	ds1.Add(d1)

	err = ds1.Save(file.Name())
	assert.NoError(t, err)

	err = ds2.Load(file.Name())
	assert.NoError(t, err)

	d2 := ds2.Get("uuid")

	assert.True(t, d1.Equal(d2))
}
