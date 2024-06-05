package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"net/url"
	"shortener/internal/config"
	"shortener/internal/generator"
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
		dbExists := opts.ConnectionString != ""
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

		// Check if exists in DB
		if dbExists {
			if shortURL, exists := store.CheckIfExistsInDB(string(longURL)); exists {
				w.WriteHeader(http.StatusConflict)
				_, _ = fmt.Fprintf(w, "%s/%s", opts.BaseURL, shortURL)
				return
			}
		}

		// Generate a short URL
		shortURL := generator.ShortURL(string(longURL), store.Store.GetStore())

		// Save the URL
		store.Store.Save(shortURL, string(longURL), "", opts)

		// save to db if exists
		if dbExists {
			store.SaveToDB(shortURL, string(longURL), "")
		}

		// Return the short URL
		w.WriteHeader(http.StatusCreated)
		_, _ = fmt.Fprintf(w, "%s/%s", opts.BaseURL, shortURL)
	}
}

func RedirectToURL(opts *config.Options) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		shortURL := vars["shortURL"]
		// Prio to DB
		// Look up for the long URL in the DB
		if opts.ConnectionString != "" {
			longURL, ok := store.ReadFromDB(shortURL)
			if !ok {
				http.NotFound(w, r)
				return
			}
			// Redirect to the long URL
			http.Redirect(w, r, longURL, http.StatusTemporaryRedirect)
			return
		}

		// Look up the long URL in in-memory store
		if opts.ConnectionString == "" {
			longURL, ok := store.Store.Find(shortURL)
			if !ok {
				http.NotFound(w, r)
				return
			}
			// Redirect to the long URL
			http.Redirect(w, r, longURL.Value, http.StatusTemporaryRedirect)
			return
		}
	}
}

func ShortenURLFromJSON(opts *config.Options) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dbExists := opts.ConnectionString != ""
		// Read the long URL from the request body
		body, err := io.ReadAll(r.Body)
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Error().Err(err).Msg("Error closing request body")
			}
		}(r.Body)
		w.Header().Set("Content-Type", "application/json")
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
			w.WriteHeader(http.StatusConflict)
			response := ShortenURLResponse{
				ShortURL: fmt.Sprintf("%s/%s", opts.BaseURL, shortURL),
			}
			responseJSON, errMarshal := json.Marshal(response)
			if errMarshal != nil {
				log.Error().Err(errMarshal).Msg("Error marshalling response")
				return
			}
			_, _ = w.Write(responseJSON)
			return
		}
		// Check if exists in DB
		if dbExists {
			if shortURL, exists := store.CheckIfExistsInDB(request.LongURL); exists {
				w.WriteHeader(http.StatusConflict)
				response := ShortenURLResponse{
					ShortURL: fmt.Sprintf("%s/%s", opts.BaseURL, shortURL),
				}
				responseJSON, errMarshal := json.Marshal(response)
				if errMarshal != nil {
					log.Error().Err(errMarshal).Msg("Error marshalling response")
					return
				}
				_, _ = w.Write(responseJSON)
				return
			}
		}

		// Generate a short URL
		shortURL := generator.ShortURL(request.LongURL, store.Store.GetStore())

		// Save the URL
		store.Store.Save(shortURL, request.LongURL, "", opts)

		// DB store exists
		if dbExists {
			store.SaveToDB(shortURL, request.LongURL, "")
		}

		// Return the short URL
		w.WriteHeader(http.StatusCreated)
		response := ShortenURLResponse{
			ShortURL: fmt.Sprintf("%s/%s", opts.BaseURL, shortURL),
		}
		responseJSON, _ := json.Marshal(response)
		_, _ = w.Write(responseJSON)
	}
}

func Ping(w http.ResponseWriter, r *http.Request) {
	if err := store.DB.Ping(); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "PONG")
}

// BatchInsertRequest represents a batch insert request
type BatchInsertRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// BatchInsertResponse represents a batch insert response
type BatchInsertResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// BatchInsert handles batch insert requests
// No checks for collisions are done nor they are requested
func BatchInsert(opts *config.Options) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Decode the JSON request body
		var requests []BatchInsertRequest
		err := json.NewDecoder(r.Body).Decode(&requests)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Convert the requests to URLRecords
		var records []store.BatchValues
		var responses []BatchInsertResponse
		for _, req := range requests {
			// Generate a UUID for each record
			id, err := uuid.NewRandom()
			if err != nil {
				http.Error(w, "Failed to generate UUID", http.StatusInternalServerError)
				return
			}

			// Generate a short URL
			shortURL := generator.ShortURLWithoutCheck(req.OriginalURL)

			record := store.BatchValues{
				UUID:        id.String(),
				ShortURL:    shortURL,
				OriginalURL: req.OriginalURL,
			}
			records = append(records, record)

			// Create a response object
			response := BatchInsertResponse{
				CorrelationID: req.CorrelationID,
				ShortURL:      fmt.Sprintf("%s/%s", opts.BaseURL, shortURL),
			}
			responses = append(responses, response)
		}

		// Save the URLs to DB
		store.BatchSave(records)

		// Set the response content type to JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		// Encode the responses as JSON and write to the response writer
		err = json.NewEncoder(w).Encode(responses)
		if err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}
