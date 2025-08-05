package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/tsi4456/chirpy/internal/auth"
)

func (cfg *apiConfig) handlePolkaWebhooks(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	headerKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	if headerKey != cfg.polka_key {
		respondWithJSON(w, 401, nil)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}

	if params.Event != "user.upgraded" {
		respondWithJSON(w, 204, nil)
		return
	}

	userID, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}

	success, err := cfg.db.UpgradeUser(r.Context(), userID)
	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}
	if success == uuid.Nil {
		respondWithJSON(w, 404, nil)
		return
	}

	respondWithJSON(w, 204, nil)
}
