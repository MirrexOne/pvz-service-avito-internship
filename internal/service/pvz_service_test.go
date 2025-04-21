package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"pvz-service-avito-internship/internal/domain"
	"pvz-service-avito-internship/internal/service"
	"pvz-service-avito-internship/mocks"
)

func TestPVZService_CreatePVZ(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockPVZRepo := mocks.NewPVZRepository(t)
	mockReceptionRepo := mocks.NewReceptionRepository(t)
	mockMetrics := mocks.NewMetricsCollector(t)

	pvzService := service.NewPVZService(logger, mockPVZRepo, mockReceptionRepo, mockMetrics)

	ctx := context.Background()
	testCityAllowed := domain.Moscow
	testCityNotAllowed := domain.City("Тверь")

	testCases := []struct {
		name          string
		inputCity     domain.City
		setupMocks    func()
		expectedError error
	}{
		{
			name:      "Success",
			inputCity: testCityAllowed,
			setupMocks: func() {
				mockPVZRepo.On("Create", mock.Anything, mock.MatchedBy(func(pvz *domain.PVZ) bool {
					return pvz.City == testCityAllowed && pvz.ID != uuid.Nil
				})).Return(nil).Once()
				mockMetrics.On("IncPVZCreated").Return().Once()
			},
			expectedError: nil,
		},
		{
			name:      "Fail_City_Not_Allowed",
			inputCity: testCityNotAllowed,
			setupMocks: func() {
			},
			expectedError: domain.ErrPVZCityNotAllowed,
		},
		{
			name:      "Fail_Repository_Error",
			inputCity: testCityAllowed,
			setupMocks: func() {
				dbError := errors.New("db is down")
				mockPVZRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.PVZ")).
					Return(dbError).Once()
			},
			expectedError: domain.ErrDatabaseError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			pvz, err := pvzService.CreatePVZ(ctx, tc.inputCity)

			if tc.expectedError != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.expectedError), "Expected error %v, got %v", tc.expectedError, err)
				assert.Nil(t, pvz)
			} else {
				require.NoError(t, err)
				require.NotNil(t, pvz)
				assert.Equal(t, tc.inputCity, pvz.City)
				assert.NotEqual(t, uuid.Nil, pvz.ID)
				assert.WithinDuration(t, time.Now(), pvz.RegistrationDate, 5*time.Second)
			}

			mockPVZRepo.AssertExpectations(t)
			mockMetrics.AssertExpectations(t)
			mockPVZRepo.Mock.ExpectedCalls = []*mock.Call{}
			mockMetrics.Mock.ExpectedCalls = []*mock.Call{}
		})
	}
}

