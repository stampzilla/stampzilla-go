package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/v2/pkg/node"
)

// MAIN - This is run when the init function is done.
func main() {
	node := node.New("modbus")
	config := NewConfig()

	modbusConnection := &Modbus{}
	writes := newWriteTracker()

	node.OnConfig(updatedConfig(config))
	node.OnRequestStateChange(onStateChange(config, modbusConnection, writes))
	wait := node.WaitForFirstConfig()

	err := node.Connect()
	if err != nil {
		logrus.Error(err)
		return
	}
	logrus.Info("Waiting for config from server")
	err = wait()
	if err != nil {
		logrus.Error(err)
		return
	}

	if config.Baud == 0 {
		config.Baud = 9600
	}

	logrus.Infof("Connecting to modbus device: %s baud: %d", config.Device, config.Baud)
	err = modbusConnection.Connect(config)

	if err != nil {
		logrus.Error("error connecting to modbus: ", err)
		return
	}

	defer modbusConnection.Close()

	results, _ := modbusConnection.ReadInputRegister(214)
	log.Println("REG_HC_TEMP_IN1: ", results)
	results, _ = modbusConnection.ReadInputRegister(215)
	log.Println("REG_HC_TEMP_IN2: ", results)
	results, _ = modbusConnection.ReadInputRegister(216)
	log.Println("REG_HC_TEMP_IN3: ", results)
	results, _ = modbusConnection.ReadInputRegister(217)
	log.Println("REG_HC_TEMP_IN4: ", results)
	results, _ = modbusConnection.ReadInputRegister(218)
	log.Println("REG_HC_TEMP_IN5: ", binary.BigEndian.Uint16(results))
	results, _ = modbusConnection.ReadInputRegister(207)
	log.Println("REG_HC_TEMP_LVL: ", results)
	results, _ = modbusConnection.ReadInputRegister(301)
	log.Println("REG_DAMPER_PWM: ", results)
	results, _ = modbusConnection.ReadInputRegister(204)
	log.Println("REG_HC_WC_SIGNAL: ", results)
	results, _ = modbusConnection.ReadInputRegister(209)
	log.Println("REG_HC_TEMP_LVL1-5: ", results)
	results, _ = modbusConnection.ReadInputRegister(101)
	log.Println("100 REG_FAN_SPEED_LEVEL: ", results)

	// connection := basenode.Connect()
	dev := &devices.Device{
		Name:   "modbusDevice",
		Type:   "sensor",
		ID:     devices.ID{ID: "1"},
		Online: true,
		Traits: []string{},
		State:  make(devices.State),
	}

	node.AddOrUpdate(dev)

	// node.SetState(registers)

	// This worker recives all incomming commands
	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-ticker.C:
			registers := config.GetRegisters()
			if len(registers) == 0 {
				logrus.Info("no configured registers to poll yet")
				continue
			}
			snapshot := writes.Snapshot()
			newState := fetchRegisters(registers, modbusConnection)
			writes.Publish(snapshot, newState, func(state devices.State) {
				node.UpdateState(dev.ID.ID, state)
			})
		case <-node.Stopped():
			ticker.Stop()
			log.Println("Stopping modbus node")
			return
		}
	}
}

func updatedConfig(config *Config) func(data json.RawMessage) error {
	return func(data json.RawMessage) error {
		config.Lock()
		defer config.Unlock()
		return json.Unmarshal(data, config)
	}
}

func onStateChange(config *Config, connection *Modbus, writes *writeTracker) func(state devices.State, device *devices.Device) error {
	type write struct {
		reg Register
		raw uint16
	}
	return func(state devices.State, device *devices.Device) error {
		registers := config.GetRegisters()

		// validate everything before writing so a rejected request does not partially change the device
		toWrite := make([]write, 0, len(state))
		names := make([]string, 0, len(state))
		for key, value := range state {
			reg := findRegister(registers, key)
			if reg == nil {
				return fmt.Errorf("no register configured with name %s", key)
			}
			if !reg.Write {
				return fmt.Errorf("register %s is not writable", key)
			}
			raw, err := encode(value, reg.Base)
			if err != nil {
				return fmt.Errorf("register %s: %w", key, err)
			}
			toWrite = append(toWrite, write{reg: *reg, raw: raw})
			names = append(names, key)
		}

		// mark before and after writing so polls started before or during the write are both invalidated
		writes.Written(names...)
		defer writes.Written(names...)
		for _, w := range toWrite {
			logrus.Infof("writing %d to register %s (%d)", w.raw, w.reg.Name, w.reg.Id)
			err := connection.WriteHoldingRegister(w.reg.Id, w.raw)
			if err != nil {
				return fmt.Errorf("error writing register %s: %w", w.reg.Name, err)
			}
		}
		return nil
	}
}

func findRegister(registers []Register, name string) *Register {
	for i := range registers {
		if registers[i].Name == name {
			return &registers[i]
		}
	}
	return nil
}

// encode converts a state value to the raw int16 register value, multiplying with base if set.
func encode(value interface{}, base int64) (uint16, error) {
	var f float64
	switch v := value.(type) {
	case float64:
		f = v
	case int:
		f = float64(v)
	case int64:
		f = float64(v)
	case bool:
		if v {
			f = 1
		}
	default:
		return 0, fmt.Errorf("unsupported value type %T", value)
	}
	if base != 0 {
		f *= float64(base)
	}
	f = math.Round(f)
	if f < math.MinInt16 || f > math.MaxInt16 {
		return 0, fmt.Errorf("value %v out of range", f)
	}
	return uint16(int16(f)), nil
}

func fetchRegisters(registers []Register, connection *Modbus) devices.State {
	state := make(devices.State)
	for _, v := range registers {
		var data []byte
		var err error
		if v.Write {
			data, err = connection.ReadHoldingRegister(v.Id)
		} else {
			data, err = connection.ReadInputRegister(v.Id)
		}
		if err != nil {
			/*
				if connection.handler.Logger == nil {
				log.Println("Adding debug logging to handler")
				connection.handler.Logger = log.New(os.Stdout, "modbus-debug: ", log.LstdFlags)
				}
			*/
			logrus.Error(err)
			continue
		}
		if len(data) != 2 {
			logrus.Error("Wrong length, expected 2")
			continue
		}
		if v.Base != 0 {
			state[v.Name] = decode(data) / float64(v.Base)
			continue
		}
		state[v.Name] = decode(data)
	}
	return state
}

func decode(data []byte) float64 {
	var i int16
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.BigEndian, &i)
	return float64(i)
}
