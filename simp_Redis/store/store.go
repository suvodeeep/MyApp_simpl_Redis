package store

import (
	"errors"
	"sync"
	"time"
)

// Common error handling
var (
	ErrKeyNotFound = errors.New("key not found")
	ErrKeyExists   = errors.New("key already exists")
)

// Here, Item represents value stored in the db with optional expiration
type Item struct {
	Value      string
	Expiration *time.Time
}

// Store is the struct /the data structure that holds the key-value pairs
type Store struct {
	data  map[string]Item
	mutex sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]Item),
	}
}

// The "Get" retrieves a value from the store by key
func (s *Store) Get(key string) (string, error) {

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	item, exists := s.data[key]
	if !exists {
		return "", ErrKeyNotFound
	}

	if item.Expiration != nil && time.Now().After(*item.Expiration) {

		return "", ErrKeyNotFound
	}

	return item.Value, nil
}

// The "Set" adds or updates a key-value pair in the store
func (s *Store) Set(key, value string, expiration *time.Time) {
	// We use Lock for write operations to prevent concurrent writes
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.data[key] = Item{
		Value:      value,
		Expiration: expiration,
	}
}

// here, the "SetNX (Set if Not eXists) sets a key only if it doesn't already exist
func (s *Store) SetNX(key, value string, expiration *time.Time) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if item, exists := s.data[key]; exists {
		if item.Expiration == nil || time.Now().Before(*item.Expiration) {
			return ErrKeyExists
		}
	}

	s.data[key] = Item{
		Value:      value,
		Expiration: expiration,
	}
	return nil
}

// The "Delete" removes a key from the store
func (s *Store) Delete(key string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, exists := s.data[key]
	if !exists {
		return false
	}

	delete(s.data, key)
	return true
}

// "Keys" returns all keys in the store that match the given pattern

func (s *Store) Keys() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	keys := make([]string, 0, len(s.data))
	for key, item := range s.data {

		if item.Expiration != nil && time.Now().After(*item.Expiration) {
			continue
		}
		keys = append(keys, key)
	}
	return keys
}

// CleanupExpired removes all expired keys from the store to free up memory
func (s *Store) CleanupExpired() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for key, item := range s.data {
		if item.Expiration != nil && time.Now().After(*item.Expiration) {
			delete(s.data, key)
		}
	}
}
