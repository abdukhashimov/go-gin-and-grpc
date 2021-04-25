package main

import "github.com/abdukhashimov/go_gin_example/pkg/logger"

func main() {
	cfg := config.Load()
	log := logger.New(cfg.LogLevel, "voxe_api_gateway")
	gprcClients, _ := services.NewGrpcClients(&cfg)

	server := api.New(&api.RouterOptions{
		Log:      log,
		Cfg:      &cfg,
		Services: gprcClients,
	})

	server.Run(cfg.HttpPort)

}
