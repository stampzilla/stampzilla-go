//go:generate bash -c "go get -u github.com/rakyll/statik && ~/go/bin/statik -src ./nibe -f -include=*.json"

package main

import (
	"log"
	"strconv"

	"github.com/rakyll/statik/fs"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-nibe/nibe"
	_ "github.com/stampzilla/stampzilla-go/nodes/stampzilla-nibe/statik"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
	"github.com/stampzilla/stampzilla-go/pkg/node"
)

var config = &Config{}

var registers = map[string]map[int]string{
	"Common": map[int]string{
		45001: "Alarm code",                    // Alarm
		47137: "Mode",                          // Operational mode
		47138: "Medium mode",                   // Operational mode heat medium pump
		40004: "Outdoor temperature",           // BT1 Outdoor Temperature
		40067: "Outdoor temperature (average)", // BT1 Outdoor Temperature
		40033: "Indoor temperature",            // BT50 Room Temp S1
	},
	"Degree minutes": map[int]string{
		40940: "Current",           // Degree Minutes (32 bit)
		47206: "Heating threshold", // gradminuter start heating
	},
	"Climat system 1": map[int]string{
		40008: "Supply water temperature",  // BT2 Supply temp S1
		40012: "Return water temperature",  // EB100-EP14-BT3 return temp
		40047: "Supply water2 temperature", // EB100-BT61 Supply Radiator temp
		40048: "Return water2 temperature", // EB100-BT62 Return Radiator temp
	},
	"Heatwater": map[int]string{
		40013: "Tank top temperature",   // BT7 HW Top
		40014: "Tank inlet temperature", // BT6 HW Load
	},
	"Ventilation": map[int]string{
		40050: "Air flow",                  // EB100-BS1 Air flow
		40025: "Exhaust air temperature",   // BT20 Exhaust air temp. 1
		40026: "Extracted air temperature", // BT21 Vented air temp. 1
		48206: "Silent mode",               // Silent Mode Status
		43108: "Fanspeed",
		41256: "Fanspeed 1",
		41257: "Fanspeed 2",
		41258: "Fanspeed 3",
	},
	"Compressor": map[int]string{
		43136: "Current frequenzy", // Compressor Frequency, Actual
		43420: "Production hours",  // Tot. op.time compr. EB100-EP14
		43424: "Hotwater hours",    // Tot. HotW op.time compr. EB100-EP14
		43416: "Starts",            // Compressor starts EB100-EP14
	},
	"Additive": map[int]string{
		43084: "Current power",    // Current add. Power
		43081: "Production hours", // Tot. op.time add.
		43239: "Hotwater hours",   // Tot. HotW op.time add
	},
	"Sam40": map[int]string{
		42093: "Current speed", // sam GQ3 speed
		40141: "Supply air",    // AZ2-BT22 Supply air temp
		40142: "Outdoor air",   // AZ2-BT23 Outdoor air temp
		40143: "Supply water",  // AZ2-BT68 Flow water temp
		40144: "Return water",  // AZ2-BT69 Return water temp
	},
	"Adjust+": map[int]string{
		40877: "Active",                 // adjust + activated
		40874: "Indoor temperature",     // indoor temperature
		40872: "Adjustment",             // parallel adjustment
		40878: "Requests heat",          // requests heat
		40879: "Factor",                 // parallel factor
		40880: "Max allowed adjustment", // max change allowed
	},
}

func main() {
	statikFS, err := fs.New()
	if err != nil {
		logrus.Fatal(err)
	}
	nibe := nibe.New()
	err = nibe.LoadDefinitions(statikFS, "/f750.json")
	if err != nil {
		logrus.Fatalf("failed to load: %s", err.Error())
	}

	node := node.New("nibe")

	node.OnConfig(updatedConfig(nibe))

	err = node.Connect()
	if err != nil {
		logrus.Error(err)
		return
	}

	id := 0
	list := make(map[int]*devices.Device)
	name := make(map[int]string)
	for n, r := range registers {
		d := &devices.Device{
			Name:   n,
			Type:   "sensor",
			ID:     devices.ID{ID: strconv.Itoa(id)},
			Online: false,
			Traits: []string{},
			State:  devices.State{},
		}

		for id, desc := range r {
			if id == 45001 {
				continue
			}
			d.State[desc] = 0.0
			list[id] = d
			name[id] = desc
		}

		id++
		node.AddOrUpdate(d)

	}

	other := &devices.Device{
		Name:   "Other",
		Type:   "sensor",
		ID:     devices.ID{ID: strconv.Itoa(id)},
		Online: false,
		Traits: []string{},
		State:  devices.State{},
	}
	id++

	alarm := &devices.Device{
		Name:   "Alarm",
		Type:   "",
		ID:     devices.ID{ID: strconv.Itoa(id)},
		Online: false,
		Traits: []string{
			"OnOff",
		},
		State: devices.State{
			"on":         false,
			"Alarm code": 0,
			"Alarm text": "",
		},
	}
	node.AddOrUpdate(alarm)

	// Read values from pump
	go func() {
		for {
			result, err := nibe.Read(uint16(45001))
			if err != nil {
				logrus.Errorf("Failed to read error register %d: %s", 45001, err)
			} else {
				newState := make(devices.State)

				param, err := nibe.Describe(uint16(45001))
				if err == nil {
					// Special for alarms
					newState["on"] = result != 0
					newState["Alarm code"] = result
					newState["Alarm text"] = param.Map[strconv.Itoa(int(result))]
				}

				if !alarm.Online {
					alarm.Lock()
					alarm.Online = true
					alarm.State.MergeWith(newState)
					alarm.Unlock()
					node.AddOrUpdate(alarm)
				} else {
					node.UpdateState(alarm.ID.ID, newState)
				}
			}

			for r, dev := range list {
				result, err := nibe.Read(uint16(r))
				if err != nil {
					logrus.Errorf("Failed to read register %d: %s", r, err)
					continue
				}

				newState := make(devices.State)

				param, err := nibe.Describe(uint16(r))
				if err != nil {
					log.Printf("Read %d (%s) = %d", r, name[int(r)], result)
					newState[name[int(r)]] = result
				} else {
					log.Printf("Read %d (%s) = %d", r, param.Title, result)
					newState[name[int(r)]] = float64(result) / float64(param.Factor)
				}

				if !dev.Online {
					dev.Lock()
					dev.Online = true
					dev.State.MergeWith(newState)
					dev.Unlock()
					node.AddOrUpdate(dev)
				} else {
					node.UpdateState(dev.ID.ID, newState)
				}
			}
		}
	}()

	nibe.OnUpdate(func(reg uint16, value int16) {
		newState := make(devices.State)
		dev := other

		param, err := nibe.Describe(reg)

		if d, ok := list[int(reg)]; ok {
			if err != nil {
				newState[name[int(reg)]] = value
			} else {
				newState[name[int(reg)]] = float64(value) / float64(param.Factor)
			}
			dev = d
		} else {
			if err != nil {
				newState[strconv.Itoa(int(reg))] = value
			} else {
				newState[param.Title] = float64(value) / float64(param.Factor)
			}
		}

		if !dev.Online {
			dev.Lock()
			dev.Online = true
			dev.State.MergeWith(newState)
			dev.Unlock()
			node.AddOrUpdate(dev)
		} else {
			node.UpdateState(dev.ID.ID, newState)
		}
	})

	node.Wait()
	nibe.Stop()
}
