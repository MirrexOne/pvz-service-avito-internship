package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"pvz-service-avito-internship/internal/domain"
	"pvz-service-avito-internship/internal/repository/postgres"
)

func createTestPVZ(ctx context.Context, t *testing.T, repo *postgres.PVZRepository, city domain.City) domain.PVZ {
	t.Helper()
	pvz := domain.PVZ{
		ID:               uuid.New(),
		RegistrationDate: time.Now().UTC().Truncate(time.Microsecond),
		City:             city,
	}
	err := repo.Create(ctx, &pvz)
	require.NoError(t, err, "Failed to create test PVZ")
	return pvz
}

func createTestReception(ctx context.Context, t *testing.T, repo *postgres.ReceptionRepository, pvzID uuid.UUID, dateTime time.Time) domain.Reception {
	t.Helper()

	_, err := dbPool.Exec(ctx, "COMMIT;")
	require.NoError(t, err, "Failed to commit transaction before creating reception")

	reception := domain.Reception{
		ID:       uuid.New(),
		DateTime: dateTime,
		PVZID:    pvzID,
		Status:   domain.StatusInProgress,
	}

	err = repo.Create(ctx, &reception)
	require.NoError(t, err, "Failed to create test reception")
	return reception
}

func createTestProduct(ctx context.Context, t *testing.T, repo *postgres.ProductRepository, recID uuid.UUID, prodType domain.ProductType, dateTime time.Time) domain.Product {
	t.Helper()
	product := domain.Product{
		ID:          uuid.New(),
		DateTime:    dateTime.UTC().Truncate(time.Microsecond),
		Type:        prodType,
		ReceptionID: recID,
	}
	err := repo.Create(ctx, &product)
	require.NoError(t, err, "Failed to create test product")
	return product
}
