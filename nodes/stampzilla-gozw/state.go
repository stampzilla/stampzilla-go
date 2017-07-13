package main

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/davecgh/go-spew/spew"
	"github.com/gozwave/gozw/application"
)

type state struct {
	HomeID              string `json:"home_id"`
	APIVersion          string `json:"api_version"`
	Library             string `json:"library"`
	Version             string `json:"version"`
	APIType             string `json:"api_type"`
	IsPrimaryController bool   `json:"is_primary_controller"`

	Nodes map[string]*znode `json:"nodes"`

	updateNotifications chan struct{}
	sync.RWMutex
}

func newState() *state {
	return &state{
		Nodes:               make(map[string]*znode, 0),
		updateNotifications: make(chan struct{}),
	}
}

func (s *state) sync(nodeID uint8, node application.Node) {
	fmt.Println("Node updated event")
	spew.Dump(node)

	s.RLock()
	n, _ := s.Nodes[strconv.Itoa(int(nodeID))]
	s.RUnlock()
	if n == nil {
		n = &znode{
			ID:     int(nodeID),
			parent: s,
		}
		s.Lock()
		s.Nodes[strconv.Itoa(int(nodeID))] = n
		s.Unlock()
	}

	n.sync(node)
}

type znode struct {
	parent  *state
	Failing bool `json:"failing"`
	Awake   bool `json:"awake"`
	Secure  bool `json:"secure"`

	ID                int                `json:"id"`
	DeviceClass       znode_deviceClass  `json:"device_class"`
	ManufacturerInfo  znode_manufacturer `json:"manufacturer_info"`
	Brand             string             `json:"brand"`
	Product           string             `json:"product"`
	InterviewProgress float64            `json:"interview_progress"`

	StateFloat map[string]float64 `json:"stateFloat"`
	StateBool  map[string]bool    `json:"stateBool"`
}

type znode_deviceClass struct {
	Basic    string `json:"basic"`
	Generic  string `json:"generic"`
	Specific string `json:"specific"`
}

type znode_manufacturer struct {
	ManufacturerID string `json:"manufacturer_id"`
	ProductType    string `json:"product_type"`
	ProductID      string `json:"product_id"`
}

func (zn *znode) sync(node application.Node) {

	zn.Failing = node.Failing
	zn.Awake = node.IsListening()
	zn.Secure = node.IsSecure()
	zn.DeviceClass.Basic = node.GetBasicDeviceClassName()
	zn.DeviceClass.Generic = node.GetGenericDeviceClassName()
	zn.DeviceClass.Specific = node.GetSpecificDeviceClassName()
	zn.ManufacturerInfo.ManufacturerID = fmt.Sprintf("0x%04X", node.ManufacturerID)
	zn.ManufacturerInfo.ProductType = fmt.Sprintf("0x%04X", node.ProductTypeID)
	zn.ManufacturerInfo.ProductID = fmt.Sprintf("0x%04X", node.ProductID)
	zn.InterviewProgress = node.GetInterviewProgress()

	zn.publishUpdate()
}

func (zn *znode) publishUpdate() {
	select {
	case zn.parent.updateNotifications <- struct{}{}:
	default:
	}
}
