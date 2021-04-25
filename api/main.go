package api

import (
	"net/http"

	v1 "github.com/abdukhashimov/go_gin_example/api/handlers/v1"
	"github.com/abdukhashimov/go_gin_example/config"
	"github.com/abdukhashimov/go_gin_example/pkg/grpc_client"
	"github.com/abdukhashimov/go_gin_example/pkg/logger"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//Config ...
type Config struct {
	Logger     logger.Logger
	GrpcClient *grpc_client.GrpcClient
	Cfg        *config.Config
}

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func New(cnf Config) *gin.Engine {
	router := gin.New()

	router.Static("/images", "./static/images")

	router.Use(gin.Logger())

	router.Use(gin.Recovery())

	//r.Use(middleware.NewAuthorizer(cnf.CasbinEnforcer))

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = append(config.AllowHeaders, "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	config.AllowHeaders = append(config.AllowHeaders, "*")

	router.Use(cors.New(config))

	handlerV1 := v1.New(&v1.HandlerV1Config{
		Logger:     cnf.Logger,
		GrpcClient: cnf.GrpcClient,
		Cfg:        cnf.Cfg,
	})

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "Api gateway"})
	})

	apiV1 := router.Group("/v1")
	{
		// -- Todo -->
		apiV1.GET("/todo", handlerV1.GetAllTodo)
		apiV1.POST("/todo", handlerV1.CreateNewTodo)
		apiV1.GET("/todo/:id", handlerV1.GetTodo)
		apiV1.PUT("/todo/:id", handlerV1.UpdateTodo)
		apiV1.DELETE("/todo/:id", handlerV1.DeleteTodo)
		// <-- End Todo ---
	}

	url := ginSwagger.URL("swagger/doc.json") // The url pointing to API definition
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	return router
}
