package main

import (
	"fmt"
	"sync/atomic"
	"log"
	"net/http"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	isOk bool
}

func (cfg *apiConfig) middleWareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	
	fmt.Fprintf(w, "<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileServerHits.Load())
}


func (cfg *apiConfig) resetMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	cfg.fileServerHits.Store(0)
}

func (cfg *apiConfig) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if cfg.isOk {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	} else {
		w.WriteHeader(503)
		fmt.Fprintf(w, "Not OK")
	}
}

func (cfg *apiConfig) validateChirp(w http.ResponseWriter, r *http.Request) {
	type content struct {
		body string 'json:"body"'
	}

	contentVal := content{}

	dat, err := json.Marshal(contentVal)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", dat)
}

func main() {
	servMux := http.NewServeMux()
	
	cfg := &apiConfig{}
	cfg.isOk = true

	fileServer := http.FileServer(http.Dir("."))

	servMux.Handle("/app/", http.StripPrefix("/app/", cfg.middleWareMetricsInc(fileServer)))
	servMux.Handle("/app/assets", http.StripPrefix("/app/", cfg.middleWareMetricsInc(fileServer)))
	servMux.HandleFunc("POST /api/validate_chirp" cfg.validateChirp)
	servMux.HandleFunc("GET /admin/metrics", cfg.handleMetrics)
	servMux.HandleFunc("POST /admin/reset", cfg.resetMetrics)
	servMux.HandleFunc("GET /api/healthz", cfg.healthHandler)
	
	
	fmt.Printf("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", servMux))
}
