package auth

import (
	"errors"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type TokenType string

const (
	TokenTypeAccess TokenType = "chirpy-access"
)

func HashPassword(password string) (string, error) {
	hashPW, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", err
	}
	return string(hashPW), nil
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("Authorization token not found")
	}
	tokenString, valid := strings.CutPrefix(authHeader, "Bearer")
	if !valid {
		return "", errors.New("Authorization token not found")
	}
	tokenString = strings.TrimSpace(tokenString)
	return tokenString, nil
}

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("Authorization token not found")
	}
	tokenString, valid := strings.CutPrefix(authHeader, "ApiKey")
	if !valid {
		return "", errors.New("Authorization token not found")
	}
	tokenString = strings.TrimSpace(tokenString)
	return tokenString, nil
}
