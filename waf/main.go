package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	targetURL, err := url.Parse("http://backend-app:8080")
	if err != nil {
		log.Fatalf("Failed to parse target URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[WAF LOG] Incoming request: %s %s", r.Method, r.URL.Path)

		// todo

		proxy.ServeHTTP(w, r)
	})

	log.Println("🚀 WAF Reverse Proxy running on http://localhost:9000")
	log.Printf("👉 Forwarding requests to: %s\n\n", targetURL)

	if err := http.ListenAndServe(":9000", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}