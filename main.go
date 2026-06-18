package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// version is overridden at build time:
//
//	go build -ldflags="-X main.version=1.2.3" .
var version = "dev"

var (
	broken atomic.Bool
	dataMu sync.Mutex
)

const dataFile = "/data/messages.txt"

func main() {
	appVersion := os.Getenv("APP_VERSION")
	if appVersion == "" {
		appVersion = version
	}
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}
	greeting := os.Getenv("GREETING")
	if greeting == "" {
		greeting = "Hello"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hostname, _ := os.Hostname()
		fmt.Fprintf(w, "greeting=%s  hostname=%s  version=%s  env=%s\n", greeting, hostname, appVersion, appEnv)
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if broken.Load() {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "error"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"time":   time.Now().UTC().Format(time.RFC3339),
		})
	})
	mux.HandleFunc("/break", func(w http.ResponseWriter, r *http.Request) {
		broken.Store(true)
		go func() {
			time.Sleep(60 * time.Second)
			broken.Store(false)
		}()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "breaking"})
	})
	mux.HandleFunc("/write", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		msg := r.URL.Query().Get("msg")
		if msg == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "msg query param required"})
			return
		}

		dataMu.Lock()
		defer dataMu.Unlock()

		if err := os.MkdirAll("/data", 0755); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		f, err := os.OpenFile(dataFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		fmt.Fprintln(f, msg)
		f.Close()

		count, _ := countLines(dataFile)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"count": count})
	})
	mux.HandleFunc("/read", func(w http.ResponseWriter, r *http.Request) {
		dataMu.Lock()
		defer dataMu.Unlock()

		data, err := os.ReadFile(dataFile)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprint(w, "")
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.Write(data)
	})

	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		fmt.Fprintln(w, "slow response after 3s")
	})

	mux.HandleFunc("/external", func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get("https://httpbin.org/get")
		if err != nil {
			http.Error(w, fmt.Sprintf("external call failed: %v", err), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		w.Header().Set("Content-Type", "application/json")
		io.Copy(w, resp.Body)
	})

	addr := ":8080"
	fmt.Printf("listening on %s  version=%s  env=%s  greeting=%s\n", addr, appVersion, appEnv, greeting)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func countLines(filename string) (int, error) {
	f, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			count++
		}
	}
	return count, scanner.Err()
}
