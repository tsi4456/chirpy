package auth

import (
	"crypto/rand"
	"encoding/hex"
)

func MakeRefreshToken() (string, error) {
	randString := make([]byte, 32)
	_, err := rand.Read(randString)
	if err != nil {
		return "", err
	}
	randHex := hex.EncodeToString(randString)
	return randHex, nil
}
