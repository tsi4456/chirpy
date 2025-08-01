package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/tsi4456/chirpy/internal/auth"
	"github.com/tsi4456/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	env            string
	secret         string
}

type User struct {
	Id        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Email     string `json:"email"`
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type returnError struct {
		Error string `json:"error"`
	}
	ret_err := returnError{
		Error: msg,
	}
	ret, err := json.Marshal(ret_err)
	if err != nil {
		log.Printf("Error marshalling json: %s", err)
	}
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	w.Write(ret)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	type returnValid struct {
		Valid string `json:"cleaned_body"`
	}
	val, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling json: %s", err)
	}
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	w.Write(val)
}

func getHeaderToken(r http.Header) (string, error) {
	tokenString, err := auth.GetBearerToken(r)
	if err != nil {
		return "", err
	} else if tokenString == "" {
		return "", errors.New("Authorization token not found")
	}
	return tokenString, nil
}
