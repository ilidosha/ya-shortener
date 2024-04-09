package store

// URLStore a map to store the URLs
var URLStore = make(map[string]string)

// URL a struct to hold the URL
type URL struct {
	LongURL string `json:"longURL"`
}
