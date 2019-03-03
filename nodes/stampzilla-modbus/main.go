package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"log"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/pkg/node"
)

// MAIN - This is run when the init function is done
func main() {
	node := node.New("modbus")
	config := NewConfig()

	node.OnConfig(updatedConfig(config))
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

	modbusConnection := &Modbus{}
	logrus.Infof("Connecting to modbus device: %s", config.Device)
	err = modbusConnection.Connect()

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

	//connection := basenode.Connect()
	dev := &devices.Device{
		Name:   "modbusDevice",
		Type:   "sensor", //TODO if we add modbus write support we need to have another type
		ID:     devices.ID{ID: "1"},
		Online: true,
		Traits: []string{},
		State:  make(devices.State),
	}

	node.AddOrUpdate(dev)

	//node.SetState(registers)

	// This worker recives all incomming commands
	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-ticker.C:
			if len(config.Registers) == 0 {
				logrus.Info("no configured registers to poll yet")
				continue
			}
			fetchRegisters(config.Registers, modbusConnection)

			newState := make(devices.State)
			for _, v := range config.Registers {
				newState[v.Name] = v.Value
			}
			node.UpdateState(dev.ID.ID, newState)
		case <-node.Stopped():
			ticker.Stop()
			log.Println("Stopping modbus node")
			return
		}
	}
}

func updatedConfig(config *Config) func(data json.RawMessage) error {
	return func(data json.RawMessage) error {
		return json.Unmarshal(data, config)
	}
}

func fetchRegisters(registers Registers, connection *Modbus) {
	for _, v := range registers {

		data, err := connection.ReadInputRegister(v.Id)
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
			v.Value = decode(data) / float64(v.Base)
			continue
		}
		v.Value = decode(data)
	}
}

func decode(data []byte) float64 {
	var i int16
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.BigEndian, &i)
	return float64(i)
}
