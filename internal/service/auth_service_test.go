package service_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"io"
	"log/slog"
	"testing"
	"time"

	_ "github.com/golang-migrate/migrate/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"pvz-service-avito-internship/internal/domain"
	"pvz-service-avito-internship/internal/service"
	"pvz-service-avito-internship/mocks"
)

func TestAuthService_DummyLogin(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockUserRepo := mocks.NewUserRepository(t)
	mockHasher := mocks.NewPasswordHasher(t)
	jwtSecret := "test-secret"
	jwtTTL := time.Hour

	authService := service.NewAuthService(logger, jwtSecret, jwtTTL, mockUserRepo, mockHasher)

	ctx := context.Background()

	testCases := []struct {
		name        string
		role        domain.UserRole
		expectedErr error
	}{
		{
			name:        "Success_Employee",
			role:        domain.RoleEmployee,
			expectedErr: nil,
		},
		{
			name:        "Success_Moderator",
			role:        domain.RoleModerator,
			expectedErr: nil,
		},
		{
			name:        "Fail_Invalid_Role",
			role:        domain.UserRole("invalid"),
			expectedErr: domain.ErrValidation,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := authService.DummyLogin(ctx, tc.role)

			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.expectedErr) || errors.Is(err, err), "Error should match expected type")
				assert.Empty(t, token)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, token)
				parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
					}
					return []byte(jwtSecret), nil
				})
				require.NoError(t, err)
				require.NotNil(t, parsedToken)

				claims, err := validateJWTTokenWithoutSignature(token)
				require.NoError(t, err, "Failed to validate JWT token")
				require.NotNil(t, claims, "JWT claims should not be nil")
			}
		})
	}
}

