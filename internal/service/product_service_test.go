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

func TestProductService_AddProduct(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockReceptionRepo := mocks.NewReceptionRepository(t)
	mockProductRepo := mocks.NewProductRepository(t)
	mockMetrics := mocks.NewMetricsCollector(t)

	productService := service.NewProductService(logger, mockReceptionRepo, mockProductRepo, mockMetrics)

	ctx := context.Background()
	testPvzID := uuid.New()
	testReceptionID := uuid.New()
	openReception := &domain.Reception{ID: testReceptionID, PVZID: testPvzID, Status: domain.StatusInProgress}
	validProductType := domain.TypeElectronics
	invalidProductType := domain.ProductType("gadgets")
	someError := errors.New("db error")

	testCases := []struct {
		name          string
		pvzID         uuid.UUID
		productType   domain.ProductType
		setupMocks    func()
		expectedError error
	}{
		{
			name:        "Success",
			pvzID:       testPvzID,
			productType: validProductType,
			setupMocks: func() {
				mockReceptionRepo.On("FindOpenByPVZID", ctx, testPvzID).Return(openReception, nil).Once()
				mockProductRepo.On("Create", ctx, mock.MatchedBy(func(p *domain.Product) bool {
					return p.ReceptionID == testReceptionID && p.Type == validProductType && p.ID != uuid.Nil
				})).Return(nil).Once()
				mockMetrics.On("IncProductsAdded").Return().Once()
			},
			expectedError: nil,
		},
		{
			name:          "Fail_Invalid_Product_Type",
			pvzID:         testPvzID,
			productType:   invalidProductType,
			setupMocks:    func() {},
			expectedError: domain.ErrValidation,
		},
		{
			name:        "Fail_No_Open_Reception",
			pvzID:       testPvzID,
			productType: validProductType,
			setupMocks: func() {
				mockReceptionRepo.On("FindOpenByPVZID", ctx, testPvzID).Return(nil, domain.ErrNoOpenReception).Once()
			},
			expectedError: domain.ErrNoOpenReception,
		},
		{
			name:        "Fail_FindOpen_Repo_Error",
			pvzID:       testPvzID,
			productType: validProductType,
			setupMocks: func() {
				mockReceptionRepo.On("FindOpenByPVZID", ctx, testPvzID).Return(nil, someError).Once()
			},
			expectedError: domain.ErrDatabaseError,
		},
		{
			name:        "Fail_Create_Repo_Error",
			pvzID:       testPvzID,
			productType: validProductType,
			setupMocks: func() {
				mockReceptionRepo.On("FindOpenByPVZID", ctx, testPvzID).Return(openReception, nil).Once()
				mockProductRepo.On("Create", ctx, mock.AnythingOfType("*domain.Product")).Return(someError).Once()
			},
			expectedError: domain.ErrDatabaseError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			prod, err := productService.AddProduct(ctx, tc.pvzID, tc.productType)

			if tc.expectedError != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Nil(t, prod)
			} else {
				require.NoError(t, err)
				require.NotNil(t, prod)
				assert.Equal(t, tc.productType, prod.Type)
				assert.Equal(t, testReceptionID, prod.ReceptionID)
				assert.NotEqual(t, uuid.Nil, prod.ID)
			}
			mockReceptionRepo.AssertExpectations(t)
			mockProductRepo.AssertExpectations(t)
			mockMetrics.AssertExpectations(t)
		})
	}
}

func TestProductService_DeleteLastProduct(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockReceptionRepo := mocks.NewReceptionRepository(t)
	mockProductRepo := mocks.NewProductRepository(t)
	mockMetrics := mocks.NewMetricsCollector(t)

	productService := service.NewProductService(logger, mockReceptionRepo, mockProductRepo, mockMetrics)

	ctx := context.Background()
	testPvzID := uuid.New()
	testReceptionID := uuid.New()
	testProductID := uuid.New()
	openReception := &domain.Reception{ID: testReceptionID, PVZID: testPvzID, Status: domain.StatusInProgress}
	lastProduct := &domain.Product{ID: testProductID, ReceptionID: testReceptionID, Type: domain.TypeShoes, DateTime: time.Now()}
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
				mockReceptionRepo.On("FindOpenByPVZID", ctx, testPvzID).Return(openReception, nil).Once()
				mockProductRepo.On("FindLastByReceptionID", ctx, testReceptionID).Return(lastProduct, nil).Once()
				mockProductRepo.On("DeleteByID", ctx, testProductID).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name:  "Fail_No_Open_Reception",
			pvzID: testPvzID,
			setupMocks: func() {
				mockReceptionRepo.On("FindOpenByPVZID", ctx, testPvzID).Return(nil, domain.ErrNoOpenReception).Once()
			},
			expectedError: domain.ErrNoOpenReception,
		},
		{
			name:  "Fail_No_Products_To_Delete",
			pvzID: testPvzID,
			setupMocks: func() {
				mockReceptionRepo.On("FindOpenByPVZID", ctx, testPvzID).Return(openReception, nil).Once()
				mockProductRepo.On("FindLastByReceptionID", ctx, testReceptionID).Return(nil, domain.ErrNoProductsToDelete).Once()
			},
			expectedError: domain.ErrNoProductsToDelete,
		},
		{
			name:  "Fail_FindOpen_Repo_Error",
			pvzID: testPvzID,
			setupMocks: func() {
				mockReceptionRepo.On("FindOpenByPVZID", ctx, testPvzID).Return(nil, someError).Once()
			},
			expectedError: domain.ErrDatabaseError,
		},
		{
			name:  "Fail_FindLast_Repo_Error",
			pvzID: testPvzID,
			setupMocks: func() {
				mockReceptionRepo.On("FindOpenByPVZID", ctx, testPvzID).Return(openReception, nil).Once()
				mockProductRepo.On("FindLastByReceptionID", ctx, testReceptionID).Return(nil, someError).Once()
			},
			expectedError: domain.ErrDatabaseError,
		},
		{
			name:  "Fail_DeleteByID_Repo_Error",
			pvzID: testPvzID,
			setupMocks: func() {
				mockReceptionRepo.On("FindOpenByPVZID", ctx, testPvzID).Return(openReception, nil).Once()
				mockProductRepo.On("FindLastByReceptionID", ctx, testReceptionID).Return(lastProduct, nil).Once()
				mockProductRepo.On("DeleteByID", ctx, testProductID).Return(someError).Once()
			},
			expectedError: domain.ErrDatabaseError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			err := productService.DeleteLastProduct(ctx, tc.pvzID)

			if tc.expectedError != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				require.NoError(t, err)
			}
			mockReceptionRepo.AssertExpectations(t)
			mockProductRepo.AssertExpectations(t)
		})
	}
}
