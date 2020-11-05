package logic

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
)

/* savedstate.json example
{
    "6fbaea24-6b3f-4856-9194-735b349bbf4d": {
        "name": "test state 1",
        "uuid": "6fbaea24-6b3f-4856-9194-735b349bbf4d",
        "state": {
            "nodeuuid.deviceid": {
                "on": true
            }
        }
    }
}
*/

type SavedStates map[string]*SavedState

type SavedState struct {
	Name  string                       `json:"name"`
	UUID  string                       `json:"uuid"`
	State map[devices.ID]devices.State `json:"state"`
}

type SavedStateStore struct {
	State SavedStates
	sync.RWMutex
}

func (sss *SavedStateStore) Get(id string) *SavedState {
	sss.RLock()
	defer sss.RUnlock()
	return sss.State[id]
}

func (sss *SavedStateStore) All() SavedStates {
	sss.RLock()
	defer sss.RUnlock()
	return sss.State
}

func NewSavedStateStore() *SavedStateStore {
	return &SavedStateStore{
		State: make(map[string]*SavedState),
	}
}

func (sss *SavedStateStore) SetState(s SavedStates) {
	sss.Lock()
	sss.State = s
	sss.Unlock()
}

func (sss *SavedStateStore) Save() error {
	sss.Lock()
	defer sss.Unlock()
	configFile, err := os.Create("savedstate.json")
	if err != nil {
		return fmt.Errorf("savedstate: error saving savedstate.json: %s", err.Error())
	}
	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "\t")
	err = encoder.Encode(sss.State)
	if err != nil {
		return fmt.Errorf("savedstate: error saving savedstate.json: %s", err.Error())
	}
	return nil
}

func (sss *SavedStateStore) Load() error {
	configFile, err := os.Open("savedstate.json")
	if err != nil {
		if os.IsNotExist(err) {
			return nil // We dont want to error our if the file does not exist when we start the server
		}
		return fmt.Errorf("savedstate: error loading savedstate.json: %s", err.Error())
	}

	sss.Lock()
	defer sss.Unlock()
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&sss.State); err != nil {
		return fmt.Errorf("savedstate: error loading savedstate.json: %s", err.Error())
	}
	return nil
}
