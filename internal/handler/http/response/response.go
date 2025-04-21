package response

import (
	"github.com/gin-gonic/gin"

	"pvz-service-avito-internship/internal/handler/http/api"
)

func SendError(c *gin.Context, statusCode int, message string) {
	errorResponse := api.Error{
		Message: message,
	}
	c.JSON(statusCode, errorResponse)
}

func SendSuccess(c *gin.Context, statusCode int, data interface{}) {
	if data == nil {
		c.Status(statusCode)
		return
	}
	c.JSON(statusCode, data)
}
