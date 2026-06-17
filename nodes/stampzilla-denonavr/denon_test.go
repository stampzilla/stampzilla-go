package main

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stampzilla/stampzilla-go/v2/pkg/node"
)

func TestVolumeToMV(t *testing.T) {
	tests := []struct {
		v    float64
		want string
	}{
		{0.0, "MV00"},
		{1.0, "MV80"},
		{0.5, "MV40"},
		{0.38125, "MV305"}, // 80 * 0.38125 = 30.5 → half step
		{0.75625, "MV605"}, // 80 * 0.75625 = 60.5 → half step
		{-0.1, "MV00"},     // clamped below 0
		{1.5, "MV80"},      // clamped above 1
	}
	for _, tt := range tests {
		got := volumeToMV(tt.v)
		if got != tt.want {
			t.Errorf("volumeToMV(%v) = %q, want %q", tt.v, got, tt.want)
		}
	}
}

func TestMVToVolume(t *testing.T) {
	tests := []struct {
		param string
		want  float64
	}{
		{"00", 0.0},
		{"80", 1.0},
		{"40", 0.5},
		{"805", 1.0},     // 80.5/80 clamped to 1.0
		{"395", 0.49375}, // 39.5/80
		{"305", 0.38125}, // 30.5/80
	}
	for _, tt := range tests {
		got := mvToVolume(tt.param)
		if got != tt.want {
			t.Errorf("mvToVolume(%q) = %v, want %v", tt.param, got, tt.want)
		}
	}

	// invalid inputs
	for _, bad := range []string{"", "x0", "abc", "8", "1234"} {
		if v := mvToVolume(bad); v != -1 {
			t.Errorf("mvToVolume(%q) = %v, want -1", bad, v)
		}
	}
}

func TestParseEvent(t *testing.T) {
	tests := []struct {
		line  string
		key   string
		value any
	}{
		{"PWON", "on", true},
		{"PWSTANDBY", "on", false},
		{"MV80", "volume", 1.0},
		{"MV00", "volume", 0.0},
		{"MV40", "volume", 0.5},
		{"MV805", "volume", 1.0}, // 80.5/80 clamped
		{"SIDVD", "source", "DVD"},
		{"SIMPLAY", "source", "MPLAY"},
		{"SITV", "source", "TV"},
		{"SICD", "source", "CD"},
	}
	for _, tt := range tests {
		state := parseEvent(tt.line)
		if state == nil {
			t.Errorf("parseEvent(%q) returned nil, want state with %q=%v", tt.line, tt.key, tt.value)
			continue
		}
		got, ok := state[tt.key]
		if !ok {
			t.Errorf("parseEvent(%q): missing key %q", tt.line, tt.key)
			continue
		}
		if got != tt.value {
			t.Errorf("parseEvent(%q)[%q] = %v (%T), want %v (%T)", tt.line, tt.key, got, got, tt.value, tt.value)
		}
	}

	// These lines should produce no state.
	ignored := []string{
		"MVMAX 98",
		"MVMAX98",
		"MS STEREO",
		"DC AUTO",
		"",
		"PWMUTING",
	}
	for _, line := range ignored {
		if state := parseEvent(line); len(state) != 0 {
			t.Errorf("parseEvent(%q) should return nil/empty, got %v", line, state)
		}
	}
}

type mockTCPServer struct {
	listener net.Listener
	conns    chan net.Conn
	stop     chan struct{}
	wg       sync.WaitGroup
}

func newMockTCPServer(t *testing.T) *mockTCPServer {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %s", err)
	}
	s := &mockTCPServer{
		listener: l,
		conns:    make(chan net.Conn, 10),
		stop:     make(chan struct{}),
	}
	s.wg.Add(1)
	go s.serve()
	return s
}

func (s *mockTCPServer) serve() {
	defer s.wg.Done()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.stop:
				return
			default:
				return
			}
		}
		s.conns <- conn
		// Spawn a goroutine to read and discard commands so the write buffer doesn't fill up.
		go func(c net.Conn) {
			buf := make([]byte, 1024)
			for {
				_, err := c.Read(buf)
				if err != nil {
					return
				}
			}
		}(conn)
	}
}

func (s *mockTCPServer) addr() string {
	return s.listener.Addr().String()
}

func (s *mockTCPServer) close() {
	close(s.stop)
	s.listener.Close()
	s.wg.Wait()
	// Close any accepted connections
	close(s.conns)
	for conn := range s.conns {
		conn.Close()
	}
}

