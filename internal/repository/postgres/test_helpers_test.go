package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	// Используем пакеты из проекта
	"pvz-service-avito-internship/internal/domain"
	"pvz-service-avito-internship/internal/repository/postgres"
)

// --- Общие Тестовые Хелперы для Репозиториев ---

// createTestUser создает пользователя для тестов репозиториев.
// Принимает репозиторий пользователей для выполнения операции.
func createTestUser(ctx context.Context, t *testing.T, repo *postgres.UserRepository, email string, role domain.UserRole) domain.User {
	t.Helper()
	user := domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: "test_hash_for_" + email, // Уникальный хеш для теста
		Role:         role,
	}
	err := repo.Create(ctx, &user)
	require.NoError(t, err, "Failed to create test user")
	return user
}

// createTestPVZ создает ПВЗ для тестов репозиториев.
// Принимает репозиторий ПВЗ для выполнения операции.
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

// createTestReception создает приемку для тестов репозиториев.
// Принимает репозиторий приемок для выполнения операции.
func createTestReception(ctx context.Context, t *testing.T, repo *postgres.ReceptionRepository, pvzID uuid.UUID, dateTime time.Time) domain.Reception {
	t.Helper()
	reception := domain.Reception{
		ID:       uuid.New(),
		DateTime: dateTime.UTC().Truncate(time.Microsecond),
		PVZID:    pvzID,
		Status:   domain.StatusInProgress, // По умолчанию создаем открытую
	}
	err := repo.Create(ctx, &reception)
	require.NoError(t, err, "Failed to create test reception")
	return reception
}

// createTestProduct создает товар для тестов репозиториев.
// Принимает репозиторий товаров для выполнения операции.
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
