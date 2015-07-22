package metrics

import (
	"fmt"
	"os"
	"testing"
	"time"

	log "github.com/cihub/seelog"
	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
)

func TestMain(m *testing.M) {
	testConfig := `
			<seelog type="sync" asyncinterval="1000" minlevel="warn">
				<outputs>
					<filter levels="trace">
						<console formatid="colored-trace"/>
					</filter>
					<filter levels="debug">
						<console formatid="colored-debug"/>
					</filter>
					<filter levels="info">
						<console formatid="colored-info"/>
					</filter>
					<filter levels="warn">
						<console formatid="colored-warn"/>
					</filter>
					<filter levels="error">
						<console formatid="colored-error"/>
					</filter>
					<filter levels="critical">
						<console formatid="colored-critical"/>
					</filter>
				</outputs>
				<formats>
					<format id="colored-trace"  format="%Date %Time %EscM(40)%Level%EscM(49) - %File:%Line - %Msg%n%EscM(0)"/>
					<format id="colored-debug"  format="%Date %Time %EscM(45)%Level%EscM(49) - %File:%Line - %Msg%n%EscM(0)"/>
					<format id="colored-info"  format="%Date %Time %EscM(46)%Level%EscM(49) - %File:%Line - %Msg%n%EscM(0)"/>
					<format id="colored-warn"  format="%Date %Time %EscM(43)%Level%EscM(49) - %File:%Line - %Msg%n%EscM(0)"/>
					<format id="colored-error"  format="%Date %Time %EscM(41)%Level%EscM(49) - %File:%Line - %Msg%n%EscM(0)"/>
					<format id="colored-critical"  format="%Date %Time %EscM(41)%Level%EscM(49) - %File:%Line - %Msg%n%EscM(0)"/>
				</formats>
			</seelog>`
	logger, _ := log.LoggerFromConfigAsBytes([]byte(testConfig))
	log.ReplaceLogger(logger)

	os.Exit(m.Run())
}

func TestMap(t *testing.T) {

	type Sensor struct {
		Id       int
		Name     string
		Temp     float64
		Humidity float64
	}
	type DeviceState struct {
		On  bool
		Dim int
	}
	type Device struct {
		Id       string
		Name     string
		State    DeviceState
		Type     string
		Features []string
	}
	type State struct {
		Devices map[string]*Device
		Sensors map[string]*Sensor
	}

	sensor1 := &Sensor{
		1,
		"Sensor 1",
		24,
		40,
	}
	sensor2 := &Sensor{
		2,
		"Sensor 2",
		22.42,
		42,
	}

	device1 := &Device{
		"1",
		"Sensor1",
		DeviceState{true, 100},
		"TYPE",
		nil,
	}

	state := &State{
		Sensors: make(map[string]*Sensor),
		Devices: make(map[string]*Device),
	}

	state.Sensors["1"] = sensor1
	state.Sensors["2"] = sensor2
	state.Devices["1"] = device1

	flattened := structToMetrics("", state)

	expectedKeys := map[string]interface{}{
		"Sensors_2_Name":      "Sensor 2",
		"Sensors_2_Temp":      22.42,
		"Devices_1_State_Dim": 100,
		"Devices_1_Features":  "",
		"Sensors_1_Temp":      24.0,
		"Sensors_2_Id":        2,
		"Devices_1_State_On":  1,
		"Devices_1_Type":      "TYPE",
		"Sensors_2_Humidity":  42.0,
		"Devices_1_Id":        1,
		"Sensors_1_Id":        1,
		"Sensors_1_Name":      "Sensor 1",
		"Sensors_1_Humidity":  40.0,
		"Devices_1_Name":      "Sensor1",
	}

	t.Logf("Flatten result:")
	for k, v := range flattened {
		t.Logf("%s = %v\n", k, v)
	}
	t.Logf("\n")

	for k, v := range expectedKeys {
		assertKeyExists(t, k, v, flattened)
	}
}

func assertKeyExists(t *testing.T, key string, val interface{}, m map[string]interface{}) {
	if _, ok := m[key]; !ok {
		t.Errorf("Key %s does not exist", key)
		return
	}
	if m[key] != val {
		t.Errorf("Value %s does not equal expected value %s, exist in %s", m[key], val, key)
	}
}

func assertSliceHas(t *testing.T, val string, s []string) {
	for _, v := range s {
		if v == val {
			return
		}
	}
	t.Errorf("%s does not exist in slice", val)
}

type Sensor struct {
	Id       int
	Name     string
	Temp     float64
	Humidity float64
}
type DeviceState struct {
	On  bool
	Dim int
}
type Device struct {
	Id       string
	Name     string
	State    DeviceState
	Type     string
	Features []string
}
type State struct {
	Devices map[string]*Device
	Sensors map[string]*Sensor
}

