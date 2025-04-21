package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func fetchURL(w http.ResponseWriter, r *http.Request) {
	// Get the URL path from the query parameter "url"
	urlParam := r.URL.Query().Get("url")
	if urlParam == "" {
		http.Error(w, "URL parameter is missing", http.StatusBadRequest)
		return
	}

	// Ensure the URL has a scheme (http or https) if not provided
	if !strings.HasPrefix(urlParam, "http://") && !strings.HasPrefix(urlParam, "https://") {
		urlParam = "https://" + urlParam
	}

	// Build the full URL with query parameters
	fullURL := urlParam
	queryParams := r.URL.Query()
	queryParams.Del("url") // Remove the 'url' parameter to avoid duplication
	if len(queryParams) > 0 {
		queryString := queryParams.Encode()
		if queryString != "" {
			if strings.Contains(fullURL, "?") {
				fullURL += "&" + queryString
			} else {
				fullURL += "?" + queryString
			}
		}
	}

	// Fetch the URL
	resp, err := http.Get(fullURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch URL: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read response: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write(body)
}

func main() {
	http.HandleFunc("/fetch", fetchURL)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
