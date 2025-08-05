package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/tsi4456/chirpy/internal/auth"
	"github.com/tsi4456/chirpy/internal/database"
)

type loginDetails struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (cfg *apiConfig) handleUsers(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := loginDetails{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}

	hashPW, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}
	resp, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{Email: params.Email, HashedPassword: hashPW})
	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}
	respondWithJSON(w, 201, User{Id: resp.ID.String(), CreatedAt: resp.CreatedAt.Time.String(), UpdatedAt: resp.UpdatedAt.Time.String(), Email: resp.Email, IsChirpyRed: resp.IsChirpyRed})
}

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	type loginResponse struct {
		ID           uuid.UUID `json:"user_id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		IsChirpyRed  bool      `json:"is_chirpy_red"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := loginDetails{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}

	resp, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}
	err = auth.CheckPasswordHash(params.Password, resp.HashedPassword)
	if err != nil {
		respondWithError(w, 401, "Password does not match")
		return
	}
	token, err := auth.MakeJWT(resp.ID, cfg.secret)
	if err != nil {
		respondWithError(w, 401, "Could not create authorization token")
		return
	}

	refresh_token, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, 401, "Could not create authorization token")
		return
	}
	cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{Token: refresh_token, UserID: resp.ID})

	respondWithJSON(w, 200, loginResponse{ID: resp.ID, CreatedAt: resp.CreatedAt.Time, UpdatedAt: resp.UpdatedAt.Time, Email: resp.Email, IsChirpyRed: resp.IsChirpyRed, Token: token, RefreshToken: refresh_token})
}

func (cfg *apiConfig) handleRefresh(w http.ResponseWriter, r *http.Request) {
	type tokenResponse struct {
		Token string `json:"token"`
	}

	tokenString, err := getHeaderToken(r.Header)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	refreshToken, err := cfg.db.GetRefreshToken(r.Context(), tokenString)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	} else if refreshToken.RevokedAt.Valid || refreshToken.ExpiresAt.Before(time.Now()) {
		respondWithError(w, 401, "Token has expired")
		return
	}

	token, err := auth.MakeJWT(refreshToken.UserID, cfg.secret)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}
	respondWithJSON(w, 200, tokenResponse{Token: token})
}

func (cfg *apiConfig) handleRevoke(w http.ResponseWriter, r *http.Request) {
	tokenString, err := getHeaderToken(r.Header)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}
	cfg.db.RevokeToken(r.Context(), tokenString)
	respondWithJSON(w, 204, nil)
}

func (cfg *apiConfig) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	tokenString, err := getHeaderToken(r.Header)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	authorizedID, err := auth.ValidateJWT(tokenString, cfg.secret)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := loginDetails{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}

	hashPW, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}
	resp, err := cfg.db.UpdateUserPassword(r.Context(), database.UpdateUserPasswordParams{ID: authorizedID, Email: params.Email, HashedPassword: hashPW})
	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}
	respondWithJSON(w, 200, User{Id: resp.ID.String(), CreatedAt: resp.CreatedAt.Time.String(), UpdatedAt: resp.UpdatedAt.Time.String(), Email: resp.Email})
}
