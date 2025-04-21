package http_test

import (
	"errors"
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

func TestProductHandler_PostProducts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()

	t.Run("успешное добавление продукта", func(t *testing.T) {
		mockService := new(MockProductService)
		handler := httpHandler.NewProductHandler(logger, mockService)

		pvzID := uuid.New()
		product := &domain.Product{
			ID:   uuid.New(),
			Type: domain.ProductType("electronics"),
		}

		mockService.On("AddProduct", mock.Anything, pvzID, domain.ProductType("электроника")).
			Return(product, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/products", stringToReader(`{"pvzId":"`+pvzID.String()+`","type":"электроника"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.PostProducts(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("ошибка при некорректном JSON", func(t *testing.T) {
		mockService := new(MockProductService)
		handler := httpHandler.NewProductHandler(logger, mockService)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/products", stringToReader(`{"pvzId":123}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.PostProducts(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ошибка при отсутствии pvzId", func(t *testing.T) {
		mockService := new(MockProductService)
		handler := httpHandler.NewProductHandler(logger, mockService)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/products", stringToReader(`{"type":"Electronics"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.PostProducts(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ошибка сервиса при добавлении продукта", func(t *testing.T) {
		mockService := new(MockProductService)
		handler := httpHandler.NewProductHandler(logger, mockService)

		pvzID := uuid.New()

		mockService.On("AddProduct", mock.Anything, pvzID, domain.ProductType("электроника")).
			Return(nil, errors.New("service error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/products", stringToReader(`{"pvzId":"`+pvzID.String()+`","type":"электроника"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.PostProducts(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("ошибка при пустом теле запроса", func(t *testing.T) {
		mockService := new(MockProductService)
		handler := httpHandler.NewProductHandler(logger, mockService)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/products", stringToReader(``))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.PostProducts(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ошибка при некорректном UUID в pvzId", func(t *testing.T) {
		mockService := new(MockProductService)
		handler := httpHandler.NewProductHandler(logger, mockService)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/products", stringToReader(`{"pvzId":"invalid-uuid","type":"Electronics"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.PostProducts(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ошибка при пустом поле type", func(t *testing.T) {
		mockService := new(MockProductService)
		handler := httpHandler.NewProductHandler(logger, mockService)

		pvzID := uuid.New()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/products", stringToReader(`{"pvzId":"`+pvzID.String()+`","type":""}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.PostProducts(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ошибка при некорректном значении type", func(t *testing.T) {
		mockService := new(MockProductService)
		handler := httpHandler.NewProductHandler(logger, mockService)

		pvzID := uuid.New()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/products", stringToReader(`{"pvzId":"`+pvzID.String()+`","type":"InvalidType"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.PostProducts(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ошибка сервиса при ошибке базы данных", func(t *testing.T) {
		mockService := new(MockProductService)
		handler := httpHandler.NewProductHandler(logger, mockService)

		pvzID := uuid.New()

		mockService.On("AddProduct", mock.Anything, pvzID, domain.ProductType("электроника")).
			Return(nil, errors.New("database error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/products", stringToReader(`{"pvzId":"`+pvzID.String()+`","type":"электроника"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.PostProducts(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("ошибка при недопустимом значении type", func(t *testing.T) {
		mockService := new(MockProductService)
		handler := httpHandler.NewProductHandler(logger, mockService)

		pvzID := uuid.New()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/products", stringToReader(`{"pvzId":"`+pvzID.String()+`","type":"InvalidType"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.PostProducts(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ошибка при недопустимом значении type", func(t *testing.T) {
		mockService := new(MockProductService)
		handler := httpHandler.NewProductHandler(logger, mockService)

		pvzID := uuid.New()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/products", stringToReader(`{"pvzId":"`+pvzID.String()+`","type":"InvalidType"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.PostProducts(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ошибка при пустом pvzId", func(t *testing.T) {
		mockService := new(MockProductService)
		handler := httpHandler.NewProductHandler(logger, mockService)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/products", stringToReader(`{"pvzId":"","type":"Electronics"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.PostProducts(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ошибка при пустом значении type", func(t *testing.T) {
		mockService := new(MockProductService)
		handler := httpHandler.NewProductHandler(logger, mockService)

		pvzID := uuid.New()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/products", stringToReader(`{"pvzId":"`+pvzID.String()+`","type":""}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.PostProducts(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ошибка при недопустимом значении type", func(t *testing.T) {
		mockService := new(MockProductService)
		handler := httpHandler.NewProductHandler(logger, mockService)

		pvzID := uuid.New()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/products", stringToReader(`{"pvzId":"`+pvzID.String()+`","type":"InvalidType"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.PostProducts(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ошибка сервиса при добавлении продукта", func(t *testing.T) {
		mockService := new(MockProductService)
		handler := httpHandler.NewProductHandler(logger, mockService)

		pvzID := uuid.New()

		mockService.On("AddProduct", mock.Anything, pvzID, domain.ProductType("электроника")).
			Return(nil, errors.New("service error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/products", stringToReader(`{"pvzId":"`+pvzID.String()+`","type":"электроника"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.PostProducts(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}
