package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// contextKeyRequestID - неэкспортируемый тип для ключа Request ID в контексте.
type contextKeyRequestID string

// RequestIDKey - ключ для доступа к Request ID в контексте.
const RequestIDKey contextKeyRequestID = "requestID"

// LoggingMiddleware содержит зависимости для middleware логирования.
type LoggingMiddleware struct {
	log *slog.Logger
}

// NewLoggingMiddleware создает новый экземпляр LoggingMiddleware.
func NewLoggingMiddleware(log *slog.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{log: log}
}

// LogRequest - это Gin middleware функция, которая:
//  1. Генерирует уникальный Request ID для каждого входящего запроса.
//  2. Добавляет Request ID в контекст Gin (`c.Set`) и в контекст стандартной библиотеки (`context.WithValue`).
//  3. Создает логгер `slog`, обогащенный Request ID и базовой информацией о запросе.
//  4. Логирует сообщение "Request started".
//  5. Передает управление следующему обработчику (`c.Next()`).
//  6. После выполнения всех обработчиков, логирует сообщение "Request completed" с информацией о статусе,
//     времени выполнения, User-Agent и возможных ошибках, произошедших во время обработки.
func (m *LoggingMiddleware) LogRequest(c *gin.Context) {
	start := time.Now()
	path := c.Request.URL.Path
	rawQuery := c.Request.URL.RawQuery
	method := c.Request.Method
	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()
	requestID := uuid.New().String() // Генерируем уникальный ID

	// Добавляем Request ID в контекст Gin
	c.Set(string(RequestIDKey), requestID)

	// Создаем новый стандартный контекст с Request ID для передачи в сервисы/репозитории
	// Оборачиваем исходный контекст запроса
	ctxWithReqID := context.WithValue(c.Request.Context(), RequestIDKey, requestID)
	// Обновляем контекст в Gin запросе, чтобы он был доступен через c.Request.Context()
	c.Request = c.Request.WithContext(ctxWithReqID)

	// Создаем логгер для этого конкретного запроса
	requestLogger := m.log.With(
		slog.String("request_id", requestID),
		slog.String("method", method),
		slog.String("path", path),
		slog.String("remote_ip", ip),
	)

	// Логируем начало обработки запроса (уровень Info или Debug)
	requestLogger.Info("Request started")

	// Передаем управление дальше по цепочке middleware и хендлеров
	c.Next()

	// Код ниже выполнится после того, как все хендлеры отработают

	latency := time.Since(start)
	statusCode := c.Writer.Status()
	// Собираем ошибки, которые могли быть добавлены в контекст Gin (например, при панике)
	errorMessages := c.Errors.ByType(gin.ErrorTypeAny).String() // Собираем все типы ошибок

	// Формируем аргументы для финального лога
	logArgs := []any{
		slog.Int("status_code", statusCode),
		slog.Duration("latency", latency),
		slog.String("user_agent", userAgent),
	}
	if rawQuery != "" {
		logArgs = append(logArgs, slog.String("query", rawQuery))
	}
	if errorMessages != "" {
		logArgs = append(logArgs, slog.String("errors", errorMessages))
	}

	// Логируем завершение запроса с соответствующим уровнем
	switch {
	case statusCode >= http.StatusInternalServerError:
		requestLogger.Error("Request completed with server error", logArgs...)
	case statusCode >= http.StatusBadRequest:
		requestLogger.Warn("Request completed with client error", logArgs...)
	default:
		requestLogger.Info("Request completed successfully", logArgs...)
	}
}

// GetRequestIDFromContext - хелпер для безопасного извлечения Request ID из контекста.
// Пытается получить ID как из стандартного контекста, так и из контекста Gin.
// Возвращает пустую строку, если ID не найден.
func GetRequestIDFromContext(ctx context.Context) string {
	// Попытка из стандартного контекста
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok && reqID != "" {
		return reqID
	}

	// Попытка из контекста Gin (если передан именно он)
	if gCtx, ok := ctx.(*gin.Context); ok {
		if reqID, exists := gCtx.Get(string(RequestIDKey)); exists {
			if idStr, ok := reqID.(string); ok && idStr != "" {
				return idStr
			}
		}
		// Попробуем еще раз из обернутого контекста Gin запроса (на всякий случай)
		if reqID, ok := gCtx.Request.Context().Value(RequestIDKey).(string); ok && reqID != "" {
			return reqID
		}

	}

	// Возвращаем пустую строку, если ID не найден
	return ""
}
