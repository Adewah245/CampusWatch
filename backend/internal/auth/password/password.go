package password

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword securely hashes a user password using bcrypt.
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password is required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("generate password hash: %w", err)
	}

	return string(hash), nil
}

// VerifyPassword checks that a plain password matches a stored bcrypt hash.
func VerifyPassword(hash, password string) error {
	if hash == "" {
		return errors.New("hash is required")
	}
	if password == "" {
		return errors.New("password is required")
	}

	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
