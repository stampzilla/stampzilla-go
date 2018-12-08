package store

import (
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
		//if node.Devices != nil {
		//store.Nodes[node.UUID].Devices = node.Devices
		//}
		if node.Config != nil {
			logrus.Info("Setting config to: ", string(node.Config))
			store.Nodes[node.UUID].Config = node.Config
		}

	}

	store.Unlock()

	store.runCallbacks("nodes")
}
