package http

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"pvz-service-avito-internship/internal/domain"
	"pvz-service-avito-internship/internal/handler/http/response"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestBaseHandler_HandleError(t *testing.T) {
	logBuffer := new(bytes.Buffer)
	logger := slog.New(slog.NewJSONHandler(logBuffer, &slog.HandlerOptions{Level: slog.LevelWarn}))
	handler := NewBaseHandler(logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		handler.handleError(c, "test_operation", domain.ErrNotFound)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.JSONEq(t, `{"message":"Resource not found"}`, rec.Body.String())

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, `"level":"WARN"`)
	assert.Contains(t, logOutput, `"msg":"Client error mapped"`)
}

func TestBaseHandler_ParseUUID(t *testing.T) {
	logBuffer := new(bytes.Buffer)
	logger := slog.New(slog.NewJSONHandler(logBuffer, &slog.HandlerOptions{Level: slog.LevelWarn}))
	handler := NewBaseHandler(logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test/:id", func(c *gin.Context) {
		_, err := handler.parseUUID(c, "id")
		if err != nil {
			handler.handleError(c, "parse_uuid", err)
			return
		}
		response.SendSuccess(c, http.StatusOK, gin.H{"message": "Valid UUID"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test/123e4567-e89b-12d3-a456-426614174000", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"message":"Valid UUID"}`, rec.Body.String())

	req = httptest.NewRequest(http.MethodGet, "/test/invalid-uuid", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.JSONEq(t, `{"message":"Invalid request data: validation failed"}`, rec.Body.String())

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, `"level":"WARN"`)
	assert.Contains(t, logOutput, `"msg":"Client error mapped"`)
}

func TestBaseHandler_ParseIntQuery(t *testing.T) {
	logBuffer := new(bytes.Buffer)
	logger := slog.New(slog.NewJSONHandler(logBuffer, &slog.HandlerOptions{Level: slog.LevelWarn}))
	handler := NewBaseHandler(logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		value, err := handler.parseIntQuery(c, "page", 1)
		if err != nil {
			handler.handleError(c, "parse_int_query", err)
			return
		}
		response.SendSuccess(c, http.StatusOK, gin.H{"page": value})
	})

	req := httptest.NewRequest(http.MethodGet, "/test?page=2", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"page":2}`, rec.Body.String())

	req = httptest.NewRequest(http.MethodGet, "/test?page=invalid", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.JSONEq(t, `{"message":"Invalid request data: validation failed"}`, rec.Body.String())

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, `"level":"WARN"`)
	assert.Contains(t, logOutput, `"msg":"Client error mapped"`)
}

func TestBaseHandler_ParseStringQuery(t *testing.T) {
	logBuffer := new(bytes.Buffer)
	logger := slog.New(slog.NewJSONHandler(logBuffer, &slog.HandlerOptions{Level: slog.LevelWarn}))
	handler := NewBaseHandler(logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		value, _ := handler.parseStringQuery(c, "name", "default")
		response.SendSuccess(c, http.StatusOK, gin.H{"name": value})
	})

	req := httptest.NewRequest(http.MethodGet, "/test?name=test", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"name":"test"}`, rec.Body.String())

	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"name":"default"}`, rec.Body.String())

	logOutput := logBuffer.String()
	assert.NotContains(t, logOutput, `"level":"WARN"`)
}

// parseStringQuery извлекает строковый параметр из запроса или возвращает значение по умолчанию.
func (h *BaseHandler) parseStringQuery(c *gin.Context, key string, defaultValue string) (string, error) {
	value := c.Query(key)
	if value == "" {
		return defaultValue, nil
	}
	return value, nil
}

func TestBaseHandler_parseIntQuery(t *testing.T) {
	logBuffer := new(bytes.Buffer)
	logger := slog.New(slog.NewJSONHandler(logBuffer, &slog.HandlerOptions{Level: slog.LevelWarn}))
	handler := NewBaseHandler(logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		value, err := handler.parseIntQuery(c, "page", 1)
		if err != nil {
			handler.handleError(c, "parse_int_query", err)
			return
		}
		response.SendSuccess(c, http.StatusOK, gin.H{"page": value})
	})

	req := httptest.NewRequest(http.MethodGet, "/test?page=2", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"page":2}`, rec.Body.String())

	req = httptest.NewRequest(http.MethodGet, "/test?page=invalid", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.JSONEq(t, `{"message":"Invalid request data: validation failed"}`, rec.Body.String())

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, `"level":"WARN"`)
	assert.Contains(t, logOutput, `"msg":"Client error mapped"`)
}

func TestBaseHandler_parseBoolQuery(t *testing.T) {
	logBuffer := new(bytes.Buffer)
	logger := slog.New(slog.NewJSONHandler(logBuffer, &slog.HandlerOptions{Level: slog.LevelWarn}))
	handler := NewBaseHandler(logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		value, err := handler.parseBoolQuery(c, "active", false)
		if err != nil {
			handler.handleError(c, "parse_bool_query", domain.ErrInvalidRequest)
			return
		}
		response.SendSuccess(c, http.StatusOK, gin.H{"active": value})
	})

	req := httptest.NewRequest(http.MethodGet, "/test?active=true", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"active":true}`, rec.Body.String())

	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"active":false}`, rec.Body.String())

	logOutput := logBuffer.String()
	assert.NotContains(t, logOutput, `"level":"WARN"`)
}
