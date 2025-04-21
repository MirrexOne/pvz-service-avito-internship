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
	"github.com/stretchr/testify/require"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"pvz-service-avito-internship/internal/domain"
	httpHandler "pvz-service-avito-internship/internal/handler/http"
	"testing"
)

type MockReceptionService struct {
	mock.Mock
}

func (m *MockReceptionService) CloseReception(ctx context.Context, pvzID uuid.UUID) (*domain.Reception, error) {
	args := m.Called(ctx, pvzID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Reception), args.Error(1)
}

func (m *MockReceptionService) CreateReception(ctx context.Context, pvzID uuid.UUID) (*domain.Reception, error) {
	args := m.Called(ctx, pvzID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Reception), args.Error(1)
}

func TestReceptionHandler_CreationReception(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()

	t.Run("успешное создание", func(t *testing.T) {
		mockService := new(MockReceptionService)
		handler := httpHandler.NewReceptionHandler(logger, mockService)

		pvzID := uuid.New()
		receptionID := uuid.New()
		expectedReception := &domain.Reception{
			ID:     receptionID,
			PVZID:  pvzID,
			Status: domain.StatusInProgress,
		}

		mockService.On("CreateReception", mock.Anything, pvzID).Return(expectedReception, nil)

		reqBody := map[string]interface{}{
			"pvzId": pvzID.String(),
		}
		jsonBody, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/receptions", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.PostReceptions(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, receptionID.String(), response["id"])
		assert.Equal(t, pvzID.String(), response["pvzId"])
	})

	t.Run("невалидный JSON в запросе", func(t *testing.T) {
		mockService := new(MockReceptionService)
		handler := httpHandler.NewReceptionHandler(logger, mockService)

		req, err := http.NewRequest(http.MethodPost, "/receptions", bytes.NewBuffer([]byte("invalid json")))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.PostReceptions(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("пустой UUID в запросе", func(t *testing.T) {
		mockService := new(MockReceptionService)
		handler := httpHandler.NewReceptionHandler(logger, mockService)

		reqBody := map[string]interface{}{
			"pvzId": uuid.Nil.String(),
		}
		jsonBody, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/receptions", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.PostReceptions(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ошибка сервиса", func(t *testing.T) {
		mockService := new(MockReceptionService)
		handler := httpHandler.NewReceptionHandler(logger, mockService)

		pvzID := uuid.New()
		mockService.On("CreateReception", mock.Anything, pvzID).
			Return(nil, errors.New("service error"))

		reqBody := map[string]interface{}{
			"pvzId": pvzID.String(),
		}
		jsonBody, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/receptions", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.PostReceptions(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestReceptionHandler_CloseLastReception(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()

	t.Run("успешное закрытие", func(t *testing.T) {
		mockService := new(MockReceptionService)
		handler := httpHandler.NewReceptionHandler(logger, mockService)

		pvzID := uuid.New()
		receptionID := uuid.New()
		expectedReception := &domain.Reception{
			ID:     receptionID,
			PVZID:  pvzID,
			Status: domain.StatusClosed,
		}

		mockService.On("CloseReception", mock.Anything, pvzID).Return(expectedReception, nil)

		req, err := http.NewRequest(http.MethodPost, "/receptions/"+pvzID.String()+"/close_last", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		c.Params = []gin.Param{{Key: "pvzId", Value: pvzID.String()}}

		handler.CloseLastReception(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, receptionID.String(), response["id"])
		assert.Equal(t, pvzID.String(), response["pvzId"])
		assert.Equal(t, string(domain.StatusClosed), response["status"])
	})
}
