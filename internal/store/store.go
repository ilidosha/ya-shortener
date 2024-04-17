package store

import "sync"

var Store URLStore

// URLStore a map to store the URLs
type URLStore struct {
	URLs *sync.Map
}

// ValueExistsInMap Function to check if a value exists in a sync.Map
func (s *URLStore) ValueExistsInMap(searchValue string) (string, bool) {
	var key string
	found := false

	s.URLs.Range(func(k, value interface{}) bool {
		if value == searchValue {
			key = k.(string)
			found = true
			return false // Stop iterating
		}
		return true // Continue iterating
	})

	return key, found
}

func Init() {
	Store = URLStore{URLs: &sync.Map{}}
}

// Time savers

// Save Function to store the URL
func (s *URLStore) Save(key string, value string) {
	s.URLs.Store(key, value)
}

// Find Function to find the URL
func (s *URLStore) Find(key string) (any, bool) {
	value, ok := s.URLs.Load(key)
	return value, ok
}

// Delete Function to delete the URL
func (s *URLStore) Delete(key string) {
	s.URLs.Delete(key)
}

func (s *URLStore) GetStore() *sync.Map {
	return s.URLs
}
