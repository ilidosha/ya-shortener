package store

import (
	"github.com/rs/zerolog/log"
	"shortener/internal/config"
	"sync"
)

// ValueExistsInMap Function to check if a value exists in a sync.Map
func (s *URLStore) ValueExistsInMap(searchValue string) (string, bool) {
	var key string
	found := false
	log.Info().Msgf("key %v", key)

	s.URLs.Range(func(k, value interface{}) bool {
		if mapValues, ok := value.(MapValues); ok {
			if mapValues.Value == searchValue {
				key = k.(string)
				log.Info().Msgf("Value found in map %v, search value %v, key %v, k %v", value, searchValue, key, k)
				found = true
				return false // Stop iterating
			}
		} else {
			log.Warn().Msg("Value is not of type MapValues")
		}
		return true // Continue iterating
	})

	return key, found
}

func New() {
	Store = URLStore{URLs: &sync.Map{}}
}

// Time savers

// Save Function to store the URL
func (s *URLStore) Save(key, value, uuid string, options *config.Options) {
	fileStoreExixts := options.FileStore != ""
	var values MapValues
	if uuid != "" {
		values = MapValues{
			Value: value,
			UUID:  uuid,
		}
	}
	if uuid == "" { // Generate uuid if uuid is not set
		values = MapValues{
			Value: value,
			UUID:  GenerateUUID(),
		}
	}
	s.URLs.Store(key, values)
	// Save to file if FileStore is set
	if fileStoreExixts {
		err := s.SaveToFile(options.FileStore)
		if err != nil {
			if err.Error() != "key already exists" {
				log.Error().Err(err).Msg("Failed to save to file")
				return
			}
			log.Error().Err(err).Msg("Failed to save to file")
		}
	}
}

// Find Function to find the URL
func (s *URLStore) Find(key string) (MapValues, bool) {
	value, ok := s.URLs.Load(key)
	return value.(MapValues), ok
}

// Delete Function to delete the URL
func (s *URLStore) Delete(key string) {
	s.URLs.Delete(key)
}

func (s *URLStore) GetStore() *sync.Map {
	return s.URLs
}
