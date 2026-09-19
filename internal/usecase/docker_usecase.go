package usecase

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/idnatiya/netradock/internal/model"
	"github.com/idnatiya/netradock/internal/model/converter"
	"github.com/idnatiya/netradock/internal/repository"
	"github.com/sirupsen/logrus"
)

// DockerUseCase covers system info, images, volumes and networks.
type DockerUseCase struct {
	Log        *logrus.Logger
	Validate   *validator.Validate
	Repository *repository.DockerRepository
}

func NewDockerUseCase(log *logrus.Logger, validate *validator.Validate, repo *repository.DockerRepository) *DockerUseCase {
	return &DockerUseCase{Log: log, Validate: validate, Repository: repo}
}

func (u *DockerUseCase) System(ctx context.Context) (model.SystemResponse, error) {
	info, err := u.Repository.Info(ctx)
	return converter.SystemToResponse(info), err
}

func (u *DockerUseCase) ListImages(ctx context.Context) ([]model.ImageResponse, error) {
	items, err := u.Repository.ListImages(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]model.ImageResponse, len(items))
	for i, it := range items {
		res[i] = converter.ImageToResponse(it)
	}
	return res, nil
}

func (u *DockerUseCase) PullImage(ctx context.Context, req *model.PullImageRequest) error {
	if err := u.Validate.Struct(req); err != nil {
		return fiber.ErrBadRequest
	}
	u.Log.WithField("image", req.Image).Info("pulling image")
	return u.Repository.PullImage(ctx, req.Image)
}

func (u *DockerUseCase) RemoveImage(ctx context.Context, id string, force bool) error {
	return u.Repository.RemoveImage(ctx, id, force)
}

func (u *DockerUseCase) ListVolumes(ctx context.Context) ([]model.VolumeResponse, error) {
	items, err := u.Repository.ListVolumes(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]model.VolumeResponse, len(items))
	for i, it := range items {
		res[i] = converter.VolumeToResponse(it)
	}
	return res, nil
}

func (u *DockerUseCase) RemoveVolume(ctx context.Context, name string) error {
	return u.Repository.RemoveVolume(ctx, name)
}

func (u *DockerUseCase) ListNetworks(ctx context.Context) ([]model.NetworkResponse, error) {
	items, err := u.Repository.ListNetworks(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]model.NetworkResponse, len(items))
	for i, it := range items {
		res[i] = converter.NetworkToResponse(it)
	}
	return res, nil
}

func (u *DockerUseCase) RemoveNetwork(ctx context.Context, id string) error {
	return u.Repository.RemoveNetwork(ctx, id)
}
