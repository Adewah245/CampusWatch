package auth

import "testing"

func TestHashPasswordAndVerifyPassword(t *testing.T) {
	password := "StrongPass123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if hash == "" {
		t.Fatal("expected hash to be generated")
	}
	if hash == password {
		t.Fatal("expected password hash to differ from plain password")
	}

	if err := VerifyPassword(hash, password); err != nil {
		t.Fatalf("verify password: %v", err)
	}

	if err := VerifyPassword(hash, "wrong-password"); err == nil {
		t.Fatal("expected verification to fail for wrong password")
	}
}

func TestGenerateSessionToken(t *testing.T) {
	token, err := GenerateSessionToken()
	if err != nil {
		t.Fatalf("generate session token: %v", err)
	}
	if token == "" {
		t.Fatal("expected session token to be generated")
	}
	if len(token) < 32 {
		t.Fatalf("expected session token to be at least 32 chars, got %d", len(token))
	}
}
