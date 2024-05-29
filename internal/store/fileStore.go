package store

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type fileStore struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// SaveToFile saves the short URL to a file
func (s *URLStore) SaveToFile(filePath string) error {
	// Create a slice of fileStore structs to store the URLs
	var fileStores []fileStore

	// Iterate over the URLs in the URLStore
	s.URLs.Range(func(key, value interface{}) bool {
		// Convert the value to MapValues
		mapValues := value.(MapValues)

		// Create a fileStore struct with the necessary information
		fileStore := fileStore{
			UUID:        mapValues.UUID,
			ShortURL:    key.(string),
			OriginalURL: mapValues.Value,
		}

		// Append the fileStore struct to the slice
		fileStores = append(fileStores, fileStore)

		return true
	})
	// Create the directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}

	// Create a file
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Encode the fileStores slice as JSON and write it to the file
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(fileStores)
	if err != nil {
		return err
	}

	return nil
}

// LoadFromFile loads the short URLs from a file
func (s *URLStore) LoadFromFile(filePath string) error {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Decode the file contents as JSON into a slice of fileStore structs
	var fileStores []fileStore
	err = json.NewDecoder(file).Decode(&fileStores)
	if err != nil {
		return err
	}

	// Iterate over the fileStores slice and add the URLs to the URLStore
	for _, fileStore := range fileStores {
		mapValues := MapValues{
			Value: fileStore.OriginalURL,
			UUID:  fileStore.UUID,
		}
		s.URLs.Store(fileStore.ShortURL, mapValues)
	}

	return nil
}
