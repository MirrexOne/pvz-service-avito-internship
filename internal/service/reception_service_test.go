package service_test

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"testing"

	"pvz-service-avito-internship/internal/domain"
	"pvz-service-avito-internship/internal/service"
	"pvz-service-avito-internship/mocks"
)

func TestReceptionService_CreateReception(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockPVZRepo := mocks.NewPVZRepository(t)
	mockReceptionRepo := mocks.NewReceptionRepository(t)
	mockMetrics := mocks.NewMetricsCollector(t)

	receptionService := service.NewReceptionService(logger, mockPVZRepo, mockReceptionRepo, mockMetrics)

	ctx := context.Background()
	testPvzID := uuid.New()
	existingPvz := &domain.PVZ{ID: testPvzID, City: domain.Moscow}
	someError := errors.New("some db error")

	testCases := []struct {
		name          string
		pvzID         uuid.UUID
		setupMocks    func()
		expectedError error
	}{
		{
			name:  "Success",
			pvzID: testPvzID,
			setupMocks: func() {
				mockPVZRepo.On("GetByID", ctx, testPvzID).Return(existingPvz, nil).Once()
				// Ожидаем, что FindOpen вернет ошибку "не найдено"
				mockReceptionRepo.On("FindOpenByPVZID", ctx, testPvzID).Return(nil, domain.ErrNoOpenReception).Once()
				// Ожидаем вызов Create
				mockReceptionRepo.On("Create", ctx, mock.MatchedBy(func(rec *domain.Reception) bool {
					return rec.PVZID == testPvzID && rec.Status == domain.StatusInProgress && rec.ID != uuid.Nil
				})).Return(nil).Once()
				// Ожидаем метрику
				mockMetrics.On("IncReceptionsCreated").Return().Once()
			},
			expectedError: nil,
		},
		{
			name:  "Fail_PVZ_NotFound",
			pvzID: testPvzID,
			setupMocks: func() {
				mockPVZRepo.On("GetByID", ctx, testPvzID).Return(nil, domain.ErrNotFound).Once()
				// Другие вызовы не ожидаются
			},
			expectedError: domain.ErrValidation, // Ожидаем ошибку валидации, т.к. ПВЗ не существует
		},
		{
			name:  "Fail_PVZ_Repo_Error",
			pvzID: testPvzID,
			setupMocks: func() {
				mockPVZRepo.On("GetByID", ctx, testPvzID).Return(nil, someError).Once()
			},
			expectedError: domain.ErrDatabaseError,
		},
		{
			name:  "Fail_Reception_In_Progress",
			pvzID: testPvzID,
			setupMocks: func() {
				mockPVZRepo.On("GetByID", ctx, testPvzID).Return(existingPvz, nil).Once()
				// FindOpen возвращает существующую приемку (без ошибки)
				mockReceptionRepo.On("FindOpenByPVZID", ctx, testPvzID).Return(&domain.Reception{}, nil).Once()
				// Create и метрика не вызываются
			},
			expectedError: domain.ErrReceptionInProgress,
		},
		{
			name:  "Fail_FindOpen_Repo_Error",
			pvzID: testPvzID,
			setupMocks: func() {
				mockPVZRepo.On("GetByID", ctx, testPvzID).Return(existingPvz, nil).Once()
				mockReceptionRepo.On("FindOpenByPVZID", ctx, testPvzID).Return(nil, someError).Once()
			},
			expectedError: domain.ErrDatabaseError,
		},
		{
			name:  "Fail_Create_Repo_Error",
			pvzID: testPvzID,
			setupMocks: func() {
				mockPVZRepo.On("GetByID", ctx, testPvzID).Return(existingPvz, nil).Once()
				mockReceptionRepo.On("FindOpenByPVZID", ctx, testPvzID).Return(nil, domain.ErrNoOpenReception).Once()
				// Create возвращает ошибку
				mockReceptionRepo.On("Create", ctx, mock.AnythingOfType("*domain.Reception")).Return(someError).Once()
				// Метрика не вызывается
			},
			expectedError: domain.ErrDatabaseError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			rec, err := receptionService.CreateReception(ctx, tc.pvzID)

			if tc.expectedError != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Nil(t, rec)
			} else {
				require.NoError(t, err)
				require.NotNil(t, rec)
				assert.Equal(t, tc.pvzID, rec.PVZID)
				assert.Equal(t, domain.StatusInProgress, rec.Status)
				assert.NotEqual(t, uuid.Nil, rec.ID)
			}
			mockPVZRepo.AssertExpectations(t)
			mockReceptionRepo.AssertExpectations(t)
			mockMetrics.AssertExpectations(t)
		})
	}
}

