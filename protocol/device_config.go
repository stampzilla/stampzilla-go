package protocol

type ConfigMap struct {
	node   Identifiable
	Config map[string]*DeviceConfigMap `json:"config"`
}

func NewConfigMap(n Identifiable) *ConfigMap {
	return &ConfigMap{
		node:   n,
		Config: make(map[string]*DeviceConfigMap),
	}
}

func (cm *ConfigMap) ListenForConfigChanges(c chan DeviceConfigSet) {
	go func() {
		for conf := range c {
			cm.SetConfig(conf.Device, conf.ID, conf.Value)
		}
	}()
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

func (cm *ConfigMap) SetConfig(device, id string, value interface{}) {
	dev, ok := cm.Config[cm.node.Uuid()+"."+device]
	if !ok {
		return
	}

	dc, ok := dev.Layout_[id]
	if !ok {
		return
	}

	dc.Value = value

	dev.handler(device, dc)
}

type DeviceConfigMap struct {
	Layout_ map[string]*DeviceConfig `json:"layout"`
	handler func(string, *DeviceConfig)
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

func (cm *DeviceConfigMap) Handler(f func(string, *DeviceConfig)) *DeviceConfigMap {
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

type DeviceConfigSet struct {
	Device string `json:"device"`
	ID     string `json:"id"`
	Value  interface{}
}
