package v1

import (
	"github.com/gin-gonic/gin"
)

func (h *handlerV1) CreateNewTodo(c *gin.Context) {

}

// @Router /v1/todo [get]
// @Summary Get Daily View Count
// @Description API for getting daily view count
// @Tags analytics
// @Accept  json
// @Produce  json
// @Success 200 {object} models.AllTodoModel
// @Failure 404 {object} models.ResponseError
// @Failure 500 {object} models.ResponseError
func (h *handlerV1) GetAllTodo(c *gin.Context) {

}

func (h *handlerV1) GetTodo(c *gin.Context) {

}

func (h *handlerV1) UpdateTodo(c *gin.Context) {

}

func (h *handlerV1) DeleteTodo(c *gin.Context) {

}
