package middlewares

import (
	"compress/gzip"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"strings"
	"time"
)

// Custom response writer to capture status code and response size
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(data)
	rw.size += size
	return size, err
}

// LoggingMiddleware is a middleware function that logs request details along with response status code, response size, and time taken.
//
// It takes a http.Handler as a parameter and returns a http.Handler.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom ResponseWriter to capture response status code and size
		rw := &responseWriter{w, http.StatusOK, 0}

		// Call the next handler in the chain
		next.ServeHTTP(rw, r)

		// Calculate the time taken
		elapsed := time.Since(start)

		// Log request details along with response status code and response size
		log.Info().Msgf("Request URI: %s, Method: %s, Status Code: %d, Response Size: %d bytes, Time taken: %s", r.RequestURI, r.Method, rw.statusCode, rw.size, elapsed)
	})
}

// GzipMiddleware is a middleware that decompresses gzip-encoded requests
func GzipAcceptMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			gzipReader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Error creating gzip reader", http.StatusInternalServerError)
				return
			}
			defer gzipReader.Close()

			// Replace the request body with the decompressed reader
			r.Body = gzipReader
		}

		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// GzipMiddleware is a middleware that compresses the response if the client supports gzip encoding
func GzipSendMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			// If gzip is not supported, serve the content as is
			next.ServeHTTP(w, r)
			return
		}

		// Set the Content-Encoding header
		w.Header().Set("Content-Encoding", "gzip")

		// Create a gzip writer
		gz := gzip.NewWriter(w)
		defer gz.Close()

		// Create a gzipResponseWriter that wraps the gzip writer
		grw := gzipResponseWriter{Writer: gz, ResponseWriter: w}

		// Call the next handler with the gzipResponseWriter
		next.ServeHTTP(grw, r)
	})
}
