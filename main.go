package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

type item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var (
	mu    sync.Mutex
	items = []item{{ID: 1, Name: "first item"}}
)

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"service":   "sandbox-go-basic",
			"language":  "go",
			"go":        runtime.Version(),
			"endpoints": []string{"/health", "/api/items", "/api/echo?msg=hi", "/api/time"},
		})
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/api/time", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"now": time.Now().UTC().Format(time.RFC3339)})
	})
	mux.HandleFunc("/api/echo", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"echo": r.URL.Query().Get("msg")})
	})
	mux.HandleFunc("/api/items", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			mu.Lock()
			defer mu.Unlock()
			writeJSON(w, http.StatusOK, items)
		case http.MethodPost:
			var body struct {
				Name string `json:"name"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
				return
			}
			mu.Lock()
			defer mu.Unlock()
			it := item{ID: len(items) + 1, Name: body.Name}
			items = append(items, it)
			writeJSON(w, http.StatusCreated, it)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		}
	})

	log.Printf("Listening on %s", port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, mux))
}
