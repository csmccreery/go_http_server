package main

import (
	"fmt"
	"sync/atomic"
	"log"
	"net/http"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

func (cfg *apiConfig) middleWareMetricsInc(next http.Handler) http.Handler {
	cfg.fileServerHits.Add(1)
	return next
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func main() {
	servMux := http.NewServeMux()
	servMux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir("."))))
	servMux.Handle("/app/assets/", http.FileServer(http.Dir(".")))

	servMux.HandleFunc("/healthz", healthHandler)
	
	fmt.Printf("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", servMux))
}
