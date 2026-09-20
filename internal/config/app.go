package config

import (
	"crypto/rand"
	"io/fs"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/idnatiya/netradock/internal/delivery/http/controller"
	"github.com/idnatiya/netradock/internal/delivery/http/middleware"
	"github.com/idnatiya/netradock/internal/delivery/http/route"
	"github.com/idnatiya/netradock/internal/repository"
	"github.com/idnatiya/netradock/internal/usecase"
	"github.com/moby/moby/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type BootstrapConfig struct {
	App      *fiber.App
	Docker   *client.Client
	Log      *logrus.Logger
	Validate *validator.Validate
	Config   *viper.Viper
	Web      fs.FS
}

func Bootstrap(config *BootstrapConfig) {
	username := config.Config.GetString("username")
	password, hash := config.Config.GetString("password"), config.Config.GetString("password_hash")
	if username == "" || (password == "" && hash == "") {
		config.Log.Fatal("NETRADOCK_USERNAME and NETRADOCK_PASSWORD (or NETRADOCK_PASSWORD_HASH) must be set")
	}
	if hash == "" {
		config.Log.Warn("NETRADOCK_PASSWORD is stored in plaintext; prefer NETRADOCK_PASSWORD_HASH with a bcrypt hash")
	}
	secret := []byte(config.Config.GetString("secret"))
	if len(secret) == 0 {
		config.Log.Warn("NETRADOCK_SECRET not set, sessions will not survive a restart")
		secret = []byte(rand.Text())
	}

	dockerRepository := repository.NewDockerRepository(config.Docker)

	ttl := time.Duration(config.Config.GetInt("session.hours")) * time.Hour
	authUseCase := usecase.NewAuthUseCase(config.Validate, username, password, hash, secret, ttl)
	containerUseCase := usecase.NewContainerUseCase(config.Log, dockerRepository)
	dockerUseCase := usecase.NewDockerUseCase(config.Log, config.Validate, dockerRepository)

	routeConfig := route.RouteConfig{
		App:                 config.App,
		AuthMiddleware:      middleware.NewAuth(authUseCase),
		SameOrigin:          middleware.NewSameOrigin(),
		SameOriginWrites:    middleware.NewSameOriginWrites(),
		AuthController:      controller.NewAuthController(config.Log, authUseCase, config.Config.GetBool("secure_cookie")),
		ContainerController: controller.NewContainerController(config.Log, containerUseCase),
		DockerController:    controller.NewDockerController(config.Log, dockerUseCase),
		WSController:        controller.NewWSController(config.Log, containerUseCase, authUseCase),
		Web:                 config.Web,
	}
	routeConfig.Setup()
}
