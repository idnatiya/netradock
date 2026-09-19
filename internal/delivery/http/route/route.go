package route

import (
	"io/fs"
	"net/http"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/idnatiya/netradock/internal/delivery/http/controller"
)

type RouteConfig struct {
	App                 *fiber.App
	AuthMiddleware      fiber.Handler
	SameOrigin          fiber.Handler
	AuthController      *controller.AuthController
	ContainerController *controller.ContainerController
	DockerController    *controller.DockerController
	WSController        *controller.WSController
	Web                 fs.FS
}

func (c *RouteConfig) Setup() {
	c.SetupGuestRoute()
	c.SetupAuthRoute()
	c.SetupWebRoute()
}

func (c *RouteConfig) SetupGuestRoute() {
	c.App.Post("/api/auth/login", limiter.New(limiter.Config{Max: 5, Expiration: time.Minute}), c.AuthController.Login)
	c.App.Post("/api/auth/logout", c.AuthController.Logout)
}

func (c *RouteConfig) SetupAuthRoute() {
	api := c.App.Group("/api", c.AuthMiddleware)
	api.Get("/auth/me", c.AuthController.Current)
	api.Get("/system", c.DockerController.System)

	api.Get("/containers", c.ContainerController.List)
	api.Get("/containers/:id", c.ContainerController.Get)
	api.Post("/containers/:id/:action<regex(^(start|stop|restart)$)>", c.ContainerController.Action)
	api.Delete("/containers/:id", c.ContainerController.Remove)

	api.Get("/images", c.DockerController.ListImages)
	api.Post("/images/pull", c.DockerController.PullImage)
	api.Delete("/images/:id", c.DockerController.RemoveImage)

	api.Get("/volumes", c.DockerController.ListVolumes)
	api.Delete("/volumes/:name", c.DockerController.RemoveVolume)

	api.Get("/networks", c.DockerController.ListNetworks)
	api.Delete("/networks/:id", c.DockerController.RemoveNetwork)

	api.Use(func(ctx *fiber.Ctx) error { return fiber.ErrNotFound })

	ws := c.App.Group("/ws", c.SameOrigin, c.AuthMiddleware, func(ctx *fiber.Ctx) error {
		if !websocket.IsWebSocketUpgrade(ctx) {
			return fiber.ErrUpgradeRequired
		}
		return ctx.Next()
	})
	ws.Get("/containers/:id/logs", websocket.New(c.WSController.Logs))
	ws.Get("/containers/:id/stats", websocket.New(c.WSController.Stats))
	ws.Get("/containers/:id/exec", websocket.New(c.WSController.Exec))
}

// SetupWebRoute serves the built Vue app; unknown paths get index.html for client-side routing.
func (c *RouteConfig) SetupWebRoute() {
	c.App.Use("/", filesystem.New(filesystem.Config{
		Root:         http.FS(c.Web),
		Index:        "index.html",
		NotFoundFile: "index.html",
	}))
}
