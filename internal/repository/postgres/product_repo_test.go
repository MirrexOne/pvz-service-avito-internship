package postgres_test // Пакет _test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pvz-service-avito-internship/internal/domain"
	// Не импортируем postgres напрямую
)

func TestProductRepository_Create(t *testing.T) {
	require.NotNil(t, dbPool, "Test DB pool should be initialized")
	require.NotNil(t, testPVZRepo, "Test PVZ repo should be initialized")
	require.NotNil(t, testReceptionRepo, "Test Reception repo should be initialized")
	require.NotNil(t, testProductRepo, "Test Product repo should be initialized")
	// Используем глобальные репозитории
	pvzRepo := testPVZRepo
	receptionRepo := testReceptionRepo
	repo := testProductRepo
	ctx := context.Background()

	err := clearTables(ctx, dbPool, "pvz", "receptions", "products")
	require.NoError(t, err)
	// Используем глобальные хелперы
	pvz := createTestPVZ(ctx, t, pvzRepo, domain.Moscow)
	reception := createTestReception(ctx, t, receptionRepo, pvz.ID, time.Now())

	t.Run("Success", func(t *testing.T) {
		now := time.Now()
		prodType := domain.TypeElectronics
		// Используем глобальный хелпер
		createdProd := createTestProduct(ctx, t, repo, reception.ID, prodType, now)

		// Проверяем, что товар создан
		var fetchedType string
		errScan := dbPool.QueryRow(ctx, "SELECT type FROM products WHERE id = $1", createdProd.ID).Scan(&fetchedType)
		require.NoError(t, errScan)
		assert.Equal(t, string(prodType), fetchedType)
	})

	t.Run("Fail Invalid Reception ID", func(t *testing.T) {
		product := domain.Product{
			ID:          uuid.New(),
			DateTime:    time.Now().UTC(),
			Type:        domain.TypeClothing,
			ReceptionID: uuid.New(), // Несуществующий ID
		}
		errCreate := repo.Create(ctx, &product)
		require.Error(t, errCreate)
		// Ожидаем ошибку нарушения FK
		assert.True(t, errors.Is(errCreate, domain.ErrValidation) || errors.Is(errCreate, domain.ErrDatabaseError))
	})
}

func TestProductRepository_FindLastByReceptionID(t *testing.T) {
	require.NotNil(t, dbPool, "Test DB pool should be initialized")
	require.NotNil(t, testPVZRepo, "Test PVZ repo should be initialized")
	require.NotNil(t, testReceptionRepo, "Test Reception repo should be initialized")
	require.NotNil(t, testProductRepo, "Test Product repo should be initialized")
	pvzRepo := testPVZRepo
	receptionRepo := testReceptionRepo
	repo := testProductRepo
	ctx := context.Background()

	err := clearTables(ctx, dbPool, "pvz", "receptions", "products")
	require.NoError(t, err)
	pvz := createTestPVZ(ctx, t, pvzRepo, domain.Moscow)
	reception := createTestReception(ctx, t, receptionRepo, pvz.ID, time.Now())
	recEmpty := createTestReception(ctx, t, receptionRepo, pvz.ID, time.Now().Add(time.Minute))

	now := time.Now()
	// Используем глобальный хелпер
	// Создаем товары с разными временами
	createTestProduct(ctx, t, repo, reception.ID, domain.TypeElectronics, now.Add(-time.Second*2)) // Используем _
	createTestProduct(ctx, t, repo, reception.ID, domain.TypeClothing, now.Add(-time.Second*1))    // Используем _
	prod3 := createTestProduct(ctx, t, repo, reception.ID, domain.TypeShoes, now)                  // Этот используется

	t.Run("Found Last", func(t *testing.T) {
		lastProd, err := repo.FindLastByReceptionID(ctx, reception.ID)
		require.NoError(t, err)
		require.NotNil(t, lastProd)
		assert.Equal(t, prod3.ID, lastProd.ID)
		assert.Equal(t, prod3.Type, lastProd.Type)
		assert.WithinDuration(t, prod3.DateTime, lastProd.DateTime, time.Second)
	})

	t.Run("Not Found (Empty Reception)", func(t *testing.T) {
		lastProd, err := repo.FindLastByReceptionID(ctx, recEmpty.ID)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrNoProductsToDelete)
		assert.Nil(t, lastProd)
	})

	t.Run("Not Found (Invalid Reception ID)", func(t *testing.T) {
		lastProd, err := repo.FindLastByReceptionID(ctx, uuid.New())
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrNoProductsToDelete)
		assert.Nil(t, lastProd)
	})
}

func TestProductRepository_DeleteByID(t *testing.T) {
	require.NotNil(t, dbPool, "Test DB pool should be initialized")
	require.NotNil(t, testPVZRepo, "Test PVZ repo should be initialized")
	require.NotNil(t, testReceptionRepo, "Test Reception repo should be initialized")
	require.NotNil(t, testProductRepo, "Test Product repo should be initialized")
	pvzRepo := testPVZRepo
	receptionRepo := testReceptionRepo
	repo := testProductRepo
	ctx := context.Background()

	err := clearTables(ctx, dbPool, "pvz", "receptions", "products")
	require.NoError(t, err)
	pvz := createTestPVZ(ctx, t, pvzRepo, domain.Moscow)
	reception := createTestReception(ctx, t, receptionRepo, pvz.ID, time.Now())
	// Используем глобальный хелпер
	product := createTestProduct(ctx, t, repo, reception.ID, domain.TypeElectronics, time.Now())

	t.Run("Success Delete", func(t *testing.T) {
		errDelete := repo.DeleteByID(ctx, product.ID)
		require.NoError(t, errDelete)

		// Проверяем, что товар удален
		var count int
		errScan := dbPool.QueryRow(ctx, "SELECT COUNT(*) FROM products WHERE id = $1", product.ID).Scan(&count)
		require.NoError(t, errScan)
		assert.Equal(t, 0, count)
	})

	t.Run("Fail Delete Not Found", func(t *testing.T) {
		errDelete := repo.DeleteByID(ctx, uuid.New())
		require.Error(t, errDelete)
		assert.ErrorIs(t, errDelete, domain.ErrNotFound)
	})
}
