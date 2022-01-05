package onewire

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

var ErrSensorRead = errors.New("failed to read temperature from sensor")

func SensorsWithTemperature() ([]string, error) {
	data, err := ioutil.ReadFile("/sys/bus/w1/devices/w1_bus_master1/w1_master_slaves")
	if err != nil {
		return nil, fmt.Errorf("error reading 1wire from sys filesystem: %w", err)
	}
	sensors := []string{}
	for _, s := range strings.Split(string(data), "\n") {
		s = strings.TrimSpace(s)
		if _, err := Temperature(s); err != nil {
			continue
		}
		sensors = append(sensors, s)
	}
	return sensors, nil
}

// Temperature get the temp of sensor with id.
func Temperature(sensor string) (float64, error) {
	data, err := ioutil.ReadFile("/sys/bus/w1/devices/" + sensor + "/w1_slave")
	if err != nil {
		return 0.0, ErrSensorRead
	}

	raw := string(data)

	if !strings.Contains(raw, " YES") {
		return 0.0, ErrSensorRead
	}

	i := strings.LastIndex(raw, "t=")
	if i == -1 {
		return 0.0, ErrSensorRead
	}

	c, err := strconv.ParseFloat(raw[i+2:len(raw)-1], 64)
	if err != nil {
		return 0.0, ErrSensorRead
	}

	return c / 1000.0, nil
}
