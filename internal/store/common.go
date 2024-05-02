package store

import (
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"sync"
)

var Store URLStore

// URLStore a map to store the URLs
type URLStore struct {
	URLs *sync.Map
}

// MapValues a struct to represent values in ORLStore.URLs sync Map
type MapValues struct {
	Value string
	UUID  string
}

func GenerateUUID() string {
	// Generate a UUID for each record
	id, err := uuid.NewRandom()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to generate UUID")
	}
	return id.String()
}