func TestAuthService_Register(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockUserRepo := mocks.NewUserRepository(t)
	mockHasher := mocks.NewPasswordHasher(t)
	jwtSecret := "test-secret"
	jwtTTL := time.Hour

	authService := service.NewAuthService(logger, jwtSecret, jwtTTL, mockUserRepo, mockHasher)
	ctx := context.Background()

	testCases := []struct {
		name          string
		email         string
		password      string
		role          domain.UserRole
		setupMocks    func(email, hashedPassword string, role domain.UserRole)
		expectedUser  *domain.User
		expectedError error
	}{
		{
			name:     "Success",
			email:    "test@example.com",
			password: "password123",
			role:     domain.RoleEmployee,
			setupMocks: func(email, hashedPassword string, role domain.UserRole) {
				mockHasher.On("Hash", "password123").Return("hashed_password", nil).Once()
				mockUserRepo.On("Create", mock.Anything, mock.MatchedBy(func(user *domain.User) bool {
					return user.Email == email && user.PasswordHash == "hashed_password" && user.Role == role
				})).Return(nil).Once()
			},
			expectedUser:  &domain.User{Email: "test@example.com", Role: domain.RoleEmployee},
			expectedError: nil,
		},
		{
			name:     "Fail_Duplicate_Email",
			email:    "duplicate@example.com",
			password: "password123",
			role:     domain.RoleModerator,
			setupMocks: func(email, hashedPassword string, role domain.UserRole) {
				mockHasher.On("Hash", "password123").Return("hashed_password", nil).Once()
				mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).
					Return(domain.ErrConflict).Once()
			},
			expectedUser:  nil,
			expectedError: domain.ErrConflict,
		},
		{
			name:          "Fail_Short_Password",
			email:         "short@example.com",
			password:      "123",
			role:          domain.RoleEmployee,
			setupMocks:    func(email, hashedPassword string, role domain.UserRole) {},
			expectedUser:  nil,
			expectedError: domain.ErrValidation,
		},
		{
			name:     "Fail_Hashing_Error",
			email:    "hashfail@example.com",
			password: "password123",
			role:     domain.RoleEmployee,
			setupMocks: func(email, hashedPassword string, role domain.UserRole) {
				mockHasher.On("Hash", "password123").Return("", errors.New("bcrypt error")).Once()
			},
			expectedUser:  nil,
			expectedError: domain.ErrInternalServer,
		},
		{
			name:     "Fail_Repo_Create_Error",
			email:    "repofail@example.com",
			password: "password123",
			role:     domain.RoleEmployee,
			setupMocks: func(email, hashedPassword string, role domain.UserRole) {
				mockHasher.On("Hash", "password123").Return("hashed_password", nil).Once()
				mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).
					Return(errors.New("some db error")).Once()
			},
			expectedUser:  nil,
			expectedError: domain.ErrDatabaseError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks(tc.email, "hashed_password", tc.role)

			user, err := authService.Register(ctx, tc.email, tc.password, tc.role)

			if tc.expectedError != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedError, "Error should match expected type")
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, tc.expectedUser.Email, user.Email)
				assert.Equal(t, tc.expectedUser.Role, user.Role)
				assert.NotEqual(t, uuid.Nil, user.ID)
				assert.Empty(t, user.PasswordHash)
			}
			mockUserRepo.AssertExpectations(t)
			mockHasher.AssertExpectations(t)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockUserRepo := mocks.NewUserRepository(t)
	mockHasher := mocks.NewPasswordHasher(t)
	jwtSecret := "test-secret"
	jwtTTL := time.Hour

	authService := service.NewAuthService(logger, jwtSecret, jwtTTL, mockUserRepo, mockHasher)
	ctx := context.Background()

	validEmail := "user@example.com"
	validPassword := "password123"
	validHash := "$2a$10$abcdefghijklmnopqrstuv"
	validUser := &domain.User{
		ID:           uuid.New(),
		Email:        validEmail,
		PasswordHash: validHash,
		Role:         domain.RoleEmployee,
	}

	testCases := []struct {
		name          string
		email         string
		password      string
		setupMocks    func()
		expectedToken bool
		expectedError error
	}{
		{
			name:     "Success",
			email:    validEmail,
			password: validPassword,
			setupMocks: func() {
				mockUserRepo.On("GetByEmail", mock.Anything, validEmail).Return(validUser, nil).Once()
				mockHasher.On("Compare", validHash, validPassword).Return(nil).Once()
			},
			expectedToken: true,
			expectedError: nil,
		},
		{
			name:     "Fail_User_NotFound",
			email:    "notfound@example.com",
			password: validPassword,
			setupMocks: func() {
				mockUserRepo.On("GetByEmail", mock.Anything, "notfound@example.com").Return(nil, domain.ErrNotFound).Once()
			},
			expectedToken: false,
			expectedError: domain.ErrUnauthorized,
		},
		{
			name:     "Fail_Incorrect_Password",
			email:    validEmail,
			password: "wrongpassword",
			setupMocks: func() {
				mockUserRepo.On("GetByEmail", mock.Anything, validEmail).Return(validUser, nil).Once()
				mockHasher.On("Compare", validHash, "wrongpassword").Return(bcrypt.ErrMismatchedHashAndPassword).Once()
			},
			expectedToken: false,
			expectedError: domain.ErrUnauthorized,
		},
		{
			name:     "Fail_Repo_GetByEmail_Error",
			email:    validEmail,
			password: validPassword,
			setupMocks: func() {
				mockUserRepo.On("GetByEmail", mock.Anything, validEmail).Return(nil, errors.New("db error")).Once()
			},
			expectedToken: false,
			expectedError: domain.ErrDatabaseError,
		},
		{
			name:     "Fail_Hasher_Compare_Error",
			email:    validEmail,
			password: validPassword,
			setupMocks: func() {
				mockUserRepo.On("GetByEmail", mock.Anything, validEmail).Return(validUser, nil).Once()
				mockHasher.On("Compare", validHash, validPassword).Return(errors.New("compare error")).Once()
			},
			expectedToken: false,
			expectedError: domain.ErrInternalServer,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			token, err := authService.Login(ctx, tc.email, tc.password)

			if tc.expectedError != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Empty(t, token)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, token)
			}
			mockUserRepo.AssertExpectations(t)
			mockHasher.AssertExpectations(t)
		})
	}
}

func validateJWTTokenWithoutSignature(tokenString string) (*jwt.Token, error) {

	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	if exp, ok := claims["exp"].(float64); ok {
		if time.Unix(int64(exp), 0).Before(time.Now()) {
			return nil, errors.New("token is expired")
		}
	}

	return token, nil
}
