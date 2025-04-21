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

// ProductRepository реализует интерфейс domain.ProductRepository для PostgreSQL.
type ProductRepository struct {
	BaseRepository
}

// NewProductRepository создает новый экземпляр ProductRepository.
func NewProductRepository(db *pgxpool.Pool, log *slog.Logger) *ProductRepository {
	return &ProductRepository{
		BaseRepository: NewBaseRepository(db, log),
	}
}

// Create сохраняет новый товар в базу данных.
func (r *ProductRepository) Create(ctx context.Context, product *domain.Product) error {
	const op = "ProductRepository.Create"

	query, args, err := r.sq.Insert("products").
		Columns("id", "date_time", "type", "reception_id").
		Values(product.ID, product.DateTime, product.Type, product.ReceptionID).
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

// FindLastByReceptionID находит самый последний добавленный товар для указанной приемки.
func (r *ProductRepository) FindLastByReceptionID(ctx context.Context, receptionID uuid.UUID) (*domain.Product, error) {
	const op = "ProductRepository.FindLastByReceptionID"

	query, args, err := r.sq.Select("id", "date_time", "type", "reception_id").
		From("products").
		Where(sq.Eq{"reception_id": receptionID}).
		OrderBy("date_time DESC").
		Limit(1).
		ToSql()
	if err != nil {
		return nil, r.wrapErr(op, fmt.Errorf("failed to build query: %w", err))
	}

	r.logQuery(ctx, op, query, args...)
	row := r.db.QueryRow(ctx, query, args...)

	var prod domain.Product
	err = row.Scan(&prod.ID, &prod.DateTime, &prod.Type, &prod.ReceptionID)
	if err != nil {
		if isErrNoRows(err) {
			return nil, fmt.Errorf("repository.%s: %w", op, domain.ErrNoProductsToDelete)
		}
		return nil, r.wrapErr(op, err)
	}
	return &prod, nil
}

// DeleteByID удаляет товар по его ID.
func (r *ProductRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	const op = "ProductRepository.DeleteByID"

	query, args, err := r.sq.Delete("products").
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

// ListByReceptionIDs - вспомогательный метод для получения всех товаров для списка ID приемок.
func (r *ProductRepository) ListByReceptionIDs(ctx context.Context, receptionIDs []uuid.UUID) (map[uuid.UUID][]domain.Product, error) {
	const op = "ProductRepository.ListByReceptionIDs"
	log := r.log.With(slog.String("op", op))

	if len(receptionIDs) == 0 {
		log.Debug("No reception IDs provided, returning empty map")
		return make(map[uuid.UUID][]domain.Product), nil
	}
	log.Debug("Fetching products for reception IDs", slog.Any("reception_ids", receptionIDs))

	query, args, err := r.sq.Select("id", "date_time", "type", "reception_id").
		From("products").
		Where(sq.Eq{"reception_id": receptionIDs}).
		OrderBy("reception_id", "date_time ASC").
		ToSql()
	if err != nil {
		return nil, r.wrapErr(op, fmt.Errorf("failed to build query: %w", err))
	}

	r.logQuery(ctx, op, query, args...)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		if isErrNoRows(err) {
			log.Debug("No products found for the given reception IDs")
			return make(map[uuid.UUID][]domain.Product), nil
		}
		log.Error("Failed to query products", slog.String("error", err.Error()))
		return nil, r.wrapErr(op, fmt.Errorf("querying products: %w", err))
	}
	defer rows.Close()

	productsByReception := make(map[uuid.UUID][]domain.Product)
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(&p.ID, &p.DateTime, &p.Type, &p.ReceptionID); err != nil {
			log.Error("Failed to scan product", slog.String("error", err.Error()))
			return nil, r.wrapErr(op, fmt.Errorf("scanning product: %w", err))
		}
		productsByReception[p.ReceptionID] = append(productsByReception[p.ReceptionID], p)
	}
	if err = rows.Err(); err != nil {
		log.Error("Error iterating product rows", slog.String("error", err.Error()))
		return nil, r.wrapErr(op, fmt.Errorf("iterating products: %w", err))
	}

	log.Info("Successfully listed products by reception IDs", slog.Int("reception_count_with_products", len(productsByReception)))
	return productsByReception, nil
}
