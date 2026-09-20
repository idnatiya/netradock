package middleware

import (
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/idnatiya/netradock/internal/usecase"
)

const SessionCookie = "netradock_session"

func NewAuth(auth *usecase.AuthUseCase) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if !auth.Verify(ctx.Cookies(SessionCookie)) {
			return fiber.ErrUnauthorized
		}
		return ctx.Next()
	}
}

// NewSameOriginWrites applies the origin check to state-changing requests only.
// Browsers omit Origin on same-origin GETs, so requiring it everywhere would break reads.
func NewSameOriginWrites() fiber.Handler {
	check := NewSameOrigin()
	return func(ctx *fiber.Ctx) error {
		switch ctx.Method() {
		case fiber.MethodGet, fiber.MethodHead, fiber.MethodOptions:
			return ctx.Next()
		}
		return check(ctx)
	}
}

// NewSameOrigin rejects cross-origin WebSocket upgrades; browsers attach cookies to them regardless of CORS.
func NewSameOrigin() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		origin, err := url.Parse(ctx.Get(fiber.HeaderOrigin))
		if err != nil || origin.Host != string(ctx.Request().Host()) {
			return fiber.ErrForbidden
		}
		return ctx.Next()
	}
}
