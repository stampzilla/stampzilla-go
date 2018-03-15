package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"

	"github.com/RangelReale/osin"
	"github.com/sirupsen/logrus"
)

type JSONStorage struct {
	Clients         map[string]osin.Client
	Authorize       map[string]*osin.AuthorizeData
	Access          map[string]*osin.AccessData
	Refresh         map[string]string
	AccessClientMap map[string]string
	sync.RWMutex
}

func NewTestStorage() *JSONStorage {
	r := &JSONStorage{
		Clients:         make(map[string]osin.Client),
		Authorize:       make(map[string]*osin.AuthorizeData),
		Access:          make(map[string]*osin.AccessData),
		Refresh:         make(map[string]string),
		AccessClientMap: make(map[string]string),
	}

	return r
}

func (s *JSONStorage) Clone() osin.Storage {
	return s
}

func (s *JSONStorage) Close() {
}

func (s *JSONStorage) LoadFromDisk(filename string) {

	file, err := os.Open(filename)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer file.Close()

	dec := json.NewDecoder(file)
	dec.Decode(s)

}
func (s *JSONStorage) saveToDisk(filename string) {

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

func (s *JSONStorage) Unlock() {
	s.saveToDisk("storage.json")
	s.RWMutex.Unlock()
}

func (s *JSONStorage) GetClient(id string) (osin.Client, error) {
	s.RLock()
	defer s.RUnlock()
	log.Printf("GetClient: %s\n", id)
	if c, ok := s.Clients[id]; ok {
		return c, nil
	}
	return nil, osin.ErrNotFound
}

func (s *JSONStorage) SetClient(id string, client osin.Client) error {
	s.Lock()
	log.Printf("SetClient: %s\n", id)
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

func (s *JSONStorage) SaveAuthorize(data *osin.AuthorizeData) error {
	s.Lock()
	log.Printf("SaveAuthorize: %s\n", data.Code)
	s.Authorize[data.Code] = data
	s.Unlock()
	return nil
}

func (s *JSONStorage) LoadAuthorize(code string) (*osin.AuthorizeData, error) {
	s.RLock()
	defer s.RUnlock()
	log.Printf("LoadAuthorize: %s\n", code)
	if d, ok := s.Authorize[code]; ok {
		return d, nil
	}
	return nil, osin.ErrNotFound
}

func (s *JSONStorage) RemoveAuthorize(code string) error {
	log.Printf("RemoveAuthorize: %s\n", code)
	s.Lock()
	delete(s.Authorize, code)
	s.Unlock()
	return nil
}

func (s *JSONStorage) SaveAccess(data *osin.AccessData) error {
	log.Printf("SaveAccess: %s\n", data.AccessToken)
	s.Lock()
	s.Access[data.AccessToken] = data
	if data.RefreshToken != "" {
		s.Refresh[data.RefreshToken] = data.AccessToken
	}
	s.Unlock()
	return nil
}

func (s *JSONStorage) LoadAccess(code string) (*osin.AccessData, error) {
	s.RLock()
	defer s.RUnlock()
	log.Printf("LoadAccess: %s\n", code)
	if d, ok := s.Access[code]; ok {
		return d, nil
	}
	return nil, osin.ErrNotFound
}

func (s *JSONStorage) RemoveAccess(code string) error {
	log.Printf("RemoveAccess: %s\n", code)
	s.Lock()
	delete(s.Access, code)
	s.Unlock()
	return nil
}

func (s *JSONStorage) LoadRefresh(code string) (*osin.AccessData, error) {
	s.RLock()
	defer s.RUnlock()
	log.Printf("LoadRefresh: %s\n", code)
	if d, ok := s.Refresh[code]; ok {
		return s.LoadAccess(d)
	}
	return nil, osin.ErrNotFound
}

func (s *JSONStorage) RemoveRefresh(code string) error {
	log.Printf("RemoveRefresh: %s\n", code)
	s.Lock()
	delete(s.Refresh, code)
	s.Unlock()
	return nil
}
