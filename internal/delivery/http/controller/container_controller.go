package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/idnatiya/netradock/internal/model"
	"github.com/idnatiya/netradock/internal/usecase"
	"github.com/sirupsen/logrus"
)

type ContainerController struct {
	Log     *logrus.Logger
	UseCase *usecase.ContainerUseCase
}

func NewContainerController(log *logrus.Logger, useCase *usecase.ContainerUseCase) *ContainerController {
	return &ContainerController{Log: log, UseCase: useCase}
}

func (c *ContainerController) List(ctx *fiber.Ctx) error {
	res, err := c.UseCase.List(ctx.UserContext(), ctx.QueryBool("all", true))
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[[]model.ContainerResponse]{Data: res})
}

func (c *ContainerController) Get(ctx *fiber.Ctx) error {
	raw, err := c.UseCase.Inspect(ctx.UserContext(), ctx.Params("id"))
	if err != nil {
		return err
	}
	ctx.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	return ctx.Send(append(append([]byte(`{"data":`), raw...), '}'))
}

func (c *ContainerController) Action(ctx *fiber.Ctx) error {
	id, action := ctx.Params("id"), ctx.Params("action")
	if err := c.UseCase.Action(ctx.UserContext(), id, action); err != nil {
		return err
	}
	c.Log.WithFields(logrus.Fields{"container": id, "action": action}).Info("container action")
	return ctx.JSON(model.WebResponse[bool]{Data: true})
}

func (c *ContainerController) Remove(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.UseCase.Remove(ctx.UserContext(), id, ctx.QueryBool("force")); err != nil {
		return err
	}
	c.Log.WithField("container", id).Info("container removed")
	return ctx.JSON(model.WebResponse[bool]{Data: true})
}
