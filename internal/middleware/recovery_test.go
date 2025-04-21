package middleware_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pvz-service-avito-internship/internal/handler/http/api"
	mw "pvz-service-avito-internship/internal/middleware"
)

func TestRecoveryMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logBuffer := new(bytes.Buffer)
	testHandler := slog.NewJSONHandler(logBuffer, &slog.HandlerOptions{Level: slog.LevelError})
	testLogger := slog.New(testHandler)

	recoveryMiddleware := mw.Recovery(testLogger)

	router := gin.New()
	router.NoMethod(func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	})
	router.Use(recoveryMiddleware)

	router.GET("/panic", func(c *gin.Context) {
		panic("something went very wrong")
	})

	router.GET("/ok", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	t.Run("Panic Recovered", func(t *testing.T) {
		logBuffer.Reset()
		req := httptest.NewRequest(http.MethodGet, "/panic", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		var errResp api.Error
		err := json.Unmarshal(rr.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, "Internal server error", errResp.Message)

		logOutput := logBuffer.String()
		assert.Contains(t, logOutput, `"level":"ERROR"`)
		assert.Contains(t, logOutput, `"msg":"Panic recovered"`)
		assert.Contains(t, logOutput, `"error":"something went very wrong"`)
		assert.Contains(t, logOutput, `"stack":`)
	})

	t.Run("No Panic", func(t *testing.T) {
		logBuffer.Reset()
		req := httptest.NewRequest(http.MethodGet, "/ok", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Empty(t, rr.Body.String())

		logOutput := logBuffer.String()
		assert.NotContains(t, logOutput, "Panic recovered")
	})
}

func TestRecoveryMiddleware_EmptyRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logBuffer := new(bytes.Buffer)
	testHandler := slog.NewJSONHandler(logBuffer, &slog.HandlerOptions{Level: slog.LevelError})
	testLogger := slog.New(testHandler)

	recoveryMiddleware := mw.Recovery(testLogger)

	router := gin.New()
	router.NoMethod(func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	})
	router.Use(recoveryMiddleware)

	router.GET("/empty", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	t.Run("Empty Request", func(t *testing.T) {
		logBuffer.Reset()
		req := httptest.NewRequest(http.MethodGet, "/empty", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Empty(t, rr.Body.String())

		logOutput := logBuffer.String()
		assert.NotContains(t, logOutput, "Panic recovered")
	})

	t.Run("Empty Request with Body", func(t *testing.T) {
		logBuffer.Reset()
		req := httptest.NewRequest(http.MethodGet, "/empty", bytes.NewBuffer([]byte{}))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Empty(t, rr.Body.String())

		logOutput := logBuffer.String()
		assert.NotContains(t, logOutput, "Panic recovered")
	})

	t.Run("Empty Request with Invalid Body", func(t *testing.T) {
		logBuffer.Reset()
		req := httptest.NewRequest(http.MethodGet, "/empty", bytes.NewBuffer([]byte("{invalid json}")))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Empty(t, rr.Body.String())

		logOutput := logBuffer.String()
		assert.NotContains(t, logOutput, "Panic recovered")
	})
}

func TestRecoveryMiddleware_EmptyRequestWithBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logBuffer := new(bytes.Buffer)
	testHandler := slog.NewJSONHandler(logBuffer, &slog.HandlerOptions{Level: slog.LevelError})
	testLogger := slog.New(testHandler)

	recoveryMiddleware := mw.Recovery(testLogger)

	router := gin.New()
	router.NoMethod(func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	})
	router.Use(recoveryMiddleware)

	router.GET("/empty", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	t.Run("Empty Request with Body", func(t *testing.T) {
		logBuffer.Reset()
		req := httptest.NewRequest(http.MethodGet, "/empty", bytes.NewBuffer([]byte{}))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Empty(t, rr.Body.String())

		logOutput := logBuffer.String()
		assert.NotContains(t, logOutput, "Panic recovered")
	})

	t.Run("Empty Request with Invalid Body", func(t *testing.T) {
		logBuffer.Reset()
		req := httptest.NewRequest(http.MethodGet, "/empty", bytes.NewBuffer([]byte("{invalid json}")))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Empty(t, rr.Body.String())

		logOutput := logBuffer.String()
		assert.NotContains(t, logOutput, "Panic recovered")
	})
}
