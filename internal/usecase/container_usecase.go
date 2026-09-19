package usecase

import (
	"context"
	"encoding/json"
	"io"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/idnatiya/netradock/internal/model"
	"github.com/idnatiya/netradock/internal/model/converter"
	"github.com/idnatiya/netradock/internal/repository"
	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/sirupsen/logrus"
)

type ContainerUseCase struct {
	Log        *logrus.Logger
	Repository *repository.DockerRepository
}

func NewContainerUseCase(log *logrus.Logger, repo *repository.DockerRepository) *ContainerUseCase {
	return &ContainerUseCase{Log: log, Repository: repo}
}

func (u *ContainerUseCase) List(ctx context.Context, all bool) ([]model.ContainerResponse, error) {
	items, err := u.Repository.ListContainers(ctx, all)
	if err != nil {
		return nil, err
	}
	res := make([]model.ContainerResponse, len(items))
	for i, c := range items {
		res[i] = converter.ContainerToResponse(c)
	}
	return res, nil
}

func (u *ContainerUseCase) Inspect(ctx context.Context, id string) (json.RawMessage, error) {
	_, raw, err := u.Repository.InspectContainer(ctx, id)
	return raw, err
}

func (u *ContainerUseCase) Action(ctx context.Context, id, action string) error {
	switch action {
	case "start":
		return u.Repository.StartContainer(ctx, id)
	case "stop":
		return u.Repository.StopContainer(ctx, id)
	case "restart":
		return u.Repository.RestartContainer(ctx, id)
	}
	return fiber.NewError(fiber.StatusBadRequest, "unknown action: "+action)
}

func (u *ContainerUseCase) Remove(ctx context.Context, id string, force bool) error {
	return u.Repository.RemoveContainer(ctx, id, force)
}

// StreamLogs follows the container logs into w until ctx is cancelled or the container exits.
func (u *ContainerUseCase) StreamLogs(ctx context.Context, id string, tail int, w io.Writer) error {
	info, _, err := u.Repository.InspectContainer(ctx, id)
	if err != nil {
		return err
	}
	logs, err := u.Repository.ContainerLogs(ctx, id, tail)
	if err != nil {
		return err
	}
	defer logs.Close()
	if info.Config != nil && info.Config.Tty {
		_, err = io.Copy(w, logs)
	} else {
		_, err = stdcopy.StdCopy(w, w, logs)
	}
	return err
}

// StreamStats calls send once per sample (about every second) until ctx is cancelled.
func (u *ContainerUseCase) StreamStats(ctx context.Context, id string, send func(model.StatsResponse) error) error {
	body, err := u.Repository.ContainerStats(ctx, id)
	if err != nil {
		return err
	}
	defer body.Close()
	dec := json.NewDecoder(body)
	for {
		var s container.StatsResponse
		if err := dec.Decode(&s); err != nil {
			return err
		}
		if err := send(converter.StatsToResponse(s)); err != nil {
			return err
		}
	}
}

// StreamAllStats sends the latest sample of every running container every 2 seconds.
// ponytail: one daemon stats stream per running container; fine for dozens, not thousands.
func (u *ContainerUseCase) StreamAllStats(ctx context.Context, send func(map[string]model.StatsResponse) error) error {
	var mu sync.Mutex
	latest := map[string]model.StatsResponse{}
	streams := map[string]context.CancelFunc{}
	defer func() {
		for _, cancel := range streams {
			cancel()
		}
	}()

	refresh := func() error {
		items, err := u.Repository.ListContainers(ctx, false)
		if err != nil {
			return err
		}
		running := make(map[string]bool, len(items))
		for _, c := range items {
			running[c.ID] = true
			if streams[c.ID] != nil {
				continue
			}
			sctx, cancel := context.WithCancel(ctx)
			streams[c.ID] = cancel
			go func(id string) {
				err := u.StreamStats(sctx, id, func(s model.StatsResponse) error {
					mu.Lock()
					latest[id] = s
					mu.Unlock()
					return nil
				})
				if err != nil && sctx.Err() == nil {
					u.Log.WithError(err).WithField("container", id).Debug("stats stream ended")
				}
			}(c.ID)
		}
		mu.Lock()
		for id, cancel := range streams {
			if !running[id] {
				cancel()
				delete(streams, id)
				delete(latest, id)
			}
		}
		mu.Unlock()
		return nil
	}

	if err := refresh(); err != nil {
		return err
	}
	tick := time.NewTicker(2 * time.Second)
	defer tick.Stop()
	for n := 1; ; n++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tick.C:
		}
		// Pick up started and stopped containers every other tick (~4s).
		if n%2 == 0 {
			if err := refresh(); err != nil {
				return err
			}
		}
		mu.Lock()
		snapshot := make(map[string]model.StatsResponse, len(latest))
		for id, s := range latest {
			snapshot[id] = s
		}
		mu.Unlock()
		if err := send(snapshot); err != nil {
			return err
		}
	}
}

func (u *ContainerUseCase) Exec(ctx context.Context, id string, cmd []string) (string, client.HijackedResponse, error) {
	return u.Repository.Exec(ctx, id, cmd)
}

func (u *ContainerUseCase) ExecResize(ctx context.Context, execID string, cols, rows uint) error {
	return u.Repository.ExecResize(ctx, execID, cols, rows)
}
