package config

import (
	"github.com/moby/moby/client"
	"github.com/sirupsen/logrus"
)

// NewDocker reads DOCKER_HOST etc. from the environment, defaulting to the local socket.
func NewDocker(log *logrus.Logger) *client.Client {
	cli, err := client.New(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("failed to create docker client: %v", err)
	}
	return cli
}
