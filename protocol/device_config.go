package protocol

import "github.com/stampzilla/stampzilla-go/protocol/devices"

type ConfigMap struct {
	node   Identifiable
	Config map[string]*DeviceConfigMap `json:"config"`
}

type DeviceConfigMap struct {
	Layout_ map[string]*DeviceConfig `json:"layout"`
	handler func(devices.Device, *DeviceConfig)
}

func (cm *ConfigMap) Add(devid string) *DeviceConfigMap {
	if dcm, ok := cm.Config[cm.node.Uuid()+"."+devid]; ok {
		return dcm
	}

	dcm := &DeviceConfigMap{
		Layout_: make(map[string]*DeviceConfig),
	}
	cm.Config[cm.node.Uuid()+"."+devid] = dcm

	return dcm
}

func (cm *DeviceConfigMap) Layout(layout ...*DeviceConfig) *DeviceConfigMap {
	for _, v := range layout {
		if _, ok := cm.Layout_[v.ID]; ok {
			continue
		}

		cm.Layout_[v.ID] = v
	}

	return cm
}

func (cm *DeviceConfigMap) Handler(f func(devices.Device, *DeviceConfig)) *DeviceConfigMap {
	cm.handler = f

	return cm
}

type DeviceConfig struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Type    string            `json:"type"`
	Options map[string]string `json:"options,omitempty"`
	Min     int               `json:"min,omitempty"`
	Max     int               `json:"max,omitempty"`
	Value   interface{}       `json:"value,omitempty"`
}
