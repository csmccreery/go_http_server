package main

import _ "github.com/lib/pq"

import (
	"github.com/go_http_server/internal/database"
	"github.com/go_http_server/internal/server"
	"github.com/joho/godotenv"
	"os"
	"fmt"
	"database/sql"
	"log"
	"net/http"
)

func Run() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("Failed to load .env file: %w", err)
	}
	
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("Failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("Failed to ping database: %w", err)
	}
	defer db.Close()
	
	dbQueries := database.New(db)
	servMux := http.NewServeMux()
	
	cfg := &server.ApiConfig{}
	cfg.Ok = true
	cfg.Queries = dbQueries

	fileServer := http.FileServer(http.Dir("."))

	servMux.Handle("/app/", http.StripPrefix("/app/", cfg.MiddleWareMetricsInc(fileServer)))
	servMux.Handle("/app/assets", http.StripPrefix("/app/", cfg.MiddleWareMetricsInc(fileServer)))
	
	servMux.HandleFunc("POST /api/users", cfg.CreateUser)
	servMux.HandleFunc("POST /api/login", cfg.Login)
	servMux.HandleFunc("POST /api/chirps", cfg.Chirp)
	servMux.HandleFunc("GET /api/chirps", cfg.GetChirps)
	servMux.HandleFunc("GET /api/chirps/{chirpID}", cfg.GetChirp)
	servMux.HandleFunc("GET /api/userss/{userID}", cfg.GetUser)
	servMux.HandleFunc("GET /api/users", cfg.GetUsers)
	servMux.HandleFunc("GET /admin/metrics", cfg.HandleMetrics)
	servMux.HandleFunc("POST /admin/reset", cfg.ResetMetrics)
	servMux.HandleFunc("GET /api/healthz", cfg.HealthHandler)
	
	
	fmt.Printf("Listening on port 8080")
	return http.ListenAndServe(":8080", servMux)
}
	
func main() {
	if err := Run(); err != nil {
		log.Printf("Application encountered a fatal error: %v", err)
		os.Exit(1)
	}
}
