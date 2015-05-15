package metrics

import "testing"

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

	flattened := structToMetrics(state)

	expectedKeys := map[string]string{
		"State.Sensors.2.Name":      "Sensor 2",
		"State.Sensors.2.Temp":      "22.42",
		"State.Devices.1.State.Dim": "100",
		"State.Devices.1.Features":  "",
		"State.Sensors.1.Temp":      "24",
		"State.Sensors.2.Id":        "2",
		"State.Devices.1.State.On":  "1",
		"State.Devices.1.Type":      "TYPE",
		"State.Sensors.2.Humidity":  "42",
		"State.Devices.1.Id":        "1",
		"State.Sensors.1.Id":        "1",
		"State.Sensors.1.Name":      "Sensor 1",
		"State.Sensors.1.Humidity":  "40",
		"State.Devices.1.Name":      "Sensor1",
	}

	for k, v := range expectedKeys {
		assertKeyExists(t, k, v, flattened)
	}
	//for k, v := range flattened {
	//fmt.Printf("%s = %v\n", k, v)
	//}
}

func assertKeyExists(t *testing.T, key, val string, m map[string]string) {
	if _, ok := m[key]; !ok {
		t.Errorf("Key %s does not exist", key)
	}
	if m[key] != val {
		t.Errorf("Value %s does not exist", val)
	}
}
