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

// ReceptionRepository реализует интерфейс domain.ReceptionRepository для PostgreSQL.
type ReceptionRepository struct {
	BaseRepository
}

// NewReceptionRepository создает новый экземпляр ReceptionRepository.
func NewReceptionRepository(db *pgxpool.Pool, log *slog.Logger) *ReceptionRepository {
	return &ReceptionRepository{
		BaseRepository: NewBaseRepository(db, log),
	}
}

// Create сохраняет новую приемку в базу данных.
func (r *ReceptionRepository) Create(ctx context.Context, reception *domain.Reception) error {
	const op = "ReceptionRepository.Create"

	query, args, err := r.sq.Insert("receptions").
		Columns("id", "date_time", "pvz_id", "status").
		Values(reception.ID, reception.DateTime, reception.PVZID, reception.Status).
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

// GetByID находит приемку по ID.
func (r *ReceptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Reception, error) {
	const op = "ReceptionRepository.GetByID"

	query, args, err := r.sq.Select("id", "date_time", "pvz_id", "status").
		From("receptions").
		Where(sq.Eq{"id": id}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, r.wrapErr(op, fmt.Errorf("failed to build query: %w", err))
	}

	r.logQuery(ctx, op, query, args...)
	row := r.db.QueryRow(ctx, query, args...)

	var rec domain.Reception
	err = row.Scan(&rec.ID, &rec.DateTime, &rec.PVZID, &rec.Status)
	if err != nil {
		return nil, r.wrapErr(op, err)
	}
	return &rec, nil
}

// FindOpenByPVZID находит последнюю активную ('in_progress') приемку для указанного ПВЗ.
func (r *ReceptionRepository) FindOpenByPVZID(ctx context.Context, pvzID uuid.UUID) (*domain.Reception, error) {
	const op = "ReceptionRepository.FindOpenByPVZID"

	query, args, err := r.sq.Select("id", "date_time", "pvz_id", "status").
		From("receptions").
		Where(sq.Eq{"pvz_id": pvzID, "status": domain.StatusInProgress}).
		OrderBy("date_time DESC").
		Limit(1).
		ToSql()
	if err != nil {
		return nil, r.wrapErr(op, fmt.Errorf("failed to build query: %w", err))
	}

	r.logQuery(ctx, op, query, args...)
	row := r.db.QueryRow(ctx, query, args...)

	var rec domain.Reception
	err = row.Scan(&rec.ID, &rec.DateTime, &rec.PVZID, &rec.Status)
	if err != nil {
		if isErrNoRows(err) {
			return nil, fmt.Errorf("repository.%s: %w", op, domain.ErrNoOpenReception)
		}
		return nil, r.wrapErr(op, err)
	}
	return &rec, nil
}

// UpdateStatus обновляет статус приемки по ее ID.
func (r *ReceptionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ReceptionStatus) error {
	const op = "ReceptionRepository.UpdateStatus"

	query, args, err := r.sq.Update("receptions").
		Set("status", status).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return r.wrapErr(op, fmt.Errorf("failed to build query: %w", err))
	}

	r.logQuery(ctx, op, query, args...)
	cmdTag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return r.wrapErr(op, err)
	}

	if cmdTag.RowsAffected() == 0 {
		return r.wrapErr(op, domain.ErrNotFound)
	}

	return nil
}

