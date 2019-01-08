package logic

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models/devices"
)

type SavedState struct {
	Name  string
	UUID  string
	State map[devices.ID]devices.State
}
type SavedStateStore struct {
	State map[string]*SavedState
	sync.RWMutex
}

func (sss *SavedStateStore) Get(id string) *SavedState {
	sss.RLock()
	defer sss.RUnlock()
	return sss.State[id]

}

func NewSavedStateStore() *SavedStateStore {
	return &SavedStateStore{
		State: make(map[string]*SavedState),
	}
}

func (sss *SavedStateStore) Save(path string) error {
	sss.Lock()
	defer sss.Unlock()
	configFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("savedstate: error saving state: %s", err.Error())
	}
	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "\t")
	err = encoder.Encode(sss.State)
	if err != nil {
		return fmt.Errorf("savedstate: error saving state: %s", err.Error())
	}
	return nil
}

func (sss *SavedStateStore) Load(path string) error {
	configFile, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // We dont want to error our if the file does not exist when we start the server
		}
		return fmt.Errorf("savedstate: error loading state: %s", err.Error())
	}

	sss.Lock()
	defer sss.Unlock()
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&sss.State); err != nil {
		return fmt.Errorf("savedstate: error loading state: %s", err.Error())
	}
	return nil
}
