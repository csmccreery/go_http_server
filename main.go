package main

import _ "github.com/lib/pq"

import (
	"strings"
	"github.com/go_http_server/internal/database"
	"github.com/go_http_server/internal/auth"
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
	Passwor string `json:"password"`
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

func (cfg *ApiConfig) respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintf(w, "Error: %s: %v", msg, err)
}

func (cfg *ApiConfig) CreateUser(w http.ResponseWriter, r *http.Request) {
	user := User{}

	type Content struct {
		Email string `json:"email"`
	}

	payload := Content{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&payload)
	if err != nil {
		log.Printf("Error decoding user creation request: %s", err)
		w.WriteHeader(500)
		return
	}

	user.Email = payload.Email
	newUser, err := cfg.Queries.CreateUser(r.Context(), payload.Email)
	if err != nil {
		cfg.respondWithError(w, 400, "Failed to create new user in database", err)
	}

	type UserResponse struct {
		ID uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email string `json:"email"`
	}

	resp := UserResponse{
		ID: newUser.ID,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
		Email: newUser.Email,
	}
	
	cfg.respondWithJSON(w, 201, resp)

}

func (cfg *ApiConfig) GetChirps(w http.ResponseWriter, r *http.Request) {
	allChirps, err := cfg.Queries.GetChirps(r.Context())
	if err != nil {
		cfg.respondWithError(w, 400, "Failed to get users from DB", err)
		return
	}

	type ChirpResponse struct {
		ID uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}


	respChirps := []ChirpResponse{}
	
	for _, chirp := range allChirps {
		resp := ChirpResponse{
			ID: chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body: chirp.Body,
			UserID: chirp.UserID,
		}
		
		respChirps = append(respChirps, resp)
	}

	cfg.respondWithJSON(w, 200, respChirps)
}

func (cfg *ApiConfig) GetChirp(w http.ResponseWriter, r *http.Request) {
	rawChirpId := r.PathValue("chirpID")
	chirpId, err := uuid.Parse(rawChirpId)
	if err != nil {
		cfg.respondWithError(w, 404, "Invalid UUID", err)
		return
	}
	
	chirp, err := cfg.Queries.GetChirp(r.Context(), chirpId)
	if err != nil {
		cfg.respondWithError(w, 404, "Failed to retrieve chirp", err)
		return
	}

	type ChirpResponse struct {
		ID uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	resp := ChirpResponse{
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body,
		UserID: chirp.UserID,
	}

	cfg.respondWithJSON(w, 200, resp)
}

func (cfg *ApiConfig) GetUsers(w http.ResponseWriter, r *http.Request) {
	allUsers, err := cfg.Queries.GetUsers(r.Context())
	if err != nil {
		cfg.respondWithError(w, 400, "Failed to get users from DB", err)
		return
	}

	type UserResponse struct {
		ID uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email string `json:"email"`
	}

	respUsers := []UserResponse{}
	
	for _, user := range allUsers {
		resp := UserResponse{
			ID: user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email: user.Email,
		}
		
		respUsers = append(respUsers, resp)
	}

	cfg.respondWithJSON(w, 200, respUsers)
}

func (cfg *ApiConfig) GetUser(w http.ResponseWriter, r *http.Request) {
	rawUserId := r.PathValue("userID")
	userId, err := uuid.Parse(rawUserId)
	if err != nil {
		cfg.respondWithError(w, 404, "Invalid UUID", err)
		return
	}
	
	user, err := cfg.Queries.GetUser(r.Context(), userId)
	if err != nil {
		cfg.respondWithError(w, 404, "Failed to retrieve user", err)
		return
	}

	type UserResponse struct {
		ID uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email string `json:"email"`
	}

	resp := UserResponse{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	}

	cfg.respondWithJSON(w, 200, resp)
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
		UserId string `json:"user_id"`
	}

	content := Content{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&content)
	if err != nil {
		cfg.respondWithError(w, 400, "Failed to decode JSON request", err)
		return
	}

	if len(content.Body) > 140 {
		cfg.respondWithError(w, 400, "Chirp is too long", err)
		return
	}

	profaneMap := map[string]bool{
		"kerfuffle": true,
		"sharbert": true,
		"fornax": true,
	}
	cleanedBody := cleanProfaneWords(content.Body, profaneMap)

	chirp.Body = cleanedBody
	chirp.UserID = content.UserId

	user_id, err := uuid.Parse(content.UserId)
	fmt.Printf("raw user id: %q\n", content.UserId)
	if err != nil {
		cfg.respondWithError(w, 400, "Failed to extract uuid from chirp", err)
		return
	}

	params := database.CreateChirpParams{
		Body: cleanedBody,
		UserID: user_id,
	}

	newChirp, err := cfg.Queries.CreateChirp(r.Context(), params)
	if err != nil {
		cfg.respondWithError(w, 400, "Failed to update database", err)
		return
	}

	type ChirpResponse struct {
		ID uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	resp := ChirpResponse{
		ID: newChirp.ID,
		CreatedAt: newChirp.CreatedAt,
		UpdatedAt: newChirp.UpdatedAt,
		Body: newChirp.Body,
		UserID: newChirp.UserID,
	}

	cfg.respondWithJSON(w, 201, resp)
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
	servMux.HandleFunc("GET /api/chirps", cfg.GetChirps)
	servMux.HandleFunc("GET /api/chirps/{chirpID}", cfg.GetChirp)
	servMux.HandleFunc("GET /api/userss/{userID}", cfg.GetUser)
	servMux.HandleFunc("GET /api/users", cfg.GetUsers)
	servMux.HandleFunc("GET /admin/metrics", cfg.handleMetrics)
	servMux.HandleFunc("POST /admin/reset", cfg.resetMetrics)
	servMux.HandleFunc("GET /api/healthz", cfg.healthHandler)
	
	
	fmt.Printf("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", servMux))
}
