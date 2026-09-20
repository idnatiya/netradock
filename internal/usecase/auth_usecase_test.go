package usecase

import (
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/idnatiya/netradock/internal/model"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthUseCase(t *testing.T) {
	u := NewAuthUseCase(validator.New(), "admin", "secret", "", []byte("key"), time.Hour)

	if _, _, err := u.Login(&model.LoginRequest{Username: "admin", Password: "wrong"}); err == nil {
		t.Fatal("wrong password accepted")
	}
	token, _, err := u.Login(&model.LoginRequest{Username: "admin", Password: "secret"})
	if err != nil || !u.Verify(token) {
		t.Fatalf("valid login rejected: %v", err)
	}

	exp, sig, _ := strings.Cut(token, ".")
	if strings.Count(token, ".") != 2 {
		t.Fatalf("token is not <exp>.<nonce>.<mac>: %q", token)
	}
	if u.Verify(exp + "9." + sig) {
		t.Fatal("tampered expiry accepted")
	}
	if u.Verify(exp + "." + strings.Repeat("0", len(sig))) {
		t.Fatal("tampered signature accepted")
	}
	if NewAuthUseCase(validator.New(), "admin", "secret", "", []byte("other"), time.Hour).Verify(token) {
		t.Fatal("token from another secret accepted")
	}
	// Rotating the credential must invalidate tokens already issued (no other revocation exists).
	if NewAuthUseCase(validator.New(), "admin", "rotated", "", []byte("key"), time.Hour).Verify(token) {
		t.Fatal("token survived a password change")
	}
	if NewAuthUseCase(validator.New(), "other", "secret", "", []byte("key"), time.Hour).Verify(token) {
		t.Fatal("token survived a username change")
	}
	u.Revoke(token)
	if u.Verify(token) {
		t.Fatal("revoked token accepted")
	}
	fresh, _, _ := u.Login(&model.LoginRequest{Username: "admin", Password: "secret"})
	if fresh == token {
		t.Fatal("two logins produced the same token")
	}
	if !u.Verify(fresh) {
		t.Fatal("revoking one token killed a later one")
	}

	expired := u.sign(time.Now().Add(-time.Second), "nonce")
	if u.Verify(expired) {
		t.Fatal("expired token accepted")
	}
}

func TestAuthUseCaseBcrypt(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	u := NewAuthUseCase(validator.New(), "admin", "", string(hash), []byte("key"), time.Hour)

	if _, _, err := u.Login(&model.LoginRequest{Username: "admin", Password: "wrong"}); err == nil {
		t.Fatal("wrong password accepted")
	}
	if _, _, err := u.Login(&model.LoginRequest{Username: "nobody", Password: "secret"}); err == nil {
		t.Fatal("wrong username accepted")
	}
	token, _, err := u.Login(&model.LoginRequest{Username: "admin", Password: "secret"})
	if err != nil || !u.Verify(token) {
		t.Fatalf("valid login rejected: %v", err)
	}
}
