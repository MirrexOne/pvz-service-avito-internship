package postgres_test

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pvz-service-avito-internship/internal/domain"
)

func TestPVZRepository_Create_GetByID(t *testing.T) {
	require.NotNil(t, dbPool, "Test DB pool should be initialized")
	repo := testPVZRepo
	ctx := context.Background()

	err := clearTables(ctx, dbPool, "pvz", "receptions", "products")
	require.NoError(t, err)

	t.Run("Create and Get Success", func(t *testing.T) {
		city := domain.Moscow
		createdPVZ := createTestPVZ(ctx, t, repo, city)

		fetchedPVZ, errGet := repo.GetByID(ctx, createdPVZ.ID)
		require.NoError(t, errGet)
		require.NotNil(t, fetchedPVZ)
		assert.Equal(t, createdPVZ.ID, fetchedPVZ.ID)
		assert.Equal(t, createdPVZ.City, fetchedPVZ.City)
		assert.WithinDuration(t, createdPVZ.RegistrationDate, fetchedPVZ.RegistrationDate, time.Second)
	})

	t.Run("Get Not Found", func(t *testing.T) {
		_, errGet := repo.GetByID(ctx, uuid.New())
		require.Error(t, errGet)
		assert.ErrorIs(t, errGet, domain.ErrNotFound)
	})
}

func TestPVZRepository_GetByIDs(t *testing.T) {
	require.NotNil(t, dbPool, "Test DB pool should be initialized")
	repo := testPVZRepo
	ctx := context.Background()

	err := clearTables(ctx, dbPool, "pvz", "receptions", "products")
	require.NoError(t, err)

	pvz1 := createTestPVZ(ctx, t, repo, domain.Moscow)
	pvz2 := createTestPVZ(ctx, t, repo, domain.Kazan)
	_ = createTestPVZ(ctx, t, repo, domain.SaintPetersburg)

	t.Run("Found Multiple", func(t *testing.T) {
		idsToFind := []uuid.UUID{pvz1.ID, pvz2.ID}
		foundPVZs, err := repo.GetByIDs(ctx, idsToFind)
		require.NoError(t, err)
		require.Len(t, foundPVZs, 2)
		foundMap := make(map[uuid.UUID]domain.PVZ)
		for _, p := range foundPVZs {
			foundMap[p.ID] = p
		}
		_, ok1 := foundMap[pvz1.ID]
		_, ok2 := foundMap[pvz2.ID]
		assert.True(t, ok1)
		assert.True(t, ok2)
	})
	t.Run("Found One", func(t *testing.T) {
		idsToFind := []uuid.UUID{pvz1.ID, uuid.New()}
		foundPVZs, err := repo.GetByIDs(ctx, idsToFind)
		require.NoError(t, err)
		require.Len(t, foundPVZs, 1)
		assert.Equal(t, pvz1.ID, foundPVZs[0].ID)
	})
	t.Run("Found None", func(t *testing.T) {
		idsToFind := []uuid.UUID{uuid.New(), uuid.New()}
		foundPVZs, err := repo.GetByIDs(ctx, idsToFind)
		require.NoError(t, err)
		assert.Empty(t, foundPVZs)
	})
	t.Run("Empty Input", func(t *testing.T) {
		var idsToFind []uuid.UUID
		foundPVZs, err := repo.GetByIDs(ctx, idsToFind)
		require.NoError(t, err)
		assert.Empty(t, foundPVZs)
	})
}

