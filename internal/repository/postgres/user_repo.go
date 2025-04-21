package postgres

import (
	"context"
	"fmt"
	"log/slog"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"pvz-service-avito-internship/internal/domain"
)

// UserRepository реализует интерфейс domain.UserRepository для PostgreSQL.
type UserRepository struct {
	BaseRepository
}

// NewUserRepository создает новый экземпляр UserRepository.
func NewUserRepository(db *pgxpool.Pool, log *slog.Logger) *UserRepository {
	return &UserRepository{
		BaseRepository: NewBaseRepository(db, log),
	}
}

// Create сохраняет нового пользователя в базу данных.
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	const op = "UserRepository.Create"

	query, args, err := r.sq.Insert("users").
		Columns("id", "email", "password_hash", "role").
		Values(user.ID, user.Email, user.PasswordHash, user.Role).
		ToSql()
	if err != nil {
		return r.wrapErr(op, fmt.Errorf("failed to build query: %w", err))
	}

	r.logQuery(ctx, op, query, args...)
	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return r.wrapErr(op, err)
	}
	return nil
}

// GetByEmail находит пользователя по его email.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const op = "UserRepository.GetByEmail"

	query, args, err := r.sq.Select("id", "email", "password_hash", "role").
		From("users").
		Where(sq.Eq{"email": email}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, r.wrapErr(op, fmt.Errorf("failed to build query: %w", err))
	}

	r.logQuery(ctx, op, query, args...)
	row := r.db.QueryRow(ctx, query, args...)

	var user domain.User
	err = row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role)
	if err != nil {
		return nil, r.wrapErr(op, err)
	}
	return &user, nil
}

// GetByID находит пользователя по его уникальному идентификатору.
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	const op = "UserRepository.GetByID"

	query, args, err := r.sq.Select("id", "email", "password_hash", "role").
		From("users").
		Where(sq.Eq{"id": id}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, r.wrapErr(op, fmt.Errorf("failed to build query: %w", err))
	}

	r.logQuery(ctx, op, query, args...)
	row := r.db.QueryRow(ctx, query, args...)

	var user domain.User
	err = row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role)
	if err != nil {
		return nil, r.wrapErr(op, err)
	}
	return &user, nil
}
