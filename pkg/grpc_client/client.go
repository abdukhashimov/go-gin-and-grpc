package grpc_client

import (
	"fmt"

	"github.com/abdukhashimov/go_gin_example/config"
	"github.com/abdukhashimov/go_gin_example/genproto/todo_service"
	"google.golang.org/grpc"
)

//GrpcClientI ...
type GrpcClientI interface {
	ToDoService() todo_service.TodoServiceClient
}

//GrpcClient ...
type GrpcClient struct {
	todoService todo_service.TodoServiceClient
}

//New ...
func New(cfg config.Config) (*GrpcClient, error) {

	todoService, err := grpc.Dial(
		fmt.Sprintf("%s:%d", cfg.TodoServiceHost, cfg.TodoServicePort),
		grpc.WithInsecure(),
	)

	if err != nil {
		return nil, err
	}

	return &GrpcClient{
		todoService: todo_service.NewTodoServiceClient(todoService),
	}, nil
}

func (g *GrpcClient) TodoService() todo_service.TodoServiceClient {
	return g.todoService
}
