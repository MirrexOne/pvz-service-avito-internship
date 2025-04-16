package postgres

import (
	"context"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel" // Псевдоним для squirrel
	"github.com/jackc/pgx/v5"            // Используется для pgx.ErrNoRows
	"github.com/jackc/pgx/v5/pgconn"     // Используется для pgconn.PgError
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	// Импортируем доменные ошибки
	"pvz-service-avito-internship/internal/domain"
)

// BaseRepository содержит общие компоненты и методы, используемые всеми репозиториями PostgreSQL.
// Встраивается в конкретные структуры репозиториев.
type BaseRepository struct {
	db  *pgxpool.Pool           // Пул соединений с базой данных
	log *slog.Logger            // Логгер для записи информации (например, SQL запросов)
	sq  sq.StatementBuilderType // Query builder (squirrel) с плейсхолдерами PostgreSQL ($1, $2, ...)
}

// NewBaseRepository создает и возвращает новый экземпляр BaseRepository.
func NewBaseRepository(db *pgxpool.Pool, log *slog.Logger) BaseRepository {
	return BaseRepository{
		db:  db,
		log: log,
		// Инициализируем squirrel с плейсхолдерами для PostgreSQL
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

// logQuery - вспомогательный метод для логирования выполняемых SQL запросов и их аргументов.
// Помогает при отладке. Вызывается перед выполнением запроса к БД.
func (r *BaseRepository) logQuery(ctx context.Context, op, query string, args ...interface{}) {
	// Извлекаем request_id из контекста, если он там есть
	reqID := "" // По умолчанию пустой
	if idVal := ctx.Value(contextKeyRequestID("requestID")); idVal != nil {
		if idStr, ok := idVal.(string); ok {
			reqID = idStr
		}
	}

	// Логируем с уровнем Debug, чтобы не засорять основные логи
	r.log.DebugContext(ctx, "Executing SQL query",
		slog.String("operation", op),
		slog.String("request_id", reqID), // Добавляем request_id в лог
		slog.String("query", query),
		slog.Any("args", args), // Используем Any для универсальности аргументов
	)
}

// wrapErr - вспомогательный метод для оборачивания ошибок, возникающих в репозитории.
// Добавляет контекст операции (op) к ошибке и преобразует специфичные ошибки БД в доменные ошибки.
func (r *BaseRepository) wrapErr(op string, err error) error {
	if err == nil {
		return nil
	}

	// Оборачиваем исходную ошибку для сохранения стектрейса
	wrappedErr := fmt.Errorf("repository.%s: %w", op, err)

	// Преобразуем специфичные ошибки pgx/pgconn в доменные ошибки
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		// Если запрос не вернул строк, возвращаем доменную ошибку ErrNotFound
		return fmt.Errorf("%w: %w", domain.ErrNotFound, wrappedErr)
	default:
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// Проверяем на ошибку уникальности (код 23505)
			if pgErr.Code == "23505" {
				// Если это ошибка уникальности, возвращаем доменную ErrConflict
				// Добавляем имя ограничения для контекста
				constraintName := pgErr.ConstraintName
				return fmt.Errorf("%w (constraint: %s): %w", domain.ErrConflict, constraintName, wrappedErr)
			}
			// Можно добавить обработку других кодов ошибок PostgreSQL при необходимости
			// (например, нарушение внешнего ключа - код 23503)
			if pgErr.Code == "23503" {
				constraintName := pgErr.ConstraintName
				// Для FK можно вернуть ErrValidation или ErrNotFound в зависимости от контекста
				return fmt.Errorf("%w (foreign key constraint: %s): %w", domain.ErrValidation, constraintName, wrappedErr)
			}
		}
	}

	// Если ошибка не была распознана как специфическая, возвращаем общую ошибку БД
	// Исходная ошибка все еще доступна через errors.Unwrap()
	return fmt.Errorf("%w: %w", domain.ErrDatabaseError, wrappedErr)
}

// isErrNoRows - **ФУНКЦИЯ ПАКЕТА**, проверяет ошибку pgx.ErrNoRows.
// Используется внутри этого пакета, чтобы избежать экспорта зависимости от pgx.
func isErrNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

// contextKeyRequestID используется для ключа в контексте
type contextKeyRequestID string
