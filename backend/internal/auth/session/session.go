package session

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
)

// GenerateSessionToken creates a cryptographically random session token.
func GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

// ValidateSessionToken ensures the token is not empty.
func ValidateSessionToken(token string) error {
	if token == "" {
		return errors.New("session token is required")
	}

	return nil
}
