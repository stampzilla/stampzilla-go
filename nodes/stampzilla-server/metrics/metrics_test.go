package metrics

import (
	"fmt"
	"testing"
)

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

	for k, v := range flattened {
		fmt.Printf("%s = %v\n", k, v)
	}
	for k, v := range expectedKeys {
		assertKeyExists(t, k, v, flattened)
	}
}

func assertKeyExists(t *testing.T, key string, val interface{}, m map[string]interface{}) {
	if _, ok := m[key]; !ok {
		t.Errorf("Key %s does not exist", key)
	}
	if m[key] != val {
		t.Errorf("Value %s does not equal %s exist in %s", m[key], val, key)
	}
}
