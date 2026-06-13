package main

import _ "github.com/lib/pq"

import (
	"strings"
	"github.com/go_http_server/internal/database"
	"time"
	"os"
	"fmt"
	"database/sql"
	"sync/atomic"
	"log"
	"net/http"
	"encoding/json"
	"github.com/joho/godotenv"
	"github.com/google/uuid"
)


type ApiConfig struct {
	fileServerHits atomic.Int32
	Queries *database.Queries
	Ok bool
}

type User struct {
	ID uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email string `json:"email"`
}

type Chirp struct {
	ID string `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body string `json:"body"`
	UserID string `json:"user_id"`
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
	cfg.fileServerHits.Store(0)
	cfg.Queries.ClearUsers(r.Context())
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Sucessfully cleared users table")

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
	w.WriteHeader(code)
	fmt.Fprintf(w, string(data))
}

func (cfg *ApiConfig) respondWithError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintf(w, "Error: %s", msg)
}

func (cfg *ApiConfig) CreateUser(w http.ResponseWriter, r *http.Request) {
	user := User{}

	type content struct {
		Email string `json:"email"`
	}

	payload := content{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&payload)
	if err != nil {
		log.Printf("Error decoding user creation request: %s", err)
		w.WriteHeader(500)
		return
	}

	user.Email = payload.Email
	_, err = cfg.Queries.CreateUser(r.Context(), payload.Email)
	if err != nil {
		cfg.respondWithError(w, 400, "Failed to create new user in database")
	}
	cfg.respondWithJSON(w, 201, user)

}

func (cfg *ApiConfig) ClearUsers(w http.ResponseWriter, r *http.Request) {
	cfg.Queries.ClearUsers(r.Context())
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Sucessfully cleared users table")
}

func (cfg *ApiConfig) chirp(w http.ResponseWriter, r *http.Request) {
	chirp := Chirp{}
	type Content struct {
		Body string `json:"body"`
		User_Id string `json:"user_id"`
	}

	content := Content{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&content)
	if err != nil {
		cfg.respondWithError(w, 400, "Failed to decode JSON request")
	}

	if len(content.Body) > 140 {
		cfg.respondWithError(w, 400, "Chirp is too long")
	}

	profaneMap := map[string]bool{
		"kerfuffle": true,
		"sharbert": true,
		"fornax": true,
	}
	cleanedBody := cleanProfaneWords(content.Body, profaneMap)

	chirp.Body = cleanedBody
	chirp.UserID = content.User_Id

	user_id, err := uuid.Parse(content.User_Id)
	if err != nil {
		cfg.respondWithError(w, 400, "Failed to extract uuid from chirp")
	}

	params := database.CreateChirpParams{
		Body: cleanedBody,
		UserID: user_id,
	}

	_, err = cfg.Queries.CreateChirp(r.Context(), params)
	if err != nil {
		fmt.Printf("Error: %v", err)
		cfg.respondWithError(w, 400, "Failed to update database")
	}

	cfg.respondWithJSON(w, 201, chirp)
}
	
func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err :=sql.Open("postgres", dbURL)
	if err != nil {
		return
	}
	dbQueries := database.New(db)
		
	servMux := http.NewServeMux()
	
	cfg := &ApiConfig{}
	cfg.Ok = true
	cfg.Queries = dbQueries

	fileServer := http.FileServer(http.Dir("."))

	servMux.Handle("/app/", http.StripPrefix("/app/", cfg.middleWareMetricsInc(fileServer)))
	servMux.Handle("/app/assets", http.StripPrefix("/app/", cfg.middleWareMetricsInc(fileServer)))
	servMux.HandleFunc("POST /api/users", cfg.CreateUser)
	servMux.HandleFunc("POST /api/chirps", cfg.chirp)
	servMux.HandleFunc("GET /admin/metrics", cfg.handleMetrics)
	servMux.HandleFunc("POST /admin/reset", cfg.resetMetrics)
	servMux.HandleFunc("GET /api/healthz", cfg.healthHandler)
	
	
	fmt.Printf("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", servMux))
}
