package usecase

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/idnatiya/netradock/internal/model"
	"golang.org/x/crypto/bcrypt"
)

// AuthUseCase checks the single user from env and issues stateless session tokens
// of the form "<unix expiry>.<random nonce>.<hex hmac-sha256(expiry|nonce)>".
// The nonce makes every token unique, so revoking one cannot collide with a token
// issued later in the same second.
//
// The signing key is derived from the secret *and* the configured credential, so
// changing the username or password invalidates every token already issued.
type AuthUseCase struct {
	Validate *validator.Validate
	Username string
	TTL      time.Duration

	password string // plaintext credential; empty when hash is set
	hash     []byte // bcrypt hash of the credential; preferred over password
	key      []byte

	mu      sync.RWMutex
	revoked map[string]int64 // token -> unix expiry
}

// NewAuthUseCase takes either a bcrypt hash (preferred) or a plaintext password.
// When both are given the hash wins.
func NewAuthUseCase(validate *validator.Validate, username, password, hash string, secret []byte, ttl time.Duration) *AuthUseCase {
	u := &AuthUseCase{Validate: validate, Username: username, TTL: ttl, revoked: map[string]int64{}}
	verifier := hash
	if hash != "" {
		u.hash = []byte(hash)
	} else {
		u.password = password
		verifier = password
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(username))
	mac.Write([]byte{0})
	mac.Write([]byte(verifier))
	u.key = mac.Sum(nil)
	return u
}

// Login returns a session token and its expiry.
func (u *AuthUseCase) Login(req *model.LoginRequest) (string, time.Time, error) {
	if err := u.Validate.Struct(req); err != nil {
		return "", time.Time{}, fiber.ErrBadRequest
	}
	userOK := subtle.ConstantTimeCompare([]byte(req.Username), []byte(u.Username))
	if u.checkPassword(req.Password)&userOK != 1 {
		return "", time.Time{}, fiber.NewError(fiber.StatusUnauthorized, "invalid username or password")
	}
	exp := time.Now().Add(u.TTL)
	return u.sign(exp, rand.Text()), exp, nil
}

// checkPassword runs unconditionally, so a wrong username costs the same as a wrong password.
func (u *AuthUseCase) checkPassword(password string) int {
	if u.hash != nil {
		if bcrypt.CompareHashAndPassword(u.hash, []byte(password)) != nil {
			return 0
		}
		return 1
	}
	return subtle.ConstantTimeCompare([]byte(password), []byte(u.password))
}

func (u *AuthUseCase) Verify(token string) bool {
	exp, ok := expiryOf(token)
	if !ok || time.Now().Unix() >= exp {
		return false
	}
	u.mu.RLock()
	_, dead := u.revoked[token]
	u.mu.RUnlock()
	if dead {
		return false
	}
	_, rest, _ := strings.Cut(token, ".")
	nonce, _, ok := strings.Cut(rest, ".")
	if !ok {
		return false
	}
	return hmac.Equal([]byte(token), []byte(u.sign(time.Unix(exp, 0), nonce)))
}

// Revoke retires a token before its expiry, which is what makes logout mean something:
// the token itself is stateless, so without this a stolen cookie stays valid for the
// whole TTL. Entries are dropped once they expire anyway.
//
// ponytail: in-memory denylist, so a restart re-honours anything revoked before it and
// it does not carry across instances. Move it to a shared store if either starts to matter.
func (u *AuthUseCase) Revoke(token string) {
	exp, ok := expiryOf(token)
	if !ok {
		return
	}
	now := time.Now().Unix()
	if exp <= now {
		return
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	for t, e := range u.revoked {
		if e <= now {
			delete(u.revoked, t)
		}
	}
	u.revoked[token] = exp
}

func expiryOf(token string) (int64, bool) {
	expStr, _, ok := strings.Cut(token, ".")
	if !ok {
		return 0, false
	}
	exp, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		return 0, false
	}
	return exp, true
}

func (u *AuthUseCase) sign(exp time.Time, nonce string) string {
	expStr := strconv.FormatInt(exp.Unix(), 10)
	mac := hmac.New(sha256.New, u.key)
	mac.Write([]byte(expStr))
	mac.Write([]byte{0})
	mac.Write([]byte(nonce))
	return expStr + "." + nonce + "." + hex.EncodeToString(mac.Sum(nil))
}