func TestManager(t *testing.T) {
	n := node.NewWithClient(nil)

	// Save original functions and restore afterwards
	origEnsureDevice := ensureDevice
	origSetDeviceOnline := setDeviceOnline
	defer func() {
		ensureDevice = origEnsureDevice
		setDeviceOnline = origSetDeviceOnline
	}()

	// Mock ensureDevice and setDeviceOnline to prevent calling actual node methods
	// that would block on the unexported sendUpdate channel. We can use these mocks
	// to assert that devices are registered and their online statuses are tracked.
	var mu sync.Mutex
	ensuredDevices := make(map[string]string)
	onlineChanges := make(chan string, 10)

	ensureDevice = func(node *node.Node, id string, name string) {
		mu.Lock()
		defer mu.Unlock()
		ensuredDevices[id] = name
	}

	setDeviceOnline = func(node *node.Node, id string, online bool) {
		onlineChanges <- fmt.Sprintf("%s:%v", id, online)
	}

	// Spin up real TCP servers on random localhost ports.
	srv1 := newMockTCPServer(t)
	defer srv1.close()
	srv2 := newMockTCPServer(t)
	defer srv2.close()

	mgr := newManager(n)

	// Build the config JSON using the dynamic addresses of our real TCP servers.
	configJSON := []byte(fmt.Sprintf(`{
		"devices": {
			"1": {
				"host": "%s",
				"name": "Living Room"
			},
			"2": {
				"host": "%s",
				"name": "Bedroom"
			}
		}
	}`, srv1.addr(), srv2.addr()))

	err := mgr.updatedConfig(configJSON)
	if err != nil {
		t.Fatalf("updatedConfig failed: %s", err)
	}

	// Verify workers are created.
	mgr.mu.Lock()
	workersCount := len(mgr.workers)
	mgr.mu.Unlock()
	if workersCount != 2 {
		t.Errorf("expected 2 workers, got %d", workersCount)
	}

	// We should expect both devices to go online as our connections succeed.
	onlineTargets := map[string]bool{"1:true": false, "2:true": false}
	timeout := time.After(2 * time.Second)
	for len(onlineTargets) > 0 {
		select {
		case change := <-onlineChanges:
			if _, ok := onlineTargets[change]; ok {
				delete(onlineTargets, change)
			}
		case <-timeout:
			t.Fatalf("timed out waiting for devices to go online. missing: %v", onlineTargets)
		}
	}

	mu.Lock()
	name1 := ensuredDevices["1"]
	name2 := ensuredDevices["2"]
	mu.Unlock()

	if name1 != "Living Room" {
		t.Errorf("expected device 1 name to be Living Room, got %s", name1)
	}
	if name2 != "Bedroom" {
		t.Errorf("expected device 2 name to be Bedroom, got %s", name2)
	}

	// Now update config, remove device 2, and change host of device 1 to a new TCP server.
	srv3 := newMockTCPServer(t)
	defer srv3.close()

	configJSON2 := []byte(fmt.Sprintf(`{
		"devices": {
			"1": {
				"host": "%s",
				"name": "Living Room New"
			}
		}
	}`, srv3.addr()))

	err = mgr.updatedConfig(configJSON2)
	if err != nil {
		t.Fatalf("updatedConfig 2 failed: %s", err)
	}

	mgr.mu.Lock()
	workersCount = len(mgr.workers)
	mgr.mu.Unlock()
	if workersCount != 1 {
		t.Errorf("expected 1 worker, got %d", workersCount)
	}

	// We expect device 2 to go offline, and device 1 to go offline first then online on srv3.
	onlineTargets2 := map[string]bool{"2:false": false, "1:true": false}
	timeout2 := time.After(2 * time.Second)
	for len(onlineTargets2) > 0 {
		select {
		case change := <-onlineChanges:
			if _, ok := onlineTargets2[change]; ok {
				delete(onlineTargets2, change)
			}
		case <-timeout2:
			t.Fatalf("timed out waiting for second config changes. missing: %v", onlineTargets2)
		}
	}

	mu.Lock()
	name1New := ensuredDevices["1"]
	mu.Unlock()

	if name1New != "Living Room New" {
		t.Errorf("expected updated name for device 1 to be Living Room New, got %s", name1New)
	}

	// Clean up manager.
	mgr.close()
}
