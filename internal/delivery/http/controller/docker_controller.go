package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/idnatiya/netradock/internal/model"
	"github.com/idnatiya/netradock/internal/usecase"
	"github.com/sirupsen/logrus"
)

type DockerController struct {
	Log     *logrus.Logger
	UseCase *usecase.DockerUseCase
}

func NewDockerController(log *logrus.Logger, useCase *usecase.DockerUseCase) *DockerController {
	return &DockerController{Log: log, UseCase: useCase}
}

func (c *DockerController) System(ctx *fiber.Ctx) error {
	res, err := c.UseCase.System(ctx.UserContext())
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[model.SystemResponse]{Data: res})
}

func (c *DockerController) ListImages(ctx *fiber.Ctx) error {
	res, err := c.UseCase.ListImages(ctx.UserContext())
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[[]model.ImageResponse]{Data: res})
}

func (c *DockerController) PullImage(ctx *fiber.Ctx) error {
	req := new(model.PullImageRequest)
	if err := ctx.BodyParser(req); err != nil {
		return fiber.ErrBadRequest
	}
	if err := c.UseCase.PullImage(ctx.UserContext(), req); err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[bool]{Data: true})
}

func (c *DockerController) RemoveImage(ctx *fiber.Ctx) error {
	if err := c.UseCase.RemoveImage(ctx.UserContext(), ctx.Params("id"), ctx.QueryBool("force")); err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[bool]{Data: true})
}

func (c *DockerController) ListVolumes(ctx *fiber.Ctx) error {
	res, err := c.UseCase.ListVolumes(ctx.UserContext())
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[[]model.VolumeResponse]{Data: res})
}

func (c *DockerController) RemoveVolume(ctx *fiber.Ctx) error {
	if err := c.UseCase.RemoveVolume(ctx.UserContext(), ctx.Params("name")); err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[bool]{Data: true})
}

func (c *DockerController) ListNetworks(ctx *fiber.Ctx) error {
	res, err := c.UseCase.ListNetworks(ctx.UserContext())
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[[]model.NetworkResponse]{Data: res})
}

func (c *DockerController) RemoveNetwork(ctx *fiber.Ctx) error {
	if err := c.UseCase.RemoveNetwork(ctx.UserContext(), ctx.Params("id")); err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[bool]{Data: true})
}
