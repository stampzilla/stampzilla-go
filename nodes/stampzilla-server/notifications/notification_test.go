package notifications

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestNotificationMarshalJson(t *testing.T) {

	n := NewNotification(ErrorLevel, "test")

	b, err := json.Marshal(n)
	if err != nil {
		fmt.Println("error:", err)
	}
	t.Log(string(b))
	expected := `{"Level":"Error","Message":"test"}`
	if string(b) != expected {
		t.Errorf("Got %s expected %s", string(b), expected)
	}

}

func TestNotificationUnmarshalJson(t *testing.T) {

	n := &Notification{}

	err := json.Unmarshal([]byte(`{"Level":"Error","Message":"test"}`), &n)
	if err != nil {
		fmt.Println("error:", err)
	}
	if n.Level != ErrorLevel {
		t.Errorf("Got %s expected %s", n.Level, ErrorLevel)
	}
	if n.Message != "test" {
		t.Errorf("Got %s expected %s", n.Message, "test")
	}

}
