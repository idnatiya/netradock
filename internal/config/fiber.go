package config

import (
	"errors"
	"time"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewFiber(v *viper.Viper, log *logrus.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "Netradock",
		ErrorHandler: errorHandler(log),
		// Image IDs ("sha256:...") arrive percent-encoded; decode before matching params.
		UnescapePath: true,
		// Behind a reverse proxy set NETRADOCK_PROXY_HEADER=X-Forwarded-For so the login limiter sees client IPs.
		ProxyHeader: v.GetString("proxy_header"),
		// Caps how long a client may take to send a request; fasthttp clears deadlines on
		// hijack, so this does not cut WebSocket streams short. No WriteTimeout for the same reason.
		ReadTimeout: 15 * time.Second,
		IdleTimeout: 60 * time.Second,
	})

	app.Use(helmet.New(helmet.Config{
		// Everything is served from the embedded bundle; 'self' also covers same-origin ws:// and wss://.
		// xterm.js and Vue inject <style> at runtime, hence unsafe-inline for styles only.
		ContentSecurityPolicy: "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data:; font-src 'self' data:; connect-src 'self'; " +
			"frame-ancestors 'none'; base-uri 'none'; form-action 'self'; object-src 'none'",
		XFrameOptions:  "DENY",
		ReferrerPolicy: "no-referrer",
		// Only meaningful over HTTPS; the reverse proxy is expected to terminate TLS.
		HSTSMaxAge:       31536000,
		PermissionPolicy: "camera=(), microphone=(), geolocation=(), interest-cohort=()",
	}))

	return app
}

// errorHandler maps daemon errors onto status codes. Only errors this codebase
// raises itself carry their message through; anything else would leak host paths,
// socket names and daemon internals to the client, so it goes to the log instead.
func errorHandler(log *logrus.Logger) fiber.ErrorHandler {
	return func(ctx *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		var fe *fiber.Error
		switch {
		case errors.As(err, &fe):
			return ctx.Status(fe.Code).JSON(fiber.Map{"errors": fe.Message})
		case cerrdefs.IsNotFound(err):
			code = fiber.StatusNotFound
		case cerrdefs.IsConflict(err):
			code = fiber.StatusConflict
		case cerrdefs.IsInvalidArgument(err):
			code = fiber.StatusBadRequest
		case cerrdefs.IsPermissionDenied(err):
			code = fiber.StatusForbidden
		}
		log.WithError(err).WithField("path", ctx.Path()).Warn("request failed")
		return ctx.Status(code).JSON(fiber.Map{"errors": utils.StatusMessage(code)})
	}
}
