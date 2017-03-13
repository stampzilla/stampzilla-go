package main

import (
	"testing"
)

func TestMaxId(t *testing.T) {

	list := deviceList{
		&Device{
			Id: 5,
		},
		&Device{
			Id: 2,
		},
		&Device{
			Id: 4,
		},
	}

	max := list.maxId()

	if max != 5 {
		t.Error("Expected max id to be 5")
	}
	t.Log("max: ", max)
}
