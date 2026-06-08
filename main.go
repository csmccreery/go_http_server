package main

import (
	"strings"
	"time"
	"fmt"
	"sync/atomic"
	"log"
	"net/http"
	"encoding/json"
	"github.com/satori/uuid.go"
)

type ApiConfig struct {
	fileServerHits atomic.Int32
	Ok bool
}

type Content struct {
	Body string
}

type Chirp struct {
	U1 string
	Length int
	Timestamp time.Time
	Content Content
}

func generateUUID() (string) {
	u1 := uuid.NewV4()
	return u1.String()
}

func cleanProfaneWords(chirp string, profaneMap map[string]bool) string {
	fields := strings.Fields(chirp)
	for i:=0; i<len(fields); i++ {
		word := strings.ToLower(fields[i])
		_, ok := profaneMap[word]
		if ok {
			fields[i] = "****"
		}
	}

	return strings.Join(fields, " ")
}

func (cfg *ApiConfig) middleWareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *ApiConfig) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	
	fmt.Fprintf(w, "<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileServerHits.Load())
}


func (cfg *ApiConfig) resetMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	cfg.fileServerHits.Store(0)
}

func (cfg *ApiConfig) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if cfg.Ok {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	} else {
		w.WriteHeader(503)
		fmt.Fprintf(w, "Not OK")
	}
}

func (cfg *ApiConfig) respondWithJSON(w http.ResponseWriter, code int, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal json", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, string(data))
}

func (cfg *ApiConfig) respondWithError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(400)
}

func (cfg *ApiConfig) validateChirp(w http.ResponseWriter, r *http.Request) {
	content := Content{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&content)
	if err != nil {
		log.Printf("Error decoding chirp content: %s", err)
		w.WriteHeader(500)
		return
	}

	if (len(content.Body) > 140) {
		cfg.respondWithError(w, 400, "Chirp is too long")
	} else {
		profaneMap := map[string]bool{
			"kerfuffle": true,
			"sharbert": true,
			"fornax": true,
		}
		cleanedBody := cleanProfaneWords(content.Body, profaneMap)
		
		type Payload struct {
			Valid bool `json:"valid"`
			Cleaned string `json:"cleaned_body"`
		}

		payload := Payload{Valid: true, Cleaned: cleanedBody}
		cfg.respondWithJSON(w, 200, payload)
	}
}
	
func main() {
	servMux := http.NewServeMux()
	
	cfg := &ApiConfig{}
	cfg.Ok = true

	fileServer := http.FileServer(http.Dir("."))

	servMux.Handle("/app/", http.StripPrefix("/app/", cfg.middleWareMetricsInc(fileServer)))
	servMux.Handle("/app/assets", http.StripPrefix("/app/", cfg.middleWareMetricsInc(fileServer)))
	servMux.HandleFunc("POST /api/validate_chirp", cfg.validateChirp)
	servMux.HandleFunc("GET /admin/metrics", cfg.handleMetrics)
	servMux.HandleFunc("POST /admin/reset", cfg.resetMetrics)
	servMux.HandleFunc("GET /api/healthz", cfg.healthHandler)
	
	
	fmt.Printf("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", servMux))
}
