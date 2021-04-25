package grpc_client

import (
	"github.com/abdukhashimov/go_gin_example/config"
)

//GrpcClientI ...
type GrpcClientI interface {
}

//GrpcClient ...
type GrpcClient struct {
	cfg         config.Config
	connections map[string]interface{}
}

//New ...
func New(cfg config.Config) (*GrpcClient, error) {

	return &GrpcClient{
		cfg:         cfg,
		connections: map[string]interface{}{},
	}, nil
}
