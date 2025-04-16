package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"pvz-service-avito-internship/internal/handler/http/response" // Используем наш хелпер ответа
)

// Recovery - это Gin middleware для восстановления после паник.
// Он перехватывает панику, логирует ее с Request ID и стектрейсом,
// и возвращает клиенту стандартизированный ответ 500 Internal Server Error.
// Это аналог стандартного gin.Recovery(), но с использованием нашего логгера slog.
func Recovery(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Получаем Request ID для лога
				reqID := GetRequestIDFromContext(c)
				requestLogger := log.With(slog.String("request_id", reqID))

				// Проверяем, не является ли ошибка результатом обрыва соединения клиентом
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				// Дамп запроса для отладки (осторожно с чувствительными данными!)
				httpRequest, _ := httputil.DumpRequest(c.Request, false)

				if brokenPipe {
					// Если соединение разорвано клиентом, логируем как Error, но не возвращаем 500
					requestLogger.Error("Connection error recovered",
						slog.Any("error", err),
						slog.String("request", string(httpRequest)),
					)
					c.Abort() // Просто прерываем обработку
					return
				}

				// Логируем панику с уровнем Error, включая стектрейс
				requestLogger.Error("Panic recovered",
					slog.Any("error", err),
					slog.String("request", string(httpRequest)),
					slog.String("stack", string(debug.Stack())), // Добавляем стектрейс
				)

				// Возвращаем клиенту 500 Internal Server Error
				response.SendError(c, http.StatusInternalServerError, "Internal server error")
				c.AbortWithStatus(http.StatusInternalServerError) // Прерываем и устанавливаем статус
			}
		}()
		c.Next() // Передаем управление дальше
	}
}
