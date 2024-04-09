package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"shortener/internal/config"
	"testing"

	"github.com/gorilla/mux"
)

func TestShortenURL(t *testing.T) {
	// Create a new request with a long URL in the body
	longURL := "http://example.com/very/long/url"
	req, err := http.NewRequest("POST", "/", bytes.NewBufferString(longURL))
	if err != nil {
		t.Fatal(err)
	}
	baseURL := "localhost:8080"

	opts := config.Options{
		ServerAddress: baseURL,
		BaseURL:       baseURL,
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler function, passing in the mock OptionParser
	handler := http.HandlerFunc(ShortenURL(&opts))
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	// Check the response body is what we expect
	expected := baseURL + "/1" // Assuming the first short URL is "1"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestRedirectToURL(t *testing.T) {
	// Create a new request with a short URL in the path
	shortURL := "1"
	req, err := http.NewRequest("GET", "/"+shortURL, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Set the URL variables for the request
	req = mux.SetURLVars(req, map[string]string{
		"shortURL": shortURL,
	})

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler function, passing in the mock OptionParser
	handler := http.HandlerFunc(RedirectToURL)
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect
	if status := rr.Code; status != http.StatusTemporaryRedirect {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusTemporaryRedirect)
	}

	// Check the response redirect is what we expect
	expected := "http://example.com/very/long/url"
	if rr.Header().Get("Location") != expected {
		t.Errorf("handler returned unexpected redirect: got %v want %v", rr.Header().Get("Location"), expected)
	}
}