func TestReceptionService_CloseReception(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockPVZRepo := mocks.NewPVZRepository(t) // Не используется напрямую, но нужен конструктору
	mockReceptionRepo := mocks.NewReceptionRepository(t)
	mockMetrics := mocks.NewMetricsCollector(t) // Не используется здесь

	receptionService := service.NewReceptionService(logger, mockPVZRepo, mockReceptionRepo, mockMetrics)

	ctx := context.Background()
	testPvzID := uuid.New()
	testReceptionID := uuid.New()
	openReception := &domain.Reception{ID: testReceptionID, PVZID: testPvzID, Status: domain.StatusInProgress}
	someError := errors.New("db error")

	testCases := []struct {
		name          string
		pvzID         uuid.UUID
		setupMocks    func()
		expectedError error
	}{
		{
			name:  "Success",
			pvzID: testPvzID,
			setupMocks: func() {
				// 1. FindOpen находит открытую приемку
				mockReceptionRepo.On("FindOpenByPVZID", ctx, testPvzID).Return(openReception, nil).Once()
				// 2. UpdateStatus вызывается с ID найденной приемки и статусом close
				mockReceptionRepo.On("UpdateStatus", ctx, testReceptionID, domain.StatusClosed).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name:  "Fail_No_Open_Reception",
			pvzID: testPvzID,
			setupMocks: func() {
				// FindOpen возвращает ошибку ErrNoOpenReception
				mockReceptionRepo.On("FindOpenByPVZID", ctx, testPvzID).Return(nil, domain.ErrNoOpenReception).Once()
				// UpdateStatus не вызывается
			},
			expectedError: domain.ErrReceptionClosed, // Ожидаем эту ошибку согласно ТЗ
		},
		{
			name:  "Fail_FindOpen_Repo_Error",
			pvzID: testPvzID,
			setupMocks: func() {
				// FindOpen возвращает другую ошибку БД
				mockReceptionRepo.On("FindOpenByPVZID", ctx, testPvzID).Return(nil, someError).Once()
			},
			expectedError: domain.ErrDatabaseError,
		},
		{
			name:  "Fail_UpdateStatus_Repo_Error",
			pvzID: testPvzID,
			setupMocks: func() {
				mockReceptionRepo.On("FindOpenByPVZID", ctx, testPvzID).Return(openReception, nil).Once()
				// UpdateStatus возвращает ошибку
				mockReceptionRepo.On("UpdateStatus", ctx, testReceptionID, domain.StatusClosed).Return(someError).Once()
			},
			expectedError: domain.ErrDatabaseError,
		},
		{
			name:  "Fail_UpdateStatus_NotFound", // Случай, если приемка удалилась между Find и Update
			pvzID: testPvzID,
			setupMocks: func() {
				mockReceptionRepo.On("FindOpenByPVZID", ctx, testPvzID).Return(openReception, nil).Once()
				// UpdateStatus возвращает ErrNotFound
				mockReceptionRepo.On("UpdateStatus", ctx, testReceptionID, domain.StatusClosed).Return(domain.ErrNotFound).Once()
			},
			expectedError: domain.ErrNotFound, // Ожидаем ErrNotFound
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			rec, err := receptionService.CloseReception(ctx, tc.pvzID)

			if tc.expectedError != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Nil(t, rec)
			} else {
				require.NoError(t, err)
				require.NotNil(t, rec)
				assert.Equal(t, testReceptionID, rec.ID)
				assert.Equal(t, domain.StatusClosed, rec.Status) // Статус должен быть close
			}
			mockReceptionRepo.AssertExpectations(t)
		})
	}
}
