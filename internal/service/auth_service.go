package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"pvz-service-avito-internship/internal/domain"
	"pvz-service-avito-internship/internal/middleware"
	"pvz-service-avito-internship/pkg/jwt"
)

type AuthService struct {
	log       *slog.Logger
	jwtSecret string
	jwtTTL    time.Duration
	userRepo  domain.UserRepository
	hasher    domain.PasswordHasher
}

func NewAuthService(
	log *slog.Logger,
	jwtSecret string,
	jwtTTL time.Duration,
	userRepo domain.UserRepository,
	hasher domain.PasswordHasher,
) *AuthService {
	return &AuthService{
		log:       log,
		jwtSecret: jwtSecret,
		jwtTTL:    jwtTTL,
		userRepo:  userRepo,
		hasher:    hasher,
	}
}

func (s *AuthService) DummyLogin(ctx context.Context, role domain.UserRole) (string, error) {
	const op = "AuthService.DummyLogin"
	reqID := middleware.GetRequestIDFromContext(ctx)
	log := s.log.With(slog.String("op", op), slog.String("request_id", reqID), slog.String("role", string(role)))

	userID := uuid.New()
	log = log.With(slog.String("user_id", userID.String()))

	if !role.IsValid() {
		log.Warn("Invalid role provided for dummy login")
		return "", fmt.Errorf("%s: %w: invalid role '%s'", op, domain.ErrValidation, role)
	}

	token, err := jwt.GenerateToken(userID, role, s.jwtSecret, s.jwtTTL)
	if err != nil {
		log.Error("Failed to generate dummy token", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, domain.ErrInternalServer)
	}

	log.Info("Dummy token generated successfully")
	return token, nil
}

func (s *AuthService) Register(ctx context.Context, email, password string, role domain.UserRole) (*domain.User, error) {
	const op = "AuthService.Register"

	reqID := middleware.GetRequestIDFromContext(ctx)
	log := s.log.With(slog.String("op", op), slog.String("request_id", reqID), slog.String("email", email), slog.String("role", string(role)))

	if !role.IsValid() {
		return nil, fmt.Errorf("%s: %w: invalid user role '%s'", op, domain.ErrValidation, role)
	}

	hashedPassword, err := s.hasher.Hash(password)
	if err != nil {
		log.Error("Failed to hash password during registration", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, domain.ErrInternalServer)
	}

	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         role,
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			log.Warn("User registration failed due to email conflict")
			return nil, fmt.Errorf("%s: %w: email already exists", op, domain.ErrConflict)
		}
		log.Error("Failed to create user in repository", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, domain.ErrDatabaseError)
	}

	log.Info("User registered successfully", slog.String("user_id", user.ID.String()))
	user.PasswordHash = ""
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	const op = "AuthService.Login"
	reqID := middleware.GetRequestIDFromContext(ctx)
	log := s.log.With(slog.String("op", op), slog.String("request_id", reqID), slog.String("email", email))

	if email == "" || password == "" {
		return "", fmt.Errorf("%s: %w: email and password are required", op, domain.ErrValidation)
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			log.Warn("Login attempt failed: user not found")
			return "", fmt.Errorf("%s: %w: invalid email or password", op, domain.ErrUnauthorized)
		}
		log.Error("Failed to get user by email during login", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, domain.ErrDatabaseError)
	}

	err = s.hasher.Compare(user.PasswordHash, password)
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			log.Warn("Login attempt failed: invalid password", slog.String("user_id", user.ID.String()))
			return "", fmt.Errorf("%s: %w: invalid email or password", op, domain.ErrUnauthorized)
		}

		log.Error("Failed to compare password hash during login", slog.String("user_id", user.ID.String()), slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, domain.ErrInternalServer)
	}

	token, err := jwt.GenerateToken(user.ID, user.Role, s.jwtSecret, s.jwtTTL)
	if err != nil {
		log.Error("Failed to generate token during login", slog.String("user_id", user.ID.String()), slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, domain.ErrInternalServer)
	}

	log.Info("User logged in successfully", slog.String("user_id", user.ID.String()))
	return token, nil
}
