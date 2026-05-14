package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// version is overridden at build time:
//
//	go build -ldflags="-X main.version=1.2.3" .
var version = "dev"

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/health", healthHandler)

	addr := ":8080"
	fmt.Printf("listening on %s  version=%s\n", addr, version)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()
	fmt.Fprintf(w, "hostname=%s  version=%s\n", hostname, version)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}
