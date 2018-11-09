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

type Store struct {
	Nodes       Nodes
	Connections Connections
	onUpdate    []func(*Store) error
	sync.RWMutex
}

func New() *Store {
	return &Store{
		Nodes:       make(Nodes),
		Connections: make(Connections),
		onUpdate:    make([]func(*Store) error, 0),
	}
}

func (store *Store) AddOrUpdateNode(node *models.Node) {
	store.Lock()
	store.Nodes[node.UUID] = node
	store.Unlock()

	for _, callback := range store.onUpdate {
		if err := callback(store); err != nil {
			logrus.Error("store: ", err)
		}
	}
}

func (store *Store) Connection(id string) *models.Connection {
	store.RLock()
	defer store.RUnlock()
	if conn, ok := store.Connections["foo"]; ok {
		return conn
	}
	return nil
}

func (store *Store) AddOrUpdateConnection(id string, c *models.Connection) {
	store.Lock()
	store.Connections[id] = c
	store.Unlock()

	for _, callback := range store.onUpdate {
		if err := callback(store); err != nil {
			logrus.Error("store: ", err)
		}
	}
}

func (store *Store) RemoveConnection(id string) {
	store.Lock()
	delete(store.Connections, id)
	store.Unlock()

	for _, callback := range store.onUpdate {
		if err := callback(store); err != nil {
			logrus.Error("store: ", err)
		}
	}
}

func (store *Store) GetNodes() Nodes {
	store.RLock()
	defer store.RUnlock()
	return store.Nodes
}

func (store *Store) GetNode(uuid string) *models.Node {
	store.RLock()
	defer store.RUnlock()
	n, _ := store.Nodes[uuid]
	return n
}

func (store *Store) GetConnections() Connections {
	store.RLock()
	defer store.RUnlock()
	return store.Connections
}

func (store *Store) OnUpdate(callback func(*Store) error) {
	store.Lock()
	store.onUpdate = append(store.onUpdate, callback)
	store.Unlock()
}

func (store *Store) LoadFromDisk() error {
	path := "configs/"
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
	b, _ := ioutil.ReadAll(f)
	json.Unmarshal([]byte(b), &node)

	return node, nil
}