func getState() State {
	sensor1 := &Sensor{
		1,
		"Sensor 1",
		24,
		40,
	}
	sensor2 := &Sensor{
		2,
		"Sensor 2",
		22.42,
		42,
	}

	device1 := &Device{
		"1",
		"Sensor1",
		DeviceState{true, 100},
		"TYPE",
		nil,
	}

	state := State{
		Sensors: make(map[string]*Sensor),
		Devices: make(map[string]*Device),
	}

	state.Sensors["1"] = sensor1
	state.Sensors["2"] = sensor2
	state.Devices["1"] = device1
	return state
}

func TestUpdate(t *testing.T) {

	m := New()
	l := &LoggerStub{T: t, lastValues: make(map[string]interface{})}
	m.AddLogger(l)
	m.Start()

	node := &serverprotocol.Node{}
	state := getState()

	node.Name = "metrics-test"
	node.Uuid = "123-123"
	node.SetState(state)

	m.Update(node)

	state.Sensors["1"].Temp = 100
	m.Update(node)

	state.Sensors["2"].Temp = 100
	m.Update(node)

	// Adding an new sensor
	sensor3 := &Sensor{
		3,
		"Sensor 3",
		78,
		42,
	}
	state.Sensors["3"] = sensor3
	m.Update(node)

	// Test to delete that same sensor again
	delete(state.Sensors, "3")
	m.Update(node)

	//Wait for all metric.Log calls to finnish
	time.Sleep(100 * time.Millisecond)

	if l.logcount != 20 {
		t.Errorf("Expected Log to have ran 20 time got: %d", l.logcount)
	}
	if l.commitcount != 4 {
		t.Errorf("Expected Commit to have ran 4 time got: %d", l.commitcount)
	}
	expectedKeys := map[string]interface{}{
		"123-123_Node_State_Sensors_2_Name":      "Sensor 2",
		"123-123_Node_State_Sensors_2_Temp":      100.0,
		"123-123_Node_State_Devices_1_State_Dim": 100,
		"123-123_Node_State_Devices_1_Features":  "",
		"123-123_Node_State_Sensors_1_Temp":      100.0,
		"123-123_Node_State_Sensors_2_Id":        2,
		"123-123_Node_State_Devices_1_State_On":  1,
		"123-123_Node_State_Devices_1_Type":      "TYPE",
		"123-123_Node_State_Sensors_2_Humidity":  42.0,
		"123-123_Node_State_Devices_1_Id":        1,
		"123-123_Node_State_Sensors_1_Id":        1,
		"123-123_Node_State_Sensors_1_Name":      "Sensor 1",
		"123-123_Node_State_Sensors_1_Humidity":  40.0,
		"123-123_Node_State_Devices_1_Name":      "Sensor1",
		"123-123_Node_State_Sensors_3_Name":      "Sensor 3",
		"123-123_Node_State_Sensors_3_Humidity":  42.0,
	}

	for k, v := range expectedKeys {
		assertKeyExists(t, k, v, l.lastValues)
	}
}

func TestUpdateSameValueVeryFast(t *testing.T) {

	m := New()
	l := &LoggerStub{T: t, lastValues: make(map[string]interface{})}
	m.AddLogger(l)
	m.Start()

	node := &serverprotocol.Node{}
	state := getState()

	node.Name = "metrics-test"
	node.Uuid = "123-123"
	node.SetState(state)

	m.Update(node)

	state.Sensors["1"].Temp = 100
	m.Update(node)

	state.Sensors["1"].Temp = 101
	m.Update(node)

	//Wait for all metric.Log calls to finnish
	time.Sleep(100 * time.Millisecond)

	expectedKeys := []string{
		"123-123_Node_State_Sensors_1_Temp 24",
		"123-123_Node_State_Sensors_1_Temp 100",
		"123-123_Node_State_Sensors_1_Temp 101",
	}

	for _, v := range expectedKeys {
		assertSliceHas(t, v, l.logged)
	}
}

type LoggerStub struct {
	logcount    int
	commitcount int
	T           *testing.T
	lastValues  map[string]interface{}
	logged      []string
}

func (m *LoggerStub) Log(key string, value interface{}) {
	m.T.Log("Log: ", key, value)
	m.logcount++
	m.lastValues[key] = value
	m.logged = append(m.logged, fmt.Sprint(key, " ", value))
}
func (m *LoggerStub) Commit(node interface{}) {
	m.T.Log("Commit")
	//log.Info(node)
	m.commitcount++
}
