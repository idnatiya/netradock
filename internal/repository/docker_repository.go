package repository

import (
	"context"
	"encoding/json"
	"io"
	"strconv"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/image"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/system"
	"github.com/moby/moby/api/types/volume"
	"github.com/moby/moby/client"
)

type DockerRepository struct {
	Client *client.Client
}

func NewDockerRepository(cli *client.Client) *DockerRepository {
	return &DockerRepository{Client: cli}
}

func (r *DockerRepository) Info(ctx context.Context) (system.Info, error) {
	res, err := r.Client.Info(ctx, client.InfoOptions{})
	return res.Info, err
}

func (r *DockerRepository) ListContainers(ctx context.Context, all bool) ([]container.Summary, error) {
	res, err := r.Client.ContainerList(ctx, client.ContainerListOptions{All: all})
	return res.Items, err
}

func (r *DockerRepository) InspectContainer(ctx context.Context, id string) (container.InspectResponse, json.RawMessage, error) {
	res, err := r.Client.ContainerInspect(ctx, id, client.ContainerInspectOptions{})
	return res.Container, res.Raw, err
}

func (r *DockerRepository) StartContainer(ctx context.Context, id string) error {
	_, err := r.Client.ContainerStart(ctx, id, client.ContainerStartOptions{})
	return err
}

func (r *DockerRepository) StopContainer(ctx context.Context, id string) error {
	_, err := r.Client.ContainerStop(ctx, id, client.ContainerStopOptions{})
	return err
}

func (r *DockerRepository) RestartContainer(ctx context.Context, id string) error {
	_, err := r.Client.ContainerRestart(ctx, id, client.ContainerRestartOptions{})
	return err
}

func (r *DockerRepository) RemoveContainer(ctx context.Context, id string, force bool) error {
	_, err := r.Client.ContainerRemove(ctx, id, client.ContainerRemoveOptions{Force: force})
	return err
}

func (r *DockerRepository) ContainerLogs(ctx context.Context, id string, tail int) (io.ReadCloser, error) {
	return r.Client.ContainerLogs(ctx, id, client.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       strconv.Itoa(tail),
	})
}

func (r *DockerRepository) ContainerStats(ctx context.Context, id string) (io.ReadCloser, error) {
	res, err := r.Client.ContainerStats(ctx, id, client.ContainerStatsOptions{Stream: true})
	return res.Body, err
}

// Exec starts cmd with a TTY attached and returns the exec ID plus the hijacked connection.
func (r *DockerRepository) Exec(ctx context.Context, id string, cmd []string) (string, client.HijackedResponse, error) {
	created, err := r.Client.ExecCreate(ctx, id, client.ExecCreateOptions{
		TTY:          true,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd,
		Env:          []string{"TERM=xterm-256color"},
	})
	if err != nil {
		return "", client.HijackedResponse{}, err
	}
	attached, err := r.Client.ExecAttach(ctx, created.ID, client.ExecAttachOptions{TTY: true})
	return created.ID, attached.HijackedResponse, err
}

func (r *DockerRepository) ExecResize(ctx context.Context, execID string, cols, rows uint) error {
	_, err := r.Client.ExecResize(ctx, execID, client.ExecResizeOptions{Width: cols, Height: rows})
	return err
}

func (r *DockerRepository) ListImages(ctx context.Context) ([]image.Summary, error) {
	res, err := r.Client.ImageList(ctx, client.ImageListOptions{})
	return res.Items, err
}

// PullImage blocks until the pull finishes.
func (r *DockerRepository) PullImage(ctx context.Context, ref string) error {
	res, err := r.Client.ImagePull(ctx, ref, client.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer res.Close()
	return res.Wait(ctx)
}

func (r *DockerRepository) RemoveImage(ctx context.Context, id string, force bool) error {
	_, err := r.Client.ImageRemove(ctx, id, client.ImageRemoveOptions{Force: force})
	return err
}

func (r *DockerRepository) ListVolumes(ctx context.Context) ([]volume.Volume, error) {
	res, err := r.Client.VolumeList(ctx, client.VolumeListOptions{})
	return res.Items, err
}

func (r *DockerRepository) RemoveVolume(ctx context.Context, name string) error {
	_, err := r.Client.VolumeRemove(ctx, name, client.VolumeRemoveOptions{})
	return err
}

func (r *DockerRepository) ListNetworks(ctx context.Context) ([]network.Summary, error) {
	res, err := r.Client.NetworkList(ctx, client.NetworkListOptions{})
	return res.Items, err
}

func (r *DockerRepository) RemoveNetwork(ctx context.Context, id string) error {
	_, err := r.Client.NetworkRemove(ctx, id, client.NetworkRemoveOptions{})
	return err
}
