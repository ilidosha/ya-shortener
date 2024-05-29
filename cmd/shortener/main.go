// main.go
package main

import (
	"database/sql"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq" // Anonymous import for PostgreSQL driver
	"github.com/rs/zerolog/log"
	"net"
	"net/http"
	"shortener/internal/config"
	"shortener/internal/handlers"
	"shortener/internal/logger"
	"shortener/internal/middlewares"
	"shortener/internal/store"
	"time"
)

func main() {
	opts, err := config.ParseOptions()
	if err != nil {
		log.Fatal().Err(err)
	}
	// Setup log with debug level
	logger.SetupLog(true)

	fileStoreExists := opts.FileStore != ""
	dbExists := opts.ConnectionString != ""

	// Initialize the in-memory store
	if !dbExists {
		store.New()
	}

	// Initialize the file store if exists
	if fileStoreExists {
		// Load from file store if exists
		errLoad := store.Store.LoadFromFile(opts.FileStore)
		if errLoad != nil {
			log.Info().Msgf("Failed to load from file store: %s", errLoad)
		}
	}
	// Initialize the database store if exists
	if dbExists {
		// Open the database connection
		store.DB, err = sql.Open("postgres", opts.ConnectionString)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to open database connection")
		}
		defer store.DB.Close()
		// Initialize the database
		err = store.InitDB(store.DB)
		if err != nil {
			log.Fatal().Msgf("Failed to initialize database: %v", err)
		}
	}

	r := mux.NewRouter()
	// Middlewares
	r.Use(middlewares.LoggingMiddleware)
	r.Use(middlewares.GzipAcceptMiddleware)
	r.Use(middlewares.GzipSendMiddleware)
	// Handlers
	r.HandleFunc("/", handlers.ShortenURL(opts)).Methods("POST")
	r.HandleFunc("/api/shorten", handlers.ShortenURLFromJSON(opts)).Methods("POST")
	r.HandleFunc("/ping", handlers.Ping).Methods("GET")
	r.HandleFunc("/{shortURL}", handlers.RedirectToURL(opts)).Methods("GET")
	r.HandleFunc("/api/shorten/batch", handlers.BatchInsert(opts)).Methods("POST")
	r.HandleFunc("/api/user/urls", handlers.GetAllURLsForUser(opts)).Methods("GET")
	r.HandleFunc("/api/user/urls", handlers.DeleteFromURLs(opts)).Methods("DELETE")

	// Initialize the hard deleter if db exists
	if dbExists {
		// Starting the deleter in a separate goroutine
		go func() {
			// Hard delete function
			// Define the function to run every 30 seconds
			actualDeletingFunction := func() {
				store.HardDeleteRecord()
			}

			// Run the hard delete function every 30 seconds
			ticker := time.NewTicker(200 * time.Millisecond)
			defer ticker.Stop()
			for range ticker.C {
				actualDeletingFunction()
			}

		}()
	}

	log.Info().Msgf("Starting server on %s\n", opts.ServerAddress)
	serv := http.Server{
		Addr:    opts.ServerAddress,
		Handler: r,
	}
	// Try to start the server
	err = serv.ListenAndServe()
	if err != nil {
		// Check if the error is due to the port being in use
		if _, ok := err.(*net.OpError); ok && err.(*net.OpError).Op == "listen" {
			log.Error().Err(err).Msgf("Error: Address %s is already in use\n", opts.ServerAddress)
		} else {
			log.Error().Err(err).Msgf("Error starting server: %s\n", err)
		}
		return
	}
}
