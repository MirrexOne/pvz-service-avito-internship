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
	"pvz-service-avito-internship/internal/handler/http/response"
)

// Recovery - это функция для восстановления после паник.
// Он перехватывает панику, логирует ее с Request ID и стектрейсомю
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

				httpRequest, _ := httputil.DumpRequest(c.Request, false)

				if brokenPipe {
					requestLogger.Error("Connection error recovered",
						slog.Any("error", err),
						slog.String("request", string(httpRequest)),
					)
					c.Abort()
					return
				}

				requestLogger.Error("Panic recovered",
					slog.Any("error", err),
					slog.String("request", string(httpRequest)),
					slog.String("stack", string(debug.Stack())),
				)

				response.SendError(c, http.StatusInternalServerError, "Internal server error")
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
