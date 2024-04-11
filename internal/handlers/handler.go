package handlers

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"net/url"
	"shortener/internal/config"
	"shortener/internal/store"
	"sync"
)

func ShortenURL(opts *config.Options) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read the long URL from the request body
		longURL, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}

		// Check if the URL is valid
		if _, err := url.ParseRequestURI(string(longURL)); err != nil {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		// Check if the URL is already in the store
		if shortURL, ok := store.Store.ValueExistsInMap(string(longURL)); ok {
			w.WriteHeader(http.StatusConflict)
			_, _ = fmt.Fprintf(w, "%s/%s", opts.BaseURL, shortURL)
			return
		}

		// Generate a short URL
		shortURL := generateShortURL(string(longURL), store.Store.GetStore())

		// Save the URL
		store.Store.Save(shortURL, string(longURL))

		// Return the short URL
		w.WriteHeader(http.StatusCreated)
		_, _ = fmt.Fprintf(w, "%s/%s", opts.BaseURL, shortURL)
	}
}

func RedirectToURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortURL := vars["shortURL"]

	// Look up the long URL
	longURL, ok := store.Store.Find(shortURL)
	if !ok {
		http.NotFound(w, r)
		return
	}

	// Redirect to the long URL
	http.Redirect(w, r, longURL.(string), http.StatusTemporaryRedirect)
}

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
