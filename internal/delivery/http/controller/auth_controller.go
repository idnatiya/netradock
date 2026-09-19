package controller

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/idnatiya/netradock/internal/delivery/http/middleware"
	"github.com/idnatiya/netradock/internal/model"
	"github.com/idnatiya/netradock/internal/usecase"
	"github.com/sirupsen/logrus"
)

type AuthController struct {
	Log          *logrus.Logger
	UseCase      *usecase.AuthUseCase
	SecureCookie bool
}

func NewAuthController(log *logrus.Logger, useCase *usecase.AuthUseCase, secureCookie bool) *AuthController {
	return &AuthController{Log: log, UseCase: useCase, SecureCookie: secureCookie}
}

func (c *AuthController) Login(ctx *fiber.Ctx) error {
	req := new(model.LoginRequest)
	if err := ctx.BodyParser(req); err != nil {
		return fiber.ErrBadRequest
	}
	token, exp, err := c.UseCase.Login(req)
	if err != nil {
		c.Log.WithField("ip", ctx.IP()).Warn("failed login")
		return err
	}
	c.setCookie(ctx, token, exp)
	return ctx.JSON(model.WebResponse[model.UserResponse]{Data: model.UserResponse{Username: req.Username}})
}

func (c *AuthController) Logout(ctx *fiber.Ctx) error {
	c.setCookie(ctx, "", time.Unix(0, 0))
	return ctx.JSON(model.WebResponse[bool]{Data: true})
}

func (c *AuthController) Current(ctx *fiber.Ctx) error {
	return ctx.JSON(model.WebResponse[model.UserResponse]{Data: model.UserResponse{Username: c.UseCase.Username}})
}

func (c *AuthController) setCookie(ctx *fiber.Ctx, value string, exp time.Time) {
	ctx.Cookie(&fiber.Cookie{
		Name:     middleware.SessionCookie,
		Value:    value,
		Path:     "/",
		Expires:  exp,
		HTTPOnly: true,
		Secure:   c.SecureCookie,
		SameSite: fiber.CookieSameSiteStrictMode,
	})
}
