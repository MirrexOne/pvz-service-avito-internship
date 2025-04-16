package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt" // Для ошибки сравнения паролей

	// Используем относительные пути в рамках проекта
	"pvz-service-avito-internship/internal/domain"
	"pvz-service-avito-internship/internal/middleware" // Для RequestID
	"pvz-service-avito-internship/pkg/jwt"
)

// AuthService реализует интерфейс domain.AuthService.
type AuthService struct {
	log       *slog.Logger
	jwtSecret string                // Секрет для подписи JWT
	jwtTTL    time.Duration         // Время жизни JWT
	userRepo  domain.UserRepository // Зависимость от репозитория пользователей
	hasher    domain.PasswordHasher // Зависимость от хэшера паролей
}

// NewAuthService создает новый экземпляр AuthService.
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

// DummyLogin генерирует тестовый JWT токен.
func (s *AuthService) DummyLogin(ctx context.Context, role domain.UserRole) (string, error) {
	const op = "AuthService.DummyLogin"
	reqID := middleware.GetRequestIDFromContext(ctx)
	log := s.log.With(slog.String("op", op), slog.String("request_id", reqID), slog.String("role", string(role)))

	// Генерируем фиктивный User ID для dummy login
	userID := uuid.New()
	log = log.With(slog.String("user_id", userID.String()))

	// Валидация роли (на всякий случай, хотя хендлер тоже должен проверять)
	if !role.IsValid() {
		log.Warn("Invalid role provided for dummy login")
		return "", fmt.Errorf("%s: %w: invalid role '%s'", op, domain.ErrValidation, role)
	}

	token, err := jwt.GenerateToken(userID, role, s.jwtSecret, s.jwtTTL)
	if err != nil {
		log.Error("Failed to generate dummy token", slog.String("error", err.Error()))
		// Возвращаем общую ошибку сервера, т.к. это внутренняя проблема
		return "", fmt.Errorf("%s: %w", op, domain.ErrInternalServer)
	}

	log.Info("Dummy token generated successfully")
	return token, nil
}

// Register регистрирует нового пользователя.
func (s *AuthService) Register(ctx context.Context, email, password string, role domain.UserRole) (*domain.User, error) {
	const op = "AuthService.Register"
	reqID := middleware.GetRequestIDFromContext(ctx)
	log := s.log.With(slog.String("op", op), slog.String("request_id", reqID), slog.String("email", email), slog.String("role", string(role)))

	// Валидация входных данных (пример)
	if email == "" {
		return nil, fmt.Errorf("%s: %w: email cannot be empty", op, domain.ErrValidation)
	}
	if len(password) < 6 { // Минимальная длина пароля
		return nil, fmt.Errorf("%s: %w: password must be at least 6 characters long", op, domain.ErrValidation)
	}
	if !role.IsValid() {
		return nil, fmt.Errorf("%s: %w: invalid user role '%s'", op, domain.ErrValidation, role)
	}

	// Хешируем пароль
	hashedPassword, err := s.hasher.Hash(password)
	if err != nil {
		log.Error("Failed to hash password during registration", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, domain.ErrInternalServer)
	}

	// Создаем доменный объект User
	user := &domain.User{
		ID:           uuid.New(), // Генерируем новый ID
		Email:        email,
		PasswordHash: hashedPassword, // Сохраняем хеш
		Role:         role,
	}

	// Вызываем репозиторий для сохранения пользователя
	err = s.userRepo.Create(ctx, user)
	if err != nil {
		// Если репозиторий вернул ошибку конфликта (email занят)
		if errors.Is(err, domain.ErrConflict) {
			log.Warn("User registration failed due to email conflict")
			// Возвращаем доменную ошибку конфликта
			return nil, fmt.Errorf("%s: %w: email already exists", op, domain.ErrConflict)
		}
		// Другая ошибка репозитория
		log.Error("Failed to create user in repository", slog.String("error", err.Error()))
		// Оборачиваем как ошибку БД (репозиторий должен был это сделать, но дублируем для надежности)
		return nil, fmt.Errorf("%s: %w", op, domain.ErrDatabaseError)
	}

	log.Info("User registered successfully", slog.String("user_id", user.ID.String()))
	// Возвращаем пользователя без хеша пароля
	user.PasswordHash = ""
	return user, nil
}

// Login аутентифицирует пользователя.
func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	const op = "AuthService.Login"
	reqID := middleware.GetRequestIDFromContext(ctx)
	log := s.log.With(slog.String("op", op), slog.String("request_id", reqID), slog.String("email", email))

	// Валидация входа
	if email == "" || password == "" {
		return "", fmt.Errorf("%s: %w: email and password are required", op, domain.ErrValidation)
	}

	// Получаем пользователя из репозитория по email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Если пользователь не найден
		if errors.Is(err, domain.ErrNotFound) {
			log.Warn("Login attempt failed: user not found")
			// Возвращаем общую ошибку Unauthorized, чтобы не раскрывать существование email
			return "", fmt.Errorf("%s: %w: invalid email or password", op, domain.ErrUnauthorized)
		}
		// Другая ошибка репозитория
		log.Error("Failed to get user by email during login", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, domain.ErrDatabaseError)
	}

	// Сравниваем хеш пароля из БД с предоставленным паролем
	err = s.hasher.Compare(user.PasswordHash, password)
	if err != nil {
		// Если пароли не совпадают (основная ошибка bcrypt.ErrMismatchedHashAndPassword)
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			log.Warn("Login attempt failed: invalid password", slog.String("user_id", user.ID.String()))
			return "", fmt.Errorf("%s: %w: invalid email or password", op, domain.ErrUnauthorized)
		}
		// Другая ошибка при сравнении хеша (например, проблема с самим хешем)
		log.Error("Failed to compare password hash during login", slog.String("user_id", user.ID.String()), slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, domain.ErrInternalServer)
	}

	// Пароль верный, генерируем JWT токен
	token, err := jwt.GenerateToken(user.ID, user.Role, s.jwtSecret, s.jwtTTL)
	if err != nil {
		log.Error("Failed to generate token during login", slog.String("user_id", user.ID.String()), slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, domain.ErrInternalServer)
	}

	log.Info("User logged in successfully", slog.String("user_id", user.ID.String()))
	return token, nil
}