// ListByPVZIDsAndDate - вспомогательный метод для получения всех приемок и их товаров
// для списка ID ПВЗ в заданном диапазоне дат.
func (r *ReceptionRepository) ListByPVZIDsAndDate(ctx context.Context, pvzIDs []uuid.UUID, startDate, endDate *time.Time) (map[uuid.UUID][]domain.ReceptionWithProducts, error) {
	const op = "ReceptionRepository.ListByPVZIDsAndDate"
	log := r.log.With(slog.String("op", op))

	if len(pvzIDs) == 0 {
		log.Debug("No PVZ IDs provided, returning empty map")
		return make(map[uuid.UUID][]domain.ReceptionWithProducts), nil
	}
	log.Debug("Fetching receptions for PVZ IDs", slog.Any("pvz_ids", pvzIDs))

	receptionQueryBuilder := r.sq.Select("id", "date_time", "pvz_id", "status").
		From("receptions").
		Where(sq.Eq{"pvz_id": pvzIDs})

	if startDate != nil {
		receptionQueryBuilder = receptionQueryBuilder.Where(sq.GtOrEq{"date_time": startDate})
	}
	if endDate != nil {
		receptionQueryBuilder = receptionQueryBuilder.Where(sq.LtOrEq{"date_time": endDate})
	}

	receptionQueryBuilder = receptionQueryBuilder.OrderBy("pvz_id", "date_time DESC")

	recSql, recArgs, err := receptionQueryBuilder.ToSql()
	if err != nil {
		return nil, r.wrapErr(op, fmt.Errorf("failed to build reception query: %w", err))
	}

	r.logQuery(ctx, op+"_receptions", recSql, recArgs...)
	recRows, err := r.db.Query(ctx, recSql, recArgs...)
	if err != nil {
		if isErrNoRows(err) {
			log.Debug("No receptions found for the given PVZ IDs and date range")
			return make(map[uuid.UUID][]domain.ReceptionWithProducts), nil
		}
		log.Error("Failed to query receptions", slog.String("error", err.Error()))
		return nil, r.wrapErr(op, fmt.Errorf("querying receptions: %w", err))
	}
	defer recRows.Close()

	receptionsByPVZ := make(map[uuid.UUID][]domain.Reception)
	allReceptionIDs := make([]uuid.UUID, 0)
	for recRows.Next() {
		var rec domain.Reception
		if err := recRows.Scan(&rec.ID, &rec.DateTime, &rec.PVZID, &rec.Status); err != nil {
			log.Error("Failed to scan reception", slog.String("error", err.Error()))
			return nil, r.wrapErr(op, fmt.Errorf("scanning reception: %w", err))
		}
		receptionsByPVZ[rec.PVZID] = append(receptionsByPVZ[rec.PVZID], rec)
		allReceptionIDs = append(allReceptionIDs, rec.ID)
	}
	if err = recRows.Err(); err != nil {
		log.Error("Error iterating reception rows", slog.String("error", err.Error()))
		return nil, r.wrapErr(op, fmt.Errorf("iterating receptions: %w", err))
	}

	if len(allReceptionIDs) == 0 {
		log.Debug("No reception IDs collected, returning empty map")
		return make(map[uuid.UUID][]domain.ReceptionWithProducts), nil
	}
	log.Debug("Collected reception IDs", slog.Any("reception_ids", allReceptionIDs))

	productRepo := NewProductRepository(r.db, r.log)
	productsByReception, err := productRepo.ListByReceptionIDs(ctx, allReceptionIDs)
	if err != nil {
		log.Error("Failed to get products for receptions", slog.String("error", err.Error()))
		return nil, r.wrapErr(op, fmt.Errorf("getting products: %w", err))
	}
	log.Debug("Fetched related products")

	resultMap := make(map[uuid.UUID][]domain.ReceptionWithProducts, len(receptionsByPVZ))
	for pvzID, receptions := range receptionsByPVZ {
		receptionsWithProducts := make([]domain.ReceptionWithProducts, 0, len(receptions))
		for _, rec := range receptions {
			products := productsByReception[rec.ID]
			receptionsWithProducts = append(receptionsWithProducts, domain.ReceptionWithProducts{
				Reception: rec,
				Products:  products,
			})
		}
		resultMap[pvzID] = receptionsWithProducts
	}

	log.Info("Successfully listed receptions by PVZ IDs", slog.Int("pvz_count", len(resultMap)))
	return resultMap, nil
}
