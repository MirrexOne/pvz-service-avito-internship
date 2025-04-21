package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"pvz-service-avito-internship/internal/domain"
)

// PVZRepository реализует интерфейс domain.PVZRepository для PostgreSQL.
type PVZRepository struct {
	BaseRepository
}

// NewPVZRepository создает новый экземпляр PVZRepository.
func NewPVZRepository(db *pgxpool.Pool, log *slog.Logger) *PVZRepository {
	return &PVZRepository{
		BaseRepository: NewBaseRepository(db, log),
	}
}

// Create сохраняет новый ПВЗ в базу данных.
func (r *PVZRepository) Create(ctx context.Context, pvz *domain.PVZ) error {
	const op = "PVZRepository.Create"

	query, args, err := r.sq.Insert("pvz").
		Columns("id", "registration_date", "city").
		Values(pvz.ID, pvz.RegistrationDate, pvz.City).
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

// GetByID находит ПВЗ по ID.
func (r *PVZRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.PVZ, error) {
	const op = "PVZRepository.GetByID"

	query, args, err := r.sq.Select("id", "registration_date", "city").
		From("pvz").
		Where(sq.Eq{"id": id}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, r.wrapErr(op, fmt.Errorf("failed to build query: %w", err))
	}

	r.logQuery(ctx, op, query, args...)
	row := r.db.QueryRow(ctx, query, args...)

	var pvz domain.PVZ
	err = row.Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City)
	if err != nil {
		return nil, r.wrapErr(op, err)
	}
	return &pvz, nil
}

