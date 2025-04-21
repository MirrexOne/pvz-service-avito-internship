package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type contextKeyRequestID string

const RequestIDKey contextKeyRequestID = "requestID"

type LoggingMiddleware struct {
	log *slog.Logger
}

func NewLoggingMiddleware(log *slog.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{log: log}
}

func (m *LoggingMiddleware) LogRequest(c *gin.Context) {
	start := time.Now()
	path := c.Request.URL.Path
	rawQuery := c.Request.URL.RawQuery
	method := c.Request.Method
	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()
	requestID := uuid.New().String()

	c.Set(string(RequestIDKey), requestID)

	ctxWithReqID := context.WithValue(c.Request.Context(), RequestIDKey, requestID)
	c.Request = c.Request.WithContext(ctxWithReqID)

	requestLogger := m.log.With(
		slog.String("request_id", requestID),
		slog.String("method", method),
		slog.String("path", path),
		slog.String("remote_ip", ip),
	)

	requestLogger.Info("Request started")

	c.Next()

	latency := time.Since(start)
	statusCode := c.Writer.Status()
	errorMessages := c.Errors.ByType(gin.ErrorTypeAny).String()

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

	switch {
	case statusCode >= http.StatusInternalServerError:
		requestLogger.Error("Request completed with server error", logArgs...)
	case statusCode >= http.StatusBadRequest:
		requestLogger.Warn("Request completed with client error", logArgs...)
	default:
		requestLogger.Info("Request completed successfully", logArgs...)
	}
}

func GetRequestIDFromContext(ctx context.Context) string {
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok && reqID != "" {
		return reqID
	}

	if gCtx, ok := ctx.(*gin.Context); ok {
		if reqID, exists := gCtx.Get(string(RequestIDKey)); exists {
			if idStr, ok := reqID.(string); ok && idStr != "" {
				return idStr
			}
		}
		if reqID, ok := gCtx.Request.Context().Value(RequestIDKey).(string); ok && reqID != "" {
			return reqID
		}

	}

	return ""
}
