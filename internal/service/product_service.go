package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	// Используем относительные пути в рамках проекта
	"pvz-service-avito-internship/internal/domain"
	"pvz-service-avito-internship/internal/middleware" // Для RequestID
)

// ProductService реализует интерфейс domain.ProductService.
type ProductService struct {
	log           *slog.Logger
	receptionRepo domain.ReceptionRepository // Зависимость для поиска открытой приемки
	productRepo   domain.ProductRepository   // Зависимость для работы с товарами
	metrics       domain.MetricsCollector    // Зависимость для сбора метрик
}

// NewProductService создает новый экземпляр ProductService.
func NewProductService(
	log *slog.Logger,
	receptionRepo domain.ReceptionRepository,
	productRepo domain.ProductRepository,
	metrics domain.MetricsCollector, // Добавлена зависимость
) *ProductService {
	return &ProductService{
		log:           log,
		receptionRepo: receptionRepo,
		productRepo:   productRepo,
		metrics:       metrics, // Сохраняем зависимость
	}
}

// AddProduct добавляет товар в текущую открытую приемку ПВЗ.
func (s *ProductService) AddProduct(ctx context.Context, pvzID uuid.UUID, productType domain.ProductType) (*domain.Product, error) {
	const op = "ProductService.AddProduct"
	reqID := middleware.GetRequestIDFromContext(ctx)
	log := s.log.With(slog.String("op", op), slog.String("request_id", reqID), slog.String("pvz_id", pvzID.String()), slog.String("type", string(productType)))

	// 1. Валидация типа товара (бизнес-правило)
	if !productType.IsValid() {
		log.Warn("Invalid product type provided")
		return nil, fmt.Errorf("%s: %w: invalid product type '%s'", op, domain.ErrValidation, productType)
	}

	reception, err := s.receptionRepo.FindOpenByPVZID(ctx, pvzID)
	if err != nil {
		if errors.Is(err, domain.ErrNoOpenReception) {
			log.Warn("Attempt to add product but no open reception found")
			return nil, fmt.Errorf("%s: %w", op, domain.ErrNoOpenReception)
		}

		log.Error("Failed to find open reception to add product", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, domain.ErrDatabaseError) // Ошибка уже обернута репозиторием
	}
	log = log.With(slog.String("reception_id", reception.ID.String())) // Добавляем ID приемки в лог

	product := &domain.Product{
		ID:          uuid.New(),       // Генерируем ID
		DateTime:    time.Now().UTC(), // Время добавления товара
		Type:        productType,
		ReceptionID: reception.ID, // Привязываем к найденной открытой приемке
	}

	err = s.productRepo.Create(ctx, product)
	if err != nil {
		log.Error("Failed to create product in repository", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, domain.ErrDatabaseError) // Ошибка уже обернута репозиторием
	}

	s.metrics.IncProductsAdded()

	log.Info("Product added successfully", slog.String("product_id", product.ID.String()))
	return product, nil
}

// DeleteLastProduct удаляет последний добавленный товар из открытой приемки (LIFO).
func (s *ProductService) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {
	const op = "ProductService.DeleteLastProduct"
	reqID := middleware.GetRequestIDFromContext(ctx)
	log := s.log.With(slog.String("op", op), slog.String("request_id", reqID), slog.String("pvz_id", pvzID.String()))

	reception, err := s.receptionRepo.FindOpenByPVZID(ctx, pvzID)
	if err != nil {
		if errors.Is(err, domain.ErrNoOpenReception) {
			log.Warn("Attempt to delete product but no open reception found")
			return fmt.Errorf("%s: %w", op, domain.ErrNoOpenReception)
		}
		log.Error("Failed to find open reception to delete product from", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, domain.ErrDatabaseError)
	}
	log = log.With(slog.String("reception_id", reception.ID.String()))

	lastProduct, err := s.productRepo.FindLastByReceptionID(ctx, reception.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNoProductsToDelete) {
			log.Warn("No products found in the reception to delete")
			return fmt.Errorf("%s: %w", op, domain.ErrNoProductsToDelete)
		}
		log.Error("Failed to find last product in reception", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, domain.ErrDatabaseError)
	}
	log = log.With(slog.String("product_id_to_delete", lastProduct.ID.String()))

	err = s.productRepo.DeleteByID(ctx, lastProduct.ID)
	if err != nil {
		log.Error("Failed to delete product by ID", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, domain.ErrDatabaseError)
	}

	log.Info("Last product deleted successfully")
	return nil
}