func runTestInTransaction(t *testing.T, dbPool *pgxpool.Pool, testFunc func(ctx context.Context)) {
	ctx := context.Background()
	tx, err := dbPool.Begin(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	testFunc(ctx)
}

func TestPVZRepository_ListIDsAndTotal(t *testing.T) {
	require.NotNil(t, dbPool, "Test DB pool should be initialized")
	repo := testPVZRepo
	receptionRepo := testReceptionRepo
	ctx := context.Background()

	err := clearTables(ctx, dbPool, "pvz", "receptions", "products")
	require.NoError(t, err)

	pvzM := createTestPVZ(ctx, t, repo, domain.Moscow)
	require.NotNil(t, pvzM)
	pvzK := createTestPVZ(ctx, t, repo, domain.Kazan)
	require.NotNil(t, pvzK)
	pvzS := createTestPVZ(ctx, t, repo, domain.SaintPetersburg)
	require.NotNil(t, pvzS)

	checkPVZExists := func(pvzID uuid.UUID) {
		var exists bool
		err := dbPool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pvz WHERE id = $1)", pvzID).Scan(&exists)
		require.NoError(t, err, "Failed to check if PVZ exists")
		require.True(t, exists, "PVZ should exist in the database")
	}

	checkPVZExists(pvzM.ID)
	checkPVZExists(pvzK.ID)
	checkPVZExists(pvzS.ID)

	now := time.Now().UTC()
	createTestReception(ctx, t, receptionRepo, pvzM.ID, now.Add(-2*time.Hour))
	createTestReception(ctx, t, receptionRepo, pvzK.ID, now.Add(-1*time.Hour))
	recMS := createTestReception(ctx, t, receptionRepo, pvzM.ID, now.Add(-30*time.Minute))
	require.NotNil(t, recMS)

	errUpdate := receptionRepo.UpdateStatus(ctx, recMS.ID, domain.StatusClosed)
	require.NoError(t, errUpdate)

	createTestReception(ctx, t, receptionRepo, pvzS.ID, now.Add(time.Hour))

	t.Run("No Filters, First Page", func(t *testing.T) {
		ids, total, err := repo.ListIDsAndTotal(ctx, 2, 0, nil, nil)
		require.NoError(t, err)
		assert.Equal(t, 3, total)
		require.Len(t, ids, 2)
		assert.Contains(t, ids, pvzS.ID)
		assert.Contains(t, ids, pvzK.ID)
	})

	t.Run("No Filters, Second Page", func(t *testing.T) {
		ids, total, err := repo.ListIDsAndTotal(ctx, 2, 2, nil, nil)
		require.NoError(t, err)
		assert.Equal(t, 3, total)
		require.Len(t, ids, 1)
		if len(ids) > 0 {
			assert.Equal(t, pvzM.ID, ids[0])
		}
	})

	t.Run("Date Filter - Past Hour", func(t *testing.T) {
		startTime := now.Add(-65 * time.Minute)
		endTime := now.Add(-25 * time.Minute)
		ids, total, err := repo.ListIDsAndTotal(ctx, 10, 0, &startTime, &endTime)
		require.NoError(t, err)
		assert.Equal(t, 2, total)
		require.Len(t, ids, 2)
		assert.Contains(t, ids, pvzK.ID)
		assert.Contains(t, ids, pvzM.ID)
	})
	t.Run("Date Filter - No Match", func(t *testing.T) {
		startTime := now.Add(-10 * time.Minute)
		endTime := now.Add(10 * time.Minute)
		ids, total, err := repo.ListIDsAndTotal(ctx, 10, 0, &startTime, &endTime)
		require.NoError(t, err)
		assert.Equal(t, 0, total)
		assert.Empty(t, ids)
	})
	t.Run("Empty Result", func(t *testing.T) {
		err := clearTables(ctx, dbPool, "pvz", "receptions", "products")
		require.NoError(t, err)
		ids, total, err := repo.ListIDsAndTotal(ctx, 10, 0, nil, nil)
		require.NoError(t, err)
		assert.Equal(t, 0, total)
		assert.Empty(t, ids)
	})
}

func TestPVZRepository_ListAll(t *testing.T) {
	require.NotNil(t, dbPool, "Test DB pool should be initialized")
	repo := testPVZRepo
	ctx := context.Background()

	err := clearTables(ctx, dbPool, "pvz")
	require.NoError(t, err)

	pvz1 := createTestPVZ(ctx, t, repo, domain.Moscow)
	pvz2 := createTestPVZ(ctx, t, repo, domain.Kazan)

	t.Run("Success Multiple", func(t *testing.T) {
		pvzs, err := repo.ListAll(ctx)
		require.NoError(t, err)
		require.Len(t, pvzs, 2)
		ids := []uuid.UUID{pvzs[0].ID, pvzs[1].ID}
		assert.Contains(t, ids, pvz1.ID)
		assert.Contains(t, ids, pvz2.ID)
	})

	t.Run("Success Empty", func(t *testing.T) {
		err := clearTables(ctx, dbPool, "pvz")
		require.NoError(t, err)
		pvzs, err := repo.ListAll(ctx)
		require.NoError(t, err)
		assert.Empty(t, pvzs)
	})
}
