package auth_test

import (
	"testing"
	"time"

	"ecommerce/backend/internal/auth"
	"ecommerce/backend/internal/models"
)

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := auth.HashPassword("correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if !auth.CheckPassword(hash, "correct-horse-battery-staple") {
		t.Error("expected correct password to match hash")
	}
	if auth.CheckPassword(hash, "wrong-password") {
		t.Error("expected wrong password to not match hash")
	}
}

func TestGenerateAndParseToken(t *testing.T) {
	user := models.User{ID: 42, Email: "test@example.com", Role: models.RoleAdmin}

	token, err := auth.GenerateToken("test-secret", 1, user)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	claims, err := auth.ParseToken("test-secret", token)
	if err != nil {
		t.Fatalf("ParseToken returned error: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("claims.UserID = %d, want %d", claims.UserID, user.ID)
	}
	if claims.Role != models.RoleAdmin {
		t.Errorf("claims.Role = %s, want %s", claims.Role, models.RoleAdmin)
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	user := models.User{ID: 1, Email: "a@b.com", Role: models.RoleCustomer}
	token, _ := auth.GenerateToken("secret-a", 1, user)

	if _, err := auth.ParseToken("secret-b", token); err == nil {
		t.Error("expected error when parsing token with wrong secret")
	}
}

func TestParseToken_Expired(t *testing.T) {
	user := models.User{ID: 1, Email: "a@b.com", Role: models.RoleCustomer}
	// Negative expiry hours produces a token that is already expired.
	token, err := auth.GenerateToken("secret", 0, user)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}
	time.Sleep(1100 * time.Millisecond)

	if _, err := auth.ParseToken("secret", token); err == nil {
		t.Error("expected error when parsing an expired token")
	}
}
