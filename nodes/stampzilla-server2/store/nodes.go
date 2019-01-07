package store

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server2/models"
)

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

func (store *Store) AddOrUpdateNode(node *models.Node) {
	store.Lock()

	if _, ok := store.Nodes[node.UUID]; !ok {
		store.Nodes[node.UUID] = node
	} else {

		if node.Version != "" {
			store.Nodes[node.UUID].Version = node.Version
		}
		if node.Type != "" {
			store.Nodes[node.UUID].Type = node.Type
		}
		if node.Name != "" {
			store.Nodes[node.UUID].Name = node.Name
		}
		if node.Config != nil {
			logrus.Info("Setting config to: ", string(node.Config))
			store.Nodes[node.UUID].Config = node.Config
		}

	}

	store.Unlock()

	store.runCallbacks("nodes")
}
func (store *Store) SaveNodes() error {
	nodes := store.GetNodes()

	for _, node := range nodes {
		err := saveConfigToFile(node)
		if err != nil {
			return err
		}
	}

	return nil
}

func (store *Store) LoadNodes() error {
	path := "configs"

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

		node, err := loadNodeConfigFromFile(filepath.Join(path, f.Name()))
		if err != nil {
			return err
		}
		store.AddOrUpdateNode(node)
	}
	return nil
}
func loadNodeConfigFromFile(file string) (*models.Node, error) {
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
