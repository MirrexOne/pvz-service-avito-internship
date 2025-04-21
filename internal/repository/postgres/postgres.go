package postgres

import (
	"context"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"pvz-service-avito-internship/internal/domain"
)

type BaseRepository struct {
	db  *pgxpool.Pool
	log *slog.Logger
	sq  sq.StatementBuilderType
}

func NewBaseRepository(db *pgxpool.Pool, log *slog.Logger) BaseRepository {
	return BaseRepository{
		db:  db,
		log: log,
		sq:  sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *BaseRepository) logQuery(ctx context.Context, op, query string, args ...interface{}) {
	reqID := ""
	if idVal := ctx.Value(contextKeyRequestID("requestID")); idVal != nil {
		if idStr, ok := idVal.(string); ok {
			reqID = idStr
		}
	}

	r.log.DebugContext(ctx, "Executing SQL query",
		slog.String("operation", op),
		slog.String("request_id", reqID),
		slog.String("query", query),
		slog.Any("args", args),
	)
}

func (r *BaseRepository) wrapErr(op string, err error) error {
	if err == nil {
		return nil
	}

	wrappedErr := fmt.Errorf("repository.%s: %w", op, err)

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return fmt.Errorf("%w: %w", domain.ErrNotFound, wrappedErr)
	default:
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				constraintName := pgErr.ConstraintName
				return fmt.Errorf("%w (constraint: %s): %w", domain.ErrConflict, constraintName, wrappedErr)
			}

			if pgErr.Code == "23503" {
				constraintName := pgErr.ConstraintName
				return fmt.Errorf("%w (foreign key constraint: %s): %w", domain.ErrValidation, constraintName, wrappedErr)
			}
		}
	}

	return fmt.Errorf("%w: %w", domain.ErrDatabaseError, wrappedErr)
}

func isErrNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

type contextKeyRequestID string
