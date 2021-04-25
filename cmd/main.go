package main

import (
	"github.com/abdukhashimov/go_gin_example/api"
	"github.com/abdukhashimov/go_gin_example/config"
	"github.com/abdukhashimov/go_gin_example/pkg/grpc_client"
	"github.com/abdukhashimov/go_gin_example/pkg/logger"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.LogLevel, "test-go-gin-grpc")
	gprcClients, _ := grpc_client.New(cfg)

	server := api.New(api.Config{
		Logger:     log,
		GrpcClient: gprcClients,
		Cfg:        &cfg,
	})

	server.Run(cfg.HttpPort)

}
