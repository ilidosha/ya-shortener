package handlers

import (
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"shortener/internal/config"
	"shortener/internal/store"
)

func ShortenURL(opts *config.Options) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read the long URL from the request body
		longURL, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}

		// Generate a short URL
		shortURL := fmt.Sprintf("%d", len(store.URLStore)+1)

		// Store the URL
		store.URLStore[shortURL] = string(longURL)

		// Return the short URL
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "%s/%s", opts.BaseURL, shortURL)
	}
}

func RedirectToURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortURL := vars["shortURL"]

	// Look up the long URL
	longURL, ok := store.URLStore[shortURL]
	if !ok {
		http.NotFound(w, r)
		return
	}

	// Redirect to the long URL
	http.Redirect(w, r, longURL, http.StatusTemporaryRedirect)
}
