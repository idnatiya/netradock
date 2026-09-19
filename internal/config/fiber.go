package config

import (
	"errors"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
)

func NewFiber(v *viper.Viper) *fiber.App {
	return fiber.New(fiber.Config{
		AppName:      "Netradock",
		ErrorHandler: errorHandler,
		// Behind a reverse proxy set NETRADOCK_PROXY_HEADER=X-Forwarded-For so the login limiter sees client IPs.
		ProxyHeader: v.GetString("proxy_header"),
	})
}

func errorHandler(ctx *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	var fe *fiber.Error
	switch {
	case errors.As(err, &fe):
		code = fe.Code
	case cerrdefs.IsNotFound(err):
		code = fiber.StatusNotFound
	case cerrdefs.IsConflict(err):
		code = fiber.StatusConflict
	case cerrdefs.IsInvalidArgument(err):
		code = fiber.StatusBadRequest
	}
	return ctx.Status(code).JSON(fiber.Map{"errors": err.Error()})
}
