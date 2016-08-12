package devices

//type DeviceInterface interface {
//[> TODO: add methods <]
//SetName(string)
//SyncState(interface{})
//}

type Device struct {
	Type     string                 `json:"type"`
	Node     string                 `json:"node,omitempty"`
	Id       string                 `json:"id,omitempty"`
	Name     string                 `json:"name,omitempty"`
	Online   bool                   `json:"online"`
	StateMap map[string]string      `json:"stateMap"`
	State    map[string]interface{} `json:"state,omitempty"`
}

//func (d *Device) SetName(name string) {
//d.Name = name
//}
func (d *Device) SyncState(state interface{}) {

}

// key should be nodeuuid.deviceid
type Map map[string]*Device
