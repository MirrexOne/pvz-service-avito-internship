package http_test

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"pvz-service-avito-internship/internal/domain"
	httpHandler "pvz-service-avito-internship/internal/handler/http"
	"strings"
	"testing"
	"time"
)

type MockPVZService struct {
	mock.Mock
}

func (m *MockPVZService) CreatePVZ(ctx context.Context, city domain.City) (*domain.PVZ, error) {
	args := m.Called(ctx, city)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PVZ), args.Error(1)
}

func (m *MockPVZService) ListPVZs(ctx context.Context, limit, page int, startDate, endDate *time.Time) ([]domain.PVZWithDetails, int, error) {
	args := m.Called(ctx, limit, page, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]domain.PVZWithDetails), args.Int(1), args.Error(2)
}

func TestPVZHandler_GetPvz(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()

	t.Run("успешное получение списка ПВЗ", func(t *testing.T) {
		mockService := new(MockPVZService)
		handler := httpHandler.NewPVZHandler(logger, mockService, nil, nil)

		pvzList := []domain.PVZWithDetails{
			{PVZ: domain.PVZ{ID: uuid.New(), City: "Moscow"}},
			{PVZ: domain.PVZ{ID: uuid.New(), City: "Saint Petersburg"}},
		}

		mockService.On("ListPVZs", mock.Anything, 10, 1, (*time.Time)(nil), (*time.Time)(nil)).
			Return(pvzList, 2, nil)

		req, err := http.NewRequest(http.MethodGet, "/pvz?page=1&limit=10", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.GetPvz(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("ошибка сервиса", func(t *testing.T) {
		mockService := new(MockPVZService)
		handler := httpHandler.NewPVZHandler(logger, mockService, nil, nil)

		mockService.On("ListPVZs", mock.Anything, 10, 1, (*time.Time)(nil), (*time.Time)(nil)).
			Return([]domain.PVZWithDetails{}, 0, errors.New("service error"))

		req, err := http.NewRequest(http.MethodGet, "/pvz?page=1&limit=10", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.GetPvz(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}
func TestPVZHandler_GetPvz_InvalidParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()

	t.Run("некорректный параметр page", func(t *testing.T) {
		mockService := new(MockPVZService)
		handler := httpHandler.NewPVZHandler(logger, mockService, nil, nil)

		req, err := http.NewRequest(http.MethodGet, "/pvz?page=invalid&limit=10", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.GetPvz(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("limit превышает максимальное значение", func(t *testing.T) {
		mockService := new(MockPVZService)
		handler := httpHandler.NewPVZHandler(logger, mockService, nil, nil)

		pvzList := []domain.PVZWithDetails{
			{PVZ: domain.PVZ{ID: uuid.New(), City: "Moscow"}},
		}

		mockService.On("ListPVZs", mock.Anything, 30, 1, (*time.Time)(nil), (*time.Time)(nil)).
			Return(pvzList, 1, nil)

		req, err := http.NewRequest(http.MethodGet, "/pvz?page=1&limit=50", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.GetPvz(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}

type MockProductService struct {
	mock.Mock
}

func (m *MockProductService) AddProduct(ctx context.Context, pvzID uuid.UUID, productType domain.ProductType) (*domain.Product, error) {
	args := m.Called(ctx, pvzID, productType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *MockProductService) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {
	args := m.Called(ctx, pvzID)
	return args.Error(0)
}

func TestPVZHandler_PostPvz(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()

	t.Run("успешное создание ПВЗ", func(t *testing.T) {
		mockService := new(MockPVZService)
		handler := httpHandler.NewPVZHandler(logger, mockService, nil, nil)

		newPVZ := &domain.PVZ{
			ID:   uuid.New(),
			City: "Moscow",
		}

		mockService.On("CreatePVZ", mock.Anything, domain.City("Moscow")).
			Return(newPVZ, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/pvz", stringToReader(`{"city":"Moscow"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.PostPvz(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestPVZHandler_CloseLastReception(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()

	t.Run("успешное закрытие приема", func(t *testing.T) {
		mockReceptionService := new(MockReceptionService)
		handler := httpHandler.NewPVZHandler(logger, nil, mockReceptionService, nil)

		pvzID := uuid.New()
		reception := &domain.Reception{
			ID:    uuid.New(),
			PVZID: pvzID,
		}

		mockReceptionService.On("CloseReception", mock.Anything, pvzID).
			Return(reception, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "pvzId", Value: pvzID.String()}}
		c.Request = httptest.NewRequest(http.MethodPost, "/pvz/"+pvzID.String()+"/reception/close", nil)

		handler.CloseLastReception(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockReceptionService.AssertExpectations(t)
	})
}

func TestPVZHandler_DeleteLastProduct(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()

	t.Run("успешное удаление последнего продукта", func(t *testing.T) {
		mockProductService := new(MockProductService)
		handler := httpHandler.NewPVZHandler(logger, nil, nil, mockProductService)

		pvzID := uuid.New()
		mockProductService.On("DeleteLastProduct", mock.Anything, pvzID).
			Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "pvzId", Value: pvzID.String()}}
		c.Request = httptest.NewRequest(http.MethodDelete, "/pvz/"+pvzID.String()+"/product", nil)

		handler.DeleteLastProduct(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockProductService.AssertExpectations(t)
	})
}

func stringToReader(s string) *strings.Reader {
	return strings.NewReader(s)
}
