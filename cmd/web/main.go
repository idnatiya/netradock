package main

import (
	"fmt"

	"github.com/idnatiya/netradock/internal/config"
	"github.com/idnatiya/netradock/web"
)

func main() {
	viperConfig := config.NewViper()
	log := config.NewLogger(viperConfig)
	docker := config.NewDocker(log)
	defer docker.Close()
	app := config.NewFiber(viperConfig, log)

	config.Bootstrap(&config.BootstrapConfig{
		App:      app,
		Docker:   docker,
		Log:      log,
		Validate: config.NewValidator(),
		Config:   viperConfig,
		Web:      web.Dist(),
	})

	port := viperConfig.GetInt("port")
	if err := app.Listen(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
