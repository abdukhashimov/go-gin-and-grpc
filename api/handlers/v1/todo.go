package v1

import (
	"github.com/gin-gonic/gin"
)

func (h *handlerV1) CreateNewTodo(c *gin.Context) {

}

// @Router /v1/todo [get]
// @Summary Get List of Todo
// @Description API to retreive list of todo
// @Tags TODO
// @Accept  json
// @Produce  json
// @Success 200 {object} models.AllTodoModel
// @Failure 404 {object} models.ResponseError
// @Failure 500 {object} models.ResponseError
func (h *handlerV1) GetAllTodo(c *gin.Context) {

}

// @Router /v1/todo/{id} [get]
// @Summary Get a Todo
// @Description API to retreive a single todo
// @Tags TODO
// @Accept  json
// @Produce  json
// @Success 200 {object} models.SingleTodoModel
// @Failure 404 {object} models.ResponseError
// @Failure 500 {object} models.ResponseError
func (h *handlerV1) GetTodo(c *gin.Context) {

}

// @Router /v1/todo/{id} [put]
// @Summary Get a Todo
// @Description API to retreive a single todo
// @Tags TODO
// @Accept  json
// @Produce  json
// @Success 200 {object} models.Response
// @Failure 404 {object} models.ResponseError
// @Failure 500 {object} models.ResponseError
func (h *handlerV1) UpdateTodo(c *gin.Context) {

}

// @Router /v1/todo/{id} [get]
// @Summary Get a Todo
// @Description API to retreive a single todo
// @Tags TODO
// @Accept  json
// @Produce  json
// @Success 200 {object} models.Response
// @Failure 404 {object} models.ResponseError
// @Failure 500 {object} models.ResponseError
func (h *handlerV1) DeleteTodo(c *gin.Context) {

}
