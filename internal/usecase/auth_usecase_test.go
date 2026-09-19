package usecase

import (
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/idnatiya/netradock/internal/model"
)

func TestAuthUseCase(t *testing.T) {
	u := NewAuthUseCase(validator.New(), "admin", "secret", []byte("key"), time.Hour)

	if _, _, err := u.Login(&model.LoginRequest{Username: "admin", Password: "wrong"}); err == nil {
		t.Fatal("wrong password accepted")
	}
	token, _, err := u.Login(&model.LoginRequest{Username: "admin", Password: "secret"})
	if err != nil || !u.Verify(token) {
		t.Fatalf("valid login rejected: %v", err)
	}

	exp, sig, _ := strings.Cut(token, ".")
	if u.Verify(exp + "9." + sig) {
		t.Fatal("tampered expiry accepted")
	}
	if u.Verify(exp + "." + strings.Repeat("0", len(sig))) {
		t.Fatal("tampered signature accepted")
	}
	if NewAuthUseCase(validator.New(), "admin", "secret", []byte("other"), time.Hour).Verify(token) {
		t.Fatal("token from another secret accepted")
	}
	expired := u.sign(time.Now().Add(-time.Second))
	if u.Verify(expired) {
		t.Fatal("expired token accepted")
	}
}
