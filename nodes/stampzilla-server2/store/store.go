package store

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
)

type Nodes map[string]*models.Node
type Connections map[string]*models.Connection
type UpdateCallback func(*Store) error

type Store struct {
	Nodes       Nodes
	Devices     *models.Devices
	Connections Connections
	onUpdate    map[string][]UpdateCallback
	sync.RWMutex
}

func New() *Store {
	return &Store{
		Nodes:       make(Nodes),
		Devices:     models.NewDevices(),
		Connections: make(Connections),
		onUpdate:    make(map[string][]UpdateCallback, 0),
	}
}

func (store *Store) runCallbacks(area string) {
	for _, callback := range store.onUpdate[area] {
		if err := callback(store); err != nil {
			logrus.Error("store: ", err)
		}
	}
}

func (store *Store) OnUpdate(area string, callback UpdateCallback) {
	if _, ok := store.onUpdate[area]; !ok {
		store.onUpdate[area] = make([]UpdateCallback, 0)
	}
	store.Lock()
	store.onUpdate[area] = append(store.onUpdate[area], callback)
	store.Unlock()
}

func (store *Store) LoadFromDisk() error {
	path := "configs/"

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	for _, f := range files {
		if f.Name()[0] == '.' {
			continue
		}

		node, err := loadConfigFromFile(path + f.Name())
		if err != nil {
			return err
		}
		store.AddOrUpdateNode(node)
	}

	return nil
}

func (store *Store) WriteToDisk() error {
	nodes := store.GetNodes()

	for _, node := range nodes {
		err := saveConfigToFile(node)
		if err != nil {
			return err
		}
	}

	return nil
}

func saveConfigToFile(node *models.Node) error {
	path := "configs/"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0755)
	}

	configFile, err := os.Create(path + node.UUID + ".json")
	if err != nil {
		return err
	}

	var out bytes.Buffer
	b, err := json.MarshalIndent(node, "", "\t")
	if err != nil {
		return err
	}
	json.Indent(&out, b, "", "\t")
	_, err = out.WriteTo(configFile)
	return err
}

func loadConfigFromFile(file string) (*models.Node, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var node *models.Node

	decoder := json.NewDecoder(f)
	err = decoder.Decode(&node)

	return node, err
}
