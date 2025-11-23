package kvstore

import (
	"log/slog"
	"sync"
)

type Store interface {
	AddItem(id string, value string)
	Fetch(id string) (string, bool)
}

type store struct {
	store map[string]string
	mutex sync.RWMutex
}

//Get a new KeyValue store
func NewKVStore() *store {
	return &store{
		store: make(map[string]string),
	}
}

//Store the string value in the KVStore under ID
func (s *store) AddItem(id string, value string) {
	slog.Info("kvStore:Additem","id",id)
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.store[id] = value
}

//Retrieve the stored value 
func (s *store) Fetch(id string) (string, bool) {
	slog.Info("kvStore:Fetchitem","id",id)
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	value, ok := s.store[id]
	return value, ok
}