package http_test

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"pvz-service-avito-internship/internal/domain"
	httpHandler "pvz-service-avito-internship/internal/handler/http"
	"testing"
)

func TestIntegration_PostProducts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()

	// Создаем моковый сервис и обработчик
	mockService := &MockProductService{}
	handler := httpHandler.NewProductHandler(logger, mockService)

	// Создаем роутер Gin и регистрируем маршрут
	router := gin.Default()
	router.POST("/products", handler.PostProducts)

	t.Run("успешное добавление продукта", func(t *testing.T) {
		pvzID := uuid.New()

		mockService.On("AddProduct", mock.Anything, pvzID, domain.ProductType("Electronics")).
			Return(&domain.Product{ID: uuid.New()}, nil)

		requestBody := map[string]interface{}{
			"pvzId": pvzID.String(),
			"type":  "Electronics",
		}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("ошибка при некорректном JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader([]byte(`{"pvzId":123}`)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ошибка при пустом значении type", func(t *testing.T) {
		pvzID := uuid.New()
		requestBody := map[string]interface{}{
			"pvzId": pvzID.String(),
			"type":  "",
		}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ошибка при недопустимом значении type", func(t *testing.T) {
		pvzID := uuid.New()
		requestBody := map[string]interface{}{
			"pvzId": pvzID.String(),
			"type":  "InvalidType",
		}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
