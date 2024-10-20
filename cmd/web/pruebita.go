package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Initialize store for shortened URLs
var store = map[string]string{}
var mutex = &sync.Mutex{}

// Base62 characters for generating the shortened URL
var base62Chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func init() {
	rand.Seed(time.Now().UnixNano()) // Seed random number generator
}

// Handler to shorten a URL
func shortenURL(w http.ResponseWriter, r *http.Request) {
	// Parse JSON request body
	type RequestBody struct {
		URL string `json:"url"`
	}
	var body RequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	originalURL := body.URL
	if !strings.HasPrefix(originalURL, "http://") && !strings.HasPrefix(originalURL, "https://") {
		http.Error(w, "URL must start with http:// or https://", http.StatusBadRequest)
		return
	}

	// Generate short code
	shortCode := generateShortCode()

	// Store URL in the map with short code as the key
	mutex.Lock()
	store[shortCode] = originalURL
	mutex.Unlock()

	// Return shortened URL in response
	shortenedURL := fmt.Sprintf("http://localhost:8080/%s", shortCode)
	jsonResponse(w, http.StatusOK, map[string]string{"shortened_url": shortenedURL})
}

// Handler to redirect to original URL
func redirectURL(w http.ResponseWriter, r *http.Request) {
	shortCode := strings.TrimPrefix(r.URL.Path, "/")

	mutex.Lock()
	originalURL, exists := store[shortCode]
	mutex.Unlock()

	if !exists {
		http.Error(w, "Shortened URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

// Helper function to send JSON responses
func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func main2() {
	http.HandleFunc("/shorten", shortenURL)
	http.HandleFunc("/", redirectURL)

	fmt.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