// ListIDsAndTotal возвращает слайс ID ПВЗ для текущей страницы с учетом фильтров
// и общее количество ПВЗ, удовлетворяющих фильтрам.
func (r *PVZRepository) ListIDsAndTotal(ctx context.Context, limit, offset int, startDate, endDate *time.Time) ([]uuid.UUID, int, error) {
	const op = "PVZRepository.ListIDsAndTotal"
	log := r.log.With(slog.String("op", op))

	countQueryBuilder := r.sq.Select("count(DISTINCT p.id)").From("pvz p")
	if startDate != nil || endDate != nil {
		countQueryBuilder = countQueryBuilder.Join("receptions r ON p.id = r.pvz_id")
		if startDate != nil {
			countQueryBuilder = countQueryBuilder.Where(sq.GtOrEq{"r.date_time": startDate})
		}
		if endDate != nil {
			countQueryBuilder = countQueryBuilder.Where(sq.LtOrEq{"r.date_time": endDate})
		}
	}

	countSql, countArgs, err := countQueryBuilder.ToSql()
	if err != nil {
		return nil, 0, r.wrapErr(op, fmt.Errorf("failed to build count query: %w", err))
	}

	r.logQuery(ctx, op+"_count", countSql, countArgs...)
	var total int
	err = r.db.QueryRow(ctx, countSql, countArgs...).Scan(&total)
	if err != nil {
		log.Error("Failed to count PVZs", slog.String("error", err.Error()))
		return nil, 0, r.wrapErr(op, fmt.Errorf("counting pvz: %w", err))
	}

	if total == 0 {
		log.Debug("No PVZs found matching criteria (for ListIDsAndTotal)")
		return []uuid.UUID{}, 0, nil
	}
	log.Debug("Total PVZs matching criteria", slog.Int("total", total))

	idQueryBuilder := r.sq.Select("DISTINCT p.id", "p.registration_date").From("pvz p")
	if startDate != nil || endDate != nil {
		idQueryBuilder = idQueryBuilder.Join("receptions r ON p.id = r.pvz_id")
		if startDate != nil {
			idQueryBuilder = idQueryBuilder.Where(sq.GtOrEq{"r.date_time": startDate})
		}
		if endDate != nil {
			idQueryBuilder = idQueryBuilder.Where(sq.LtOrEq{"r.date_time": endDate})
		}
	}
	idQueryBuilder = idQueryBuilder.OrderBy("p.registration_date DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	idsSql, idsArgs, err := idQueryBuilder.ToSql()
	if err != nil {
		return nil, 0, r.wrapErr(op, fmt.Errorf("failed to build paginated IDs query: %w", err))
	}

	r.logQuery(ctx, op+"_ids", idsSql, idsArgs...)
	rows, err := r.db.Query(ctx, idsSql, idsArgs...)
	if err != nil {
		log.Error("Failed to get PVZ IDs for page", slog.String("error", err.Error()))
		return nil, 0, r.wrapErr(op, fmt.Errorf("getting pvz ids: %w", err))
	}
	defer rows.Close()

	pvzIDs := make([]uuid.UUID, 0, limit)

	addedIds := make(map[uuid.UUID]bool)
	for rows.Next() {
		var id uuid.UUID
		var regDate time.Time
		if err := rows.Scan(&id, &regDate); err != nil {
			log.Error("Failed to scan PVZ ID", slog.String("error", err.Error()))
			return nil, 0, r.wrapErr(op, fmt.Errorf("scanning pvz id: %w", err))
		}
		if !addedIds[id] {
			pvzIDs = append(pvzIDs, id)
			addedIds[id] = true
		}
	}
	if err = rows.Err(); err != nil {
		log.Error("Error iterating PVZ IDs", slog.String("error", err.Error()))
		return nil, 0, r.wrapErr(op, fmt.Errorf("iterating pvz ids: %w", err))
	}

	log.Debug("Fetched PVZ IDs for page", slog.Any("pvz_ids", pvzIDs))
	return pvzIDs, total, nil
}

// GetByIDs находит и возвращает список ПВЗ по списку их ID.
func (r *PVZRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.PVZ, error) {
	const op = "PVZRepository.GetByIDs"
	log := r.log.With(slog.String("op", op))

	if len(ids) == 0 {
		log.Debug("No PVZ IDs provided, returning empty list")
		return []domain.PVZ{}, nil
	}
	log.Debug("Fetching PVZ data for IDs", slog.Any("pvz_ids", ids))

	query, args, err := r.sq.Select("id", "registration_date", "city").
		From("pvz").
		Where(sq.Eq{"id": ids}).
		OrderBy("registration_date DESC").
		ToSql()
	if err != nil {
		return nil, r.wrapErr(op, fmt.Errorf("failed to build pvz data query: %w", err))
	}

	r.logQuery(ctx, op, query, args...)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		if isErrNoRows(err) {
			log.Warn("No PVZs found for the given IDs", slog.Any("ids", ids))
			return []domain.PVZ{}, nil
		}
		log.Error("Failed to get PVZ data by IDs", slog.String("error", err.Error()))
		return nil, r.wrapErr(op, fmt.Errorf("getting pvz data: %w", err))
	}
	defer rows.Close()

	pvzs := make([]domain.PVZ, 0, len(ids))
	for rows.Next() {
		var p domain.PVZ
		if err := rows.Scan(&p.ID, &p.RegistrationDate, &p.City); err != nil {
			log.Error("Failed to scan PVZ data", slog.String("error", err.Error()))
			return nil, r.wrapErr(op, fmt.Errorf("scanning pvz data: %w", err))
		}
		pvzs = append(pvzs, p)
	}
	if err = rows.Err(); err != nil {
		log.Error("Error iterating PVZ data", slog.String("error", err.Error()))
		return nil, r.wrapErr(op, fmt.Errorf("iterating pvz data: %w", err))
	}

	log.Debug("Fetched PVZ data by IDs successfully", slog.Int("count", len(pvzs)))
	return pvzs, nil
}

// ListAll возвращает список всех ПВЗ без пагинации и фильтров (для gRPC).
func (r *PVZRepository) ListAll(ctx context.Context) ([]domain.PVZ, error) {
	const op = "PVZRepository.ListAll"

	query, args, err := r.sq.Select("id", "registration_date", "city").
		From("pvz").
		OrderBy("registration_date DESC").
		ToSql()
	if err != nil {
		return nil, r.wrapErr(op, fmt.Errorf("failed to build query: %w", err))
	}

	r.logQuery(ctx, op, query, args...)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		if isErrNoRows(err) {
			return []domain.PVZ{}, nil
		}
		return nil, r.wrapErr(op, err)
	}
	defer rows.Close()

	pvzs := make([]domain.PVZ, 0)
	for rows.Next() {
		var p domain.PVZ
		if err := rows.Scan(&p.ID, &p.RegistrationDate, &p.City); err != nil {
			return nil, r.wrapErr(op, fmt.Errorf("scanning pvz data: %w", err))
		}
		pvzs = append(pvzs, p)
	}
	if err = rows.Err(); err != nil {
		return nil, r.wrapErr(op, fmt.Errorf("iterating pvz data: %w", err))
	}

	return pvzs, nil
}
