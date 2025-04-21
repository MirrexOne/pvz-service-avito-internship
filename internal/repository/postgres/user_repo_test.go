package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pvz-service-avito-internship/internal/domain"
	"pvz-service-avito-internship/internal/repository/postgres"
)

func TestUserRepository_Create(t *testing.T) {
	require.NotNil(t, dbPool, "Test DB pool should be initialized in TestMain")
	userRepo := postgres.NewUserRepository(dbPool, testLogger)
	ctx := context.Background()

	err := clearTables(ctx, dbPool, "users")
	require.NoError(t, err)

	newUser := &domain.User{
		ID:           uuid.New(),
		Email:        "create.test@example.com",
		PasswordHash: "some_hash",
		Role:         domain.RoleEmployee,
	}

	t.Run("Success", func(t *testing.T) {
		err := userRepo.Create(ctx, newUser)
		require.NoError(t, err)

		createdUser, errGet := userRepo.GetByEmail(ctx, newUser.Email)
		require.NoError(t, errGet)
		require.NotNil(t, createdUser)
		assert.Equal(t, newUser.ID, createdUser.ID)
		assert.Equal(t, newUser.Email, createdUser.Email)
		assert.Equal(t, newUser.PasswordHash, createdUser.PasswordHash)
		assert.Equal(t, newUser.Role, createdUser.Role)
	})

	t.Run("Fail_Duplicate_Email", func(t *testing.T) {
		duplicateUser := &domain.User{
			ID:           uuid.New(),
			Email:        "create.test@example.com",
			PasswordHash: "another_hash",
			Role:         domain.RoleModerator,
		}
		err := userRepo.Create(ctx, duplicateUser)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrConflict, "Expected conflict error for duplicate email")
	})
}

func TestUserRepository_GetByEmail(t *testing.T) {
	require.NotNil(t, dbPool, "Test DB pool should be initialized in TestMain")
	userRepo := postgres.NewUserRepository(dbPool, testLogger)
	ctx := context.Background()

	err := clearTables(ctx, dbPool, "users")
	require.NoError(t, err)
	existingUser := &domain.User{
		ID:           uuid.New(),
		Email:        "get.test@example.com",
		PasswordHash: "get_hash",
		Role:         domain.RoleModerator,
	}
	errCreate := userRepo.Create(ctx, existingUser)
	require.NoError(t, errCreate)

	t.Run("Success_Found", func(t *testing.T) {
		foundUser, err := userRepo.GetByEmail(ctx, existingUser.Email)
		require.NoError(t, err)
		require.NotNil(t, foundUser)
		assert.Equal(t, existingUser.ID, foundUser.ID)
		assert.Equal(t, existingUser.Email, foundUser.Email)
		assert.Equal(t, existingUser.PasswordHash, foundUser.PasswordHash)
		assert.Equal(t, existingUser.Role, foundUser.Role)
	})

	t.Run("Fail_NotFound", func(t *testing.T) {
		foundUser, err := userRepo.GetByEmail(ctx, "not.found@example.com")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, foundUser)
	})
}

func TestUserRepository_GetByID(t *testing.T) {
	require.NotNil(t, dbPool, "Test DB pool should be initialized in TestMain")
	userRepo := postgres.NewUserRepository(dbPool, testLogger)
	ctx := context.Background()

	err := clearTables(ctx, dbPool, "users")
	require.NoError(t, err)
	existingUser := &domain.User{
		ID:           uuid.New(),
		Email:        "getbyid.test@example.com",
		PasswordHash: "getbyid_hash",
		Role:         domain.RoleEmployee,
	}
	errCreate := userRepo.Create(ctx, existingUser)
	require.NoError(t, errCreate)

	t.Run("Success_Found", func(t *testing.T) {
		foundUser, err := userRepo.GetByID(ctx, existingUser.ID)
		require.NoError(t, err)
		require.NotNil(t, foundUser)
		assert.Equal(t, existingUser.ID, foundUser.ID)
		assert.Equal(t, existingUser.Email, foundUser.Email)
	})

	t.Run("Fail_NotFound", func(t *testing.T) {
		notFoundID := uuid.New()
		foundUser, err := userRepo.GetByID(ctx, notFoundID)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, foundUser)
	})
}