func TestPVZService_ListPVZs(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockPVZRepo := mocks.NewPVZRepository(t)
	mockReceptionRepo := mocks.NewReceptionRepository(t)
	mockMetrics := mocks.NewMetricsCollector(t)

	pvzService := service.NewPVZService(logger, mockPVZRepo, mockReceptionRepo, mockMetrics)

	ctx := context.Background()
	testLimit := 10
	testPage := 1
	testOffset := 0
	testTotal := 5

	pvzID1 := uuid.New()
	pvzID2 := uuid.New()
	testIDs := []uuid.UUID{pvzID1, pvzID2}
	testPVZs := []domain.PVZ{
		{ID: pvzID1, City: domain.Moscow, RegistrationDate: time.Now().Add(-time.Hour)},
		{ID: pvzID2, City: domain.Kazan, RegistrationDate: time.Now()},
	}
	testReceptions := map[uuid.UUID][]domain.ReceptionWithProducts{
		pvzID1: {
			{
				Reception: domain.Reception{ID: uuid.New(), PVZID: pvzID1, Status: domain.StatusClosed},
				Products:  []domain.Product{{ID: uuid.New()}},
			},
		},
	}

	testCases := []struct {
		name                string
		limit               int
		page                int
		startDate           *time.Time
		endDate             *time.Time
		setupMocks          func()
		expectedResultCount int
		expectedTotal       int
		expectedError       error
	}{
		{
			name:  "Success_No_Filters",
			limit: testLimit,
			page:  testPage,
			setupMocks: func() {
				mockPVZRepo.On("ListIDsAndTotal", ctx, testLimit, testOffset, (*time.Time)(nil), (*time.Time)(nil)).
					Return(testIDs, testTotal, nil).Once()
				mockPVZRepo.On("GetByIDs", ctx, testIDs).Return(testPVZs, nil).Once()
				mockReceptionRepo.On("ListByPVZIDsAndDate", ctx, testIDs, (*time.Time)(nil), (*time.Time)(nil)).
					Return(testReceptions, nil).Once()
			},
			expectedResultCount: len(testIDs),
			expectedTotal:       testTotal,
			expectedError:       nil,
		},
		{
			name:  "Success_Empty_Page",
			limit: testLimit,
			page:  2,
			setupMocks: func() {
				offsetPage2 := (2 - 1) * testLimit
				mockPVZRepo.On("ListIDsAndTotal", ctx, testLimit, offsetPage2, (*time.Time)(nil), (*time.Time)(nil)).
					Return([]uuid.UUID{}, testTotal, nil).Once()

			},
			expectedResultCount: 0,
			expectedTotal:       testTotal,
			expectedError:       nil,
		},
		{
			name:  "Success_No_PVZs_Found",
			limit: testLimit,
			page:  testPage,
			setupMocks: func() {
				mockPVZRepo.On("ListIDsAndTotal", ctx, testLimit, testOffset, (*time.Time)(nil), (*time.Time)(nil)).
					Return([]uuid.UUID{}, 0, nil).Once()

			},
			expectedResultCount: 0,
			expectedTotal:       0,
			expectedError:       nil,
		},
		{
			name:  "Fail_ListIDsAndTotal_Error",
			limit: testLimit,
			page:  testPage,
			setupMocks: func() {
				dbError := errors.New("db error list ids")
				mockPVZRepo.On("ListIDsAndTotal", ctx, testLimit, testOffset, (*time.Time)(nil), (*time.Time)(nil)).
					Return(nil, 0, dbError).Once()
			},
			expectedResultCount: 0,
			expectedTotal:       0,
			expectedError:       domain.ErrDatabaseError,
		},
		{
			name:  "Fail_GetByIDs_Error",
			limit: testLimit,
			page:  testPage,
			setupMocks: func() {
				dbError := errors.New("db error get by ids")
				mockPVZRepo.On("ListIDsAndTotal", ctx, testLimit, testOffset, (*time.Time)(nil), (*time.Time)(nil)).
					Return(testIDs, testTotal, nil).Once()
				mockPVZRepo.On("GetByIDs", ctx, testIDs).Return(nil, dbError).Once()

			},
			expectedResultCount: 0,
			expectedTotal:       0,
			expectedError:       domain.ErrDatabaseError,
		},
		{
			name:  "Fail_ListReceptions_Error",
			limit: testLimit,
			page:  testPage,
			setupMocks: func() {
				dbError := errors.New("db error list receptions")
				mockPVZRepo.On("ListIDsAndTotal", ctx, testLimit, testOffset, (*time.Time)(nil), (*time.Time)(nil)).
					Return(testIDs, testTotal, nil).Once()
				mockPVZRepo.On("GetByIDs", ctx, testIDs).Return(testPVZs, nil).Once()
				mockReceptionRepo.On("ListByPVZIDsAndDate", ctx, testIDs, (*time.Time)(nil), (*time.Time)(nil)).
					Return(nil, dbError).Once()
			},
			expectedResultCount: 0,
			expectedTotal:       0,
			expectedError:       domain.ErrDatabaseError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			results, total, err := pvzService.ListPVZs(ctx, tc.limit, tc.page, tc.startDate, tc.endDate)

			if tc.expectedError != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Nil(t, results)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedTotal, total)
				require.Len(t, results, tc.expectedResultCount)
				if len(results) > 0 {
					assert.Equal(t, testPVZs[0].ID, results[0].PVZ.ID)
					if tc.name == "Success_No_Filters" {
						assert.Len(t, results[0].Receptions, 1)
						assert.Len(t, results[1].Receptions, 0)
					}
				}
			}

			mockPVZRepo.AssertExpectations(t)
			mockReceptionRepo.AssertExpectations(t)
		})
	}
}
