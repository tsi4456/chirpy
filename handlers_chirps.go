package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tsi4456/chirpy/internal/auth"
	"github.com/tsi4456/chirpy/internal/database"
)

type Chirp struct {
	Id        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    string    `json:"user_id"`
}

func (cfg *apiConfig) handlePostChirps(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	token, err := getHeaderToken(r.Header)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}
	authorizedID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	msg := filterProfanity(params.Body)

	resp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{Body: msg, UserID: authorizedID})
	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}

	respondWithJSON(w, 201, Chirp{Id: resp.ID.String(), CreatedAt: resp.CreatedAt.Time, UpdatedAt: resp.UpdatedAt.Time, Body: resp.Body, UserID: resp.UserID.String()})
}

func (cfg *apiConfig) handleGetChirps(w http.ResponseWriter, r *http.Request) {
	authorQuery := r.URL.Query().Get("author_id")

	sortDirection := r.URL.Query().Get("sort")
	if !slices.Contains([]string{"", "asc", "desc"}, sortDirection) {
		respondWithError(w, 400, "Invalid sort argument")
		return
	}

	var resp []database.Chirp
	var err error

	if authorQuery != "" {
		authorID, err := uuid.Parse(authorQuery)
		if err != nil {
			respondWithError(w, 400, err.Error())
			return
		}
		resp, err = cfg.db.GetAllChirpsByID(r.Context(), authorID)
	} else {
		resp, err = cfg.db.GetAllChirps(r.Context())
	}
	if err != nil {
		respondWithError(w, 400, "Something went wrong")
		return
	}

	var chirps []Chirp
	for _, c := range resp {
		chirps = append(chirps, Chirp{Id: c.ID.String(), CreatedAt: c.CreatedAt.Time, UpdatedAt: c.UpdatedAt.Time, Body: c.Body, UserID: c.UserID.String()})
	}
	if sortDirection == "desc" {
		sort.Slice(chirps, func(i, j int) bool { return chirps[i].CreatedAt.After(chirps[j].CreatedAt) })
	}
	respondWithJSON(w, 200, chirps)
}

func (cfg *apiConfig) handleGetChirp(w http.ResponseWriter, r *http.Request) {
	searchID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 404, err.Error())
		return
	}
	resp, err := cfg.db.GetChirpByID(r.Context(), searchID)
	if err != nil {
		respondWithError(w, 404, err.Error())
		return
	}
	respondWithJSON(w, 200, Chirp{Id: resp.ID.String(), CreatedAt: resp.CreatedAt.Time, UpdatedAt: resp.UpdatedAt.Time, Body: resp.Body, UserID: resp.UserID.String()})
}

func filterProfanity(msg string) string {
	banned_words := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(msg, " ")
	validated_words := []string{}
	for _, word := range words {
		if slices.Contains(banned_words, strings.ToLower(word)) {
			validated_words = append(validated_words, "****")
		} else {
			validated_words = append(validated_words, word)
		}
	}
	validated_msg := strings.Join(validated_words, " ")
	return validated_msg
}

func (cfg *apiConfig) handleDeleteChirp(w http.ResponseWriter, r *http.Request) {
	token, err := getHeaderToken(r.Header)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}
	authorizedID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	searchID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 404, err.Error())
		return
	}
	resp, err := cfg.db.GetChirpByID(r.Context(), searchID)
	if err != nil {
		respondWithError(w, 404, err.Error())
		return
	}
	if resp.UserID != authorizedID {
		respondWithError(w, 403, "Unauthorized")
		return
	}
	cfg.db.DeleteChirpByID(r.Context(), searchID)
	respondWithJSON(w, 204, nil)
}
