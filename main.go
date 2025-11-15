package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

var hopByHop = map[string]bool{
	"Connection":          true,
	"Keep-Alive":          true,
	"Proxy-Authenticate":  true,
	"Proxy-Authorization": true,
	"TE":                  true,
	"Trailers":            true,
	"Transfer-Encoding":   true,
	"Upgrade":             true,
}

func copyHeaders(dst http.Header, src http.Header) {
	for k, vals := range src {
		if hopByHop[k] {
			continue
		}
		// Header keys are case‑insensitive, http.Header normalises them.
		for _, v := range vals {
			dst.Add(k, v)
		}
	}
}

func fetchURL(w http.ResponseWriter, r *http.Request) {
	urlParam := r.URL.Query().Get("url")
	if urlParam == "" {
		http.Error(w, "URL parameter is missing", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(urlParam, "http://") && !strings.HasPrefix(urlParam, "https://") {
		urlParam = "https://" + urlParam
	}

	fullURL := urlParam
	queryParams := r.URL.Query()
	queryParams.Del("url")

	if len(queryParams) > 0 {
		qs := queryParams.Encode()
		if qs != "" {
			if strings.Contains(fullURL, "?") {
				fullURL += "&" + qs
			} else {
				fullURL += "?" + qs
			}
		}
	}

	outReq, err := http.NewRequest(r.Method, fullURL, r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create request: %v", err), http.StatusInternalServerError)
		return
	}

	copyHeaders(outReq.Header, r.Header)

	resp, err := http.DefaultClient.Do(outReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch URL: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Log non-200 responses
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Non-200 response: %d, body: %s", resp.StatusCode, string(body))
		w.WriteHeader(resp.StatusCode)
		copyHeaders(w.Header(), resp.Header)
		w.Write(body)
		return
	}

	w.WriteHeader(resp.StatusCode)
	copyHeaders(w.Header(), resp.Header)

	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("Error streaming response body: %v", err)
	}
}

func main() {
	http.HandleFunc("/fetch", fetchURL)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
