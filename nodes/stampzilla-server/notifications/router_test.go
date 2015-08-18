package notifications

import (
	"reflect"
	"sync"
	"testing"
	"time"
)

type testTransport struct {
	T       *testing.T
	history []Notification
	wg      sync.WaitGroup
}

func (self *testTransport) Start() {
}
func (self *testTransport) Dispatch(note Notification) {
	self.history = append(self.history, note)
	self.T.Log("Dispatcher recived: ", note)

	self.wg.Done()
}
func (self *testTransport) Stop() {
}

func (self *testTransport) haveRecived(msg Notification, target int) {
	found := 0

	for _, msg2 := range self.history {
		if reflect.DeepEqual(msg, msg2) {
			found++
		}
	}

	if found != target {
		self.T.Error("Failed to recive notification (", msg.Level, " - ", msg.Message, "), recived ", found, " and target was ", target)
	}
}

func TestDispatch(t *testing.T) {
	router := NewRouter()

	transport := testTransport{T: t}
	transport2 := testTransport{T: t}

	// Add the dummy transport and request only Warnings
	router.AddTransport(&transport, []string{"Warning"})
	router.AddTransport(&transport2, []string{"Error"})

	// Test send 300 notifications
	transport.wg.Add(150)
	transport2.wg.Add(150)
	for i := 0; i < 150; i++ {
		router.Dispatch(Notification{
			Source:     "RouterTest",
			SourceUuid: "123-123",
			Level:      NewNotificationLevel("Warning"),
			Message:    "Test message",
		})

		router.Dispatch(Notification{
			Source:     "RouterTest",
			SourceUuid: "123-123",
			Level:      NewNotificationLevel("Error"),
			Message:    "Test message",
		})
	}

	WaitOrTimeout(t, &transport.wg, time.Second)

	// Check that we have recived the correct amount of notifications
	transport.haveRecived(Notification{
		Source:     "RouterTest",
		SourceUuid: "123-123",
		Level:      NewNotificationLevel("Warning"),
		Message:    "Test message",
	}, 150)

	// And check that we didnt recive notifications that we dont requested
	transport.haveRecived(Notification{
		Source:     "RouterTest",
		SourceUuid: "123-123",
		Level:      NewNotificationLevel("Error"),
		Message:    "Test message",
	}, 0)

	// Check that we have recived the correct amount of notifications
	transport2.haveRecived(Notification{
		Source:     "RouterTest",
		SourceUuid: "123-123",
		Level:      NewNotificationLevel("Warning"),
		Message:    "Test message",
	}, 0)

	// And check that we didnt recive notifications that we dont requested
	transport2.haveRecived(Notification{
		Source:     "RouterTest",
		SourceUuid: "123-123",
		Level:      NewNotificationLevel("Error"),
		Message:    "Test message",
	}, 150)
}

func WaitOrTimeout(t *testing.T, wg *sync.WaitGroup, timeout time.Duration) {
	//Wait for all metric.Log calls to finish
	done := make(chan bool)
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(timeout):
		t.Errorf("TIMEOUT, not all metrics.Update calls finished in time")
	}
}
