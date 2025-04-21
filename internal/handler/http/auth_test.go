package http_test

import (
	"bytes"
	"context"
	"encoding/json"
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

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) DummyLogin(ctx context.Context, role domain.UserRole) (string, error) {
	args := m.Called(ctx, role)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) Register(ctx context.Context, email string, password string, role domain.UserRole) (*domain.User, error) {
	args := m.Called(ctx, email, password, role)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, email string, password string) (string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.Error(1)
}

func TestIntegration_PostProducts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()

	mockService := new(MockProductService)
	handler := httpHandler.NewProductHandler(logger, mockService)

	router := gin.New()
	router.POST("/products", handler.PostProducts)

	t.Run("успешное добавление продукта", func(t *testing.T) {
		pvzID := "123e4567-e89b-12d3-a456-426614174000"
		product := &domain.Product{
			ID:   uuid.New(),
			Type: domain.ProductType("электроника"),
		}

		mockService.On("AddProduct", mock.Anything, uuid.MustParse(pvzID), domain.ProductType("электроника")).
			Return(product, nil)

		body := map[string]interface{}{
			"pvzId": pvzID,
			"type":  "электроника",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("ошибка при некорректном JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader([]byte(`{"pvzId":123}`)))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAuthHandler_PostRegister(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()

	mockAuthService := new(MockAuthService)
	handler := httpHandler.NewAuthHandler(logger, mockAuthService)

	router := gin.New()
	router.POST("/register", handler.PostRegister)

	t.Run("успешная регистрация", func(t *testing.T) {
		mockAuthService.
			On("Register", mock.Anything, "test@example.com", "password123", domain.UserRole("user")).
			Return(&domain.User{
				ID:   uuid.New(),
				Role: "user",
			}, nil)

		body := map[string]interface{}{
			"email":    "test@example.com",
			"password": "password123",
			"role":     "user",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockAuthService.AssertExpectations(t)
	})

	t.Run("ошибка при некорректном JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader([]byte(`{"email":123}`)))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAuthHandler_PostLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()

	mockAuthService := new(MockAuthService)
	handler := httpHandler.NewAuthHandler(logger, mockAuthService)

	router := gin.New()
	router.POST("/login", handler.PostLogin)

	t.Run("успешный вход", func(t *testing.T) {
		mockAuthService.
			On("Login", mock.Anything, "test@example.com", "password123").
			Return("login-token", nil)

		body := map[string]interface{}{
			"email":    "test@example.com",
			"password": "password123",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `"login-token"`, w.Body.String())
		mockAuthService.AssertExpectations(t)
	})

	t.Run("ошибка при некорректном JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader([]byte(`{"email":123}`)))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ошибка сервиса", func(t *testing.T) {
		mockAuthService.
			On("Login", mock.Anything, "test@example.com", "password123").
			Return("", errors.New("service error"))

		body := map[string]interface{}{
			"email":    "test@example.com",
			"password": "password123",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `"login-token"`, w.Body.String())
		mockAuthService.AssertExpectations(t)
	})

	t.Run("ошибка при пустом теле запроса", func(t *testing.T) {
		mockAuthService := new(MockAuthService)
		handler := httpHandler.NewAuthHandler(logger, mockAuthService)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader([]byte(``)))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.PostLogin(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ошибка при отсутствии password", func(t *testing.T) {
		mockAuthService.
			On("Login", mock.Anything, "test@example.com", "").
			Return("", domain.ErrPassIsRequired)

		body := map[string]interface{}{
			"email": "test@example.com",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"message":"password is required"}`, w.Body.String())
	})
}
