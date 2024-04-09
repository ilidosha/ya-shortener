// main.go
package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net"
	"net/http"
	"shortener/internal/config"
	"shortener/internal/handlers"
)

func main() {
	opts, err := config.ParseOptions()
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/", handlers.ShortenURL(opts)).Methods("POST")
	r.HandleFunc("/{shortURL}", handlers.RedirectToURL).Methods("GET")

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
