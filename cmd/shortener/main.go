// main.go
package main

import (
	"database/sql"
	"fmt"
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
)

func main() {
	opts, err := config.ParseOptions()
	if err != nil {
		log.Fatal().Err(err)
	}

	// Initialize the in-memory store
	store.New()
	// Initialize the file store if exists
	if opts.FileStore != "" {
		// Load from file store if exists
		errLoad := store.Store.LoadFromFile(opts.FileStore)
		if errLoad != nil {
			log.Info().Msgf("Failed to load from file store: %s", errLoad)
		}
	}
	// Initialize the database store if exists
	if opts.ConnectionString != "" {
		// Open the database connection
		store.DB, err = sql.Open("postgres", opts.ConnectionString)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to open database connection")
		}
		defer store.DB.Close()
		// Initialize the database
		errInit := store.InitDB(store.DB)
		if errInit != nil {
			log.Fatal().Msgf("Failed to initialize database: %v", err)
		}
	}
	// Setup log with debug level
	logger.SetupLog(true)

	r := mux.NewRouter()
	// Middlewares
	r.Use(middlewares.LoggingMiddleware)
	r.Use(middlewares.GzipAcceptMiddleware)
	r.Use(middlewares.GzipSendMiddleware)
	// Handlers
	r.HandleFunc("/", handlers.ShortenURL(opts)).Methods("POST")
	r.HandleFunc("/api/shorten", handlers.ShortenURLFromJSON(opts)).Methods("POST")
	r.HandleFunc("/{shortURL}", handlers.RedirectToURL(opts)).Methods("GET")
	r.HandleFunc("/ping", handlers.Ping(opts)).Methods("GET")
	r.HandleFunc("/api/shorten/batch", handlers.BatchInsert(opts)).Methods("POST")

	// Start the server
	fmt.Printf("Starting server on %s\n", opts.ServerAddress)
	serv := http.Server{
		Addr:    opts.ServerAddress,
		Handler: r,
	}
	// Try to start the server
	err = serv.ListenAndServe()
	if err != nil {
		// Check if the error is due to the port being in use
		if _, ok := err.(*net.OpError); ok && err.(*net.OpError).Op == "listen" {
			fmt.Printf("Error: Address %s is already in use\n", opts.ServerAddress)
		} else {
			fmt.Printf("Error starting server: %s\n", err)
		}
		return
	}
}
