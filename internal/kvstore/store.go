package kvstore

import (
	"log/slog"
	"sync"
)

type Store interface {
	AddItem(id string, value []byte)
	Fetch(id string) ([]byte, bool)
}

type store struct {
	store map[string][]byte
	mutex sync.RWMutex
}

//Get a new KeyValue store
func NewKVStore() *store {
	return &store{
		store: make(map[string][]byte),
	}
}

//Store the string value in the KVStore under ID
func (s *store) AddItem(id string, value []byte) {
	slog.Info("kvStore:Additem","id",id)
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.store[id] = value
}

//Retrieve the stored value 
func (s *store) Fetch(id string) ([]byte, bool) {
	slog.Info("kvStore:Fetchitem","id",id)
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	value, ok := s.store[id]
	return value, ok
}