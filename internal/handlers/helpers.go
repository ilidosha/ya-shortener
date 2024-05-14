package handlers

import (
	"crypto/sha1"
	"encoding/base64"
	"sync"
)

func generateShortURL(originalURL string, store *sync.Map) string {
	hash := sha1.New()
	hash.Write([]byte(originalURL))
	shortURL := base64.URLEncoding.EncodeToString(hash.Sum(nil))[:6]

	// Check for collisions and regenerate short URL if it already exists in the store
	for {
		if _, ok := store.Load(shortURL); !ok {
			break
		}
		// Regenerate short URL
		shortURL = base64.URLEncoding.EncodeToString(hash.Sum([]byte("some")))[:6]
	}

	return shortURL
}

func generateShortURLWithoutCheck(originalURL string) string {
	hash := sha1.New()
	hash.Write([]byte(originalURL))
	shortURL := base64.URLEncoding.EncodeToString(hash.Sum(nil))[:6]
	return shortURL
}
