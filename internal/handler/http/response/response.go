package response

import (
	"github.com/gin-gonic/gin"

	// Используем сгенерированные типы для ответа об ошибке
	"pvz-service-avito-internship/internal/handler/http/api"
)

// SendError отправляет стандартизированный JSON-ответ с ошибкой.
// Использует структуру Error из сгенерированного пакета api.
func SendError(c *gin.Context, statusCode int, message string) {
	// Создаем объект ошибки согласно спецификации OpenAPI
	errorResponse := api.Error{
		Message: message,
	}
	// Устанавливаем Content-Type и отправляем JSON
	c.JSON(statusCode, errorResponse)
}

// SendSuccess отправляет успешный JSON-ответ с данными.
// Если data равно nil, отправляет ответ со статусом statusCode без тела (например, 200 OK для удаления).
func SendSuccess(c *gin.Context, statusCode int, data interface{}) {
	if data == nil {
		// Если данных нет, просто устанавливаем статус
		c.Status(statusCode)
		return
	}
	// Отправляем JSON с данными
	c.JSON(statusCode, data)
}

// SendNoContent отправляет ответ со статусом 204 No Content.
// Используется, когда операция успешна, но возвращать тело ответа не требуется.
// (Хотя в спецификации для удаления используется 200 OK без тела, эта функция может быть полезна)
/*
func SendNoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
*/
