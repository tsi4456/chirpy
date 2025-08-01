package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/tsi4456/chirpy/internal/database"
)

func main() {
	godotenv.Load()

	db, err := sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		fmt.Println(err)
	}

	var cfg apiConfig
	cfg.db = database.New(db)
	cfg.env = os.Getenv("PLATFORM")
	cfg.secret = os.Getenv("SECRET")

	sm := http.NewServeMux()

	server := http.Server{Addr: ":8080", Handler: sm}

	handler := cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	sm.Handle("/app/", handler)

	sm.HandleFunc("GET /api/healthz", handleHealth)
	sm.HandleFunc("GET /admin/metrics", cfg.handleMetric)
	sm.HandleFunc("POST /admin/reset", cfg.handleReset)
	sm.HandleFunc("POST /api/users", cfg.handleUsers)
	sm.HandleFunc("PUT /api/users", cfg.handleUpdateUser)
	sm.HandleFunc("POST /api/login", cfg.handleLogin)
	sm.HandleFunc("POST /api/refresh", cfg.handleRefresh)
	sm.HandleFunc("POST /api/revoke", cfg.handleRevoke)
	sm.HandleFunc("POST /api/chirps", cfg.handlePostChirps)
	sm.HandleFunc("GET /api/chirps", cfg.handleGetChirps)
	sm.HandleFunc("GET /api/chirps/{chirpID}", cfg.handleGetChirp)

	server.ListenAndServe()
}
