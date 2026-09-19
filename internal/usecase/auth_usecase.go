package usecase

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/idnatiya/netradock/internal/model"
)

// AuthUseCase checks the single user from env and issues stateless session tokens
// of the form "<unix expiry>.<hex hmac-sha256(expiry)>".
type AuthUseCase struct {
	Validate *validator.Validate
	Username string
	Password string
	Secret   []byte
	TTL      time.Duration
}

func NewAuthUseCase(validate *validator.Validate, username, password string, secret []byte, ttl time.Duration) *AuthUseCase {
	return &AuthUseCase{Validate: validate, Username: username, Password: password, Secret: secret, TTL: ttl}
}

// Login returns a session token and its expiry.
func (u *AuthUseCase) Login(req *model.LoginRequest) (string, time.Time, error) {
	if err := u.Validate.Struct(req); err != nil {
		return "", time.Time{}, fiber.ErrBadRequest
	}
	userOK := subtle.ConstantTimeCompare([]byte(req.Username), []byte(u.Username))
	passOK := subtle.ConstantTimeCompare([]byte(req.Password), []byte(u.Password))
	if userOK&passOK != 1 {
		return "", time.Time{}, fiber.NewError(fiber.StatusUnauthorized, "invalid username or password")
	}
	exp := time.Now().Add(u.TTL)
	return u.sign(exp), exp, nil
}

func (u *AuthUseCase) Verify(token string) bool {
	expStr, _, ok := strings.Cut(token, ".")
	if !ok {
		return false
	}
	exp, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil || time.Now().Unix() >= exp {
		return false
	}
	return hmac.Equal([]byte(token), []byte(u.sign(time.Unix(exp, 0))))
}

func (u *AuthUseCase) sign(exp time.Time) string {
	expStr := strconv.FormatInt(exp.Unix(), 10)
	mac := hmac.New(sha256.New, u.Secret)
	mac.Write([]byte(expStr))
	return expStr + "." + hex.EncodeToString(mac.Sum(nil))
}
