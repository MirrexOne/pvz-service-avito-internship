package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pvz-service-avito-internship/internal/domain"
)

func TestReceptionRepository_Create_GetByID(t *testing.T) {
	require.NotNil(t, dbPool, "Test DB pool should be initialized")
	require.NotNil(t, testPVZRepo, "Test PVZ repo should be initialized")
	require.NotNil(t, testReceptionRepo, "Test Reception repo should be initialized")

	pvzRepo := testPVZRepo
	repo := testReceptionRepo
	ctx := context.Background()

	err := clearTables(ctx, dbPool, "pvz", "receptions", "products")
	require.NoError(t, err)

	pvz := createTestPVZ(ctx, t, pvzRepo, domain.Moscow)

	t.Run("Create and Get Success", func(t *testing.T) {
		now := time.Now()

		createdRec := createTestReception(ctx, t, repo, pvz.ID, now)

		fetchedRec, errGet := repo.GetByID(ctx, createdRec.ID)
		require.NoError(t, errGet)
		require.NotNil(t, fetchedRec)
		assert.Equal(t, createdRec.ID, fetchedRec.ID)
		assert.Equal(t, createdRec.PVZID, fetchedRec.PVZID)
		assert.Equal(t, domain.StatusInProgress, fetchedRec.Status)
		assert.WithinDuration(t, createdRec.DateTime, fetchedRec.DateTime, time.Second)
	})

	t.Run("Get Not Found", func(t *testing.T) {
		_, errGet := repo.GetByID(ctx, uuid.New())
		require.Error(t, errGet)
		assert.ErrorIs(t, errGet, domain.ErrNotFound)
	})

	t.Run("Create with Non-existent PVZ ID", func(t *testing.T) {
		reception := domain.Reception{
			ID:       uuid.New(),
			DateTime: time.Now().UTC(),
			PVZID:    uuid.New(),
			Status:   domain.StatusInProgress,
		}
		errCreate := repo.Create(ctx, &reception)
		require.Error(t, errCreate)
		assert.True(t, errors.Is(errCreate, domain.ErrValidation) || errors.Is(errCreate, domain.ErrDatabaseError))
	})
}

func TestReceptionRepository_FindOpenByPVZID(t *testing.T) {
	require.NotNil(t, dbPool, "Test DB pool should be initialized")
	require.NotNil(t, testPVZRepo, "Test PVZ repo should be initialized")
	require.NotNil(t, testReceptionRepo, "Test Reception repo should be initialized")
	pvzRepo := testPVZRepo
	repo := testReceptionRepo
	ctx := context.Background()

	err := clearTables(ctx, dbPool, "pvz", "receptions", "products")
	require.NoError(t, err)
	pvz1 := createTestPVZ(ctx, t, pvzRepo, domain.Moscow)
	pvz2 := createTestPVZ(ctx, t, pvzRepo, domain.Kazan)

	rec1Open := createTestReception(ctx, t, repo, pvz1.ID, time.Now().Add(-time.Hour))
	rec1Closed := createTestReception(ctx, t, repo, pvz1.ID, time.Now().Add(-2*time.Hour))
	errUpdate := repo.UpdateStatus(ctx, rec1Closed.ID, domain.StatusClosed)
	require.NoError(t, errUpdate)

	t.Run("Found Open", func(t *testing.T) {
		foundRec, err := repo.FindOpenByPVZID(ctx, pvz1.ID)
		require.NoError(t, err)
		require.NotNil(t, foundRec)
		assert.Equal(t, rec1Open.ID, foundRec.ID)
		assert.Equal(t, domain.StatusInProgress, foundRec.Status)
	})

	t.Run("Not Found Open", func(t *testing.T) {
		foundRec, err := repo.FindOpenByPVZID(ctx, pvz2.ID)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrNoOpenReception)
		assert.Nil(t, foundRec)
	})

	t.Run("PVZ Not Found", func(t *testing.T) {
		foundRec, err := repo.FindOpenByPVZID(ctx, uuid.New())
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrNoOpenReception)
		assert.Nil(t, foundRec)
	})
}

func TestReceptionRepository_UpdateStatus(t *testing.T) {
	require.NotNil(t, dbPool, "Test DB pool should be initialized")
	require.NotNil(t, testPVZRepo, "Test PVZ repo should be initialized")
	require.NotNil(t, testReceptionRepo, "Test Reception repo should be initialized")
	pvzRepo := testPVZRepo
	repo := testReceptionRepo
	ctx := context.Background()

	err := clearTables(ctx, dbPool, "pvz", "receptions", "products")
	require.NoError(t, err)
	pvz := createTestPVZ(ctx, t, pvzRepo, domain.Moscow)
	reception := createTestReception(ctx, t, repo, pvz.ID, time.Now())

	t.Run("Success Update to Closed", func(t *testing.T) {
		errUpdate := repo.UpdateStatus(ctx, reception.ID, domain.StatusClosed)
		require.NoError(t, errUpdate)

		fetchedRec, errGet := repo.GetByID(ctx, reception.ID)
		require.NoError(t, errGet)
		require.NotNil(t, fetchedRec)
		assert.Equal(t, domain.StatusClosed, fetchedRec.Status)
	})

	t.Run("Fail Update Not Found", func(t *testing.T) {
		errUpdate := repo.UpdateStatus(ctx, uuid.New(), domain.StatusClosed)
		require.Error(t, errUpdate)
		assert.ErrorIs(t, errUpdate, domain.ErrNotFound)
	})
}
