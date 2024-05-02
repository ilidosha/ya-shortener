package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"net/url"
	"shortener/internal/config"
	"shortener/internal/store"
)

type ShortenURLRequest struct {
	LongURL string `json:"url"`
}

type ShortenURLResponse struct {
	ShortURL string `json:"result"`
}

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
		store.Store.Save(shortURL, string(longURL), "", opts)

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
	http.Redirect(w, r, longURL.Value, http.StatusTemporaryRedirect)
}

func ShortenURLFromJSON(opts *config.Options) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read the long URL from the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}
		var request ShortenURLRequest
		err = json.Unmarshal(body, &request)
		if err != nil {
			http.Error(w, "Error unmarshalling request body", http.StatusBadRequest)
			return
		}

		// Check if the URL is valid
		if _, err := url.ParseRequestURI(request.LongURL); err != nil {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		// Check if the URL is already in the store
		if shortURL, ok := store.Store.ValueExistsInMap(request.LongURL); ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			response := ShortenURLResponse{
				ShortURL: fmt.Sprintf("%s/%s", opts.BaseURL, shortURL),
			}
			responseJSON, _ := json.Marshal(response)
			_, _ = w.Write(responseJSON)
			return
		}

		// Generate a short URL
		shortURL := generateShortURL(request.LongURL, store.Store.GetStore())

		// Save the URL
		store.Store.Save(shortURL, request.LongURL, "", opts)

		// Return the short URL
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		response := ShortenURLResponse{
			ShortURL: fmt.Sprintf("%s/%s", opts.BaseURL, shortURL),
		}
		responseJSON, _ := json.Marshal(response)
		_, _ = w.Write(responseJSON)
	}
}
