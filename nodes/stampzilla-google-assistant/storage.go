package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/RangelReale/osin"
	"github.com/sirupsen/logrus"
)

type TestStorage struct {
	Clients         map[string]osin.Client
	Authorize       map[string]*osin.AuthorizeData
	Access          map[string]*osin.AccessData
	Refresh         map[string]string
	AccessClientMap map[string]string
	sync.RWMutex
}

func NewTestStorage() *TestStorage {
	r := &TestStorage{
		Clients:         make(map[string]osin.Client),
		Authorize:       make(map[string]*osin.AuthorizeData),
		Access:          make(map[string]*osin.AccessData),
		Refresh:         make(map[string]string),
		AccessClientMap: make(map[string]string),
	}

	return r
}

func (s *TestStorage) Clone() osin.Storage {
	return s
}

func (s *TestStorage) Close() {
}

func (s *TestStorage) LoadFromDisk(filename string) {

	file, err := os.Open(filename)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer file.Close()

	dec := json.NewDecoder(file)
	dec.Decode(s)

}
func (s *TestStorage) saveToDisk(filename string) {

	file, err := os.Create(filename)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer file.Close()

	for _, v := range s.Access {
		s.AccessClientMap[v.AccessToken] = v.Client.GetId()
	}

	enc := json.NewEncoder(file)
	enc.SetIndent("", "\t")
	enc.Encode(s)
}

func (s *TestStorage) Unlock() {
	s.saveToDisk("storage.json")
	s.RWMutex.Unlock()
}

func (s *TestStorage) GetClient(id string) (osin.Client, error) {
	s.RLock()
	defer s.RUnlock()
	fmt.Printf("GetClient: %s\n", id)
	if c, ok := s.Clients[id]; ok {
		return c, nil
	}
	return nil, osin.ErrNotFound
}

func (s *TestStorage) SetClient(id string, client osin.Client) error {
	s.Lock()
	fmt.Printf("SetClient: %s\n", id)
	s.Clients[id] = client
	for _, v := range s.Access {
		if clientID, ok := s.AccessClientMap[v.AccessToken]; ok {
			if client, ok := s.Clients[clientID]; ok {
				v.Client = client
			}
		}
	}
	s.Unlock()
	return nil
}

func (s *TestStorage) SaveAuthorize(data *osin.AuthorizeData) error {
	s.Lock()
	fmt.Printf("SaveAuthorize: %s\n", data.Code)
	s.Authorize[data.Code] = data
	s.Unlock()
	return nil
}

func (s *TestStorage) LoadAuthorize(code string) (*osin.AuthorizeData, error) {
	s.RLock()
	defer s.RUnlock()
	fmt.Printf("LoadAuthorize: %s\n", code)
	if d, ok := s.Authorize[code]; ok {
		return d, nil
	}
	return nil, osin.ErrNotFound
}

func (s *TestStorage) RemoveAuthorize(code string) error {
	fmt.Printf("RemoveAuthorize: %s\n", code)
	s.Lock()
	delete(s.Authorize, code)
	s.Unlock()
	return nil
}

func (s *TestStorage) SaveAccess(data *osin.AccessData) error {
	fmt.Printf("SaveAccess: %s\n", data.AccessToken)
	s.Lock()
	s.Access[data.AccessToken] = data
	if data.RefreshToken != "" {
		s.Refresh[data.RefreshToken] = data.AccessToken
	}
	s.Unlock()
	return nil
}

func (s *TestStorage) LoadAccess(code string) (*osin.AccessData, error) {
	s.RLock()
	defer s.RUnlock()
	fmt.Printf("LoadAccess: %s\n", code)
	if d, ok := s.Access[code]; ok {
		return d, nil
	}
	return nil, osin.ErrNotFound
}

func (s *TestStorage) RemoveAccess(code string) error {
	fmt.Printf("RemoveAccess: %s\n", code)
	s.Lock()
	delete(s.Access, code)
	s.Unlock()
	return nil
}

func (s *TestStorage) LoadRefresh(code string) (*osin.AccessData, error) {
	s.RLock()
	defer s.RUnlock()
	fmt.Printf("LoadRefresh: %s\n", code)
	if d, ok := s.Refresh[code]; ok {
		return s.LoadAccess(d)
	}
	return nil, osin.ErrNotFound
}

func (s *TestStorage) RemoveRefresh(code string) error {
	fmt.Printf("RemoveRefresh: %s\n", code)
	s.Lock()
	delete(s.Refresh, code)
	s.Unlock()
	return nil
}
