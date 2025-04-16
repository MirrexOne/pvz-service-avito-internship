package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	// Используем относительные пути в рамках проекта
	"pvz-service-avito-internship/internal/domain"
	"pvz-service-avito-internship/internal/middleware" // Для RequestID
)

// PVZService реализует интерфейс domain.PVZService.
type PVZService struct {
	log           *slog.Logger
	pvzRepo       domain.PVZRepository       // Зависимость от репозитория ПВЗ
	receptionRepo domain.ReceptionRepository // Зависимость для получения приемок в ListPVZs
	metrics       domain.MetricsCollector    // Зависимость для сбора метрик
}

// NewPVZService создает новый экземпляр PVZService.
func NewPVZService(
	log *slog.Logger,
	pvzRepo domain.PVZRepository,
	receptionRepo domain.ReceptionRepository, // Добавлена зависимость
	metrics domain.MetricsCollector,
) *PVZService {
	return &PVZService{
		log:           log,
		pvzRepo:       pvzRepo,
		receptionRepo: receptionRepo, // Сохраняем зависимость
		metrics:       metrics,
	}
}

// CreatePVZ создает новый ПВЗ.
func (s *PVZService) CreatePVZ(ctx context.Context, city domain.City) (*domain.PVZ, error) {
	const op = "PVZService.CreatePVZ"
	reqID := middleware.GetRequestIDFromContext(ctx)
	log := s.log.With(slog.String("op", op), slog.String("request_id", reqID), slog.String("city", string(city)))

	// 1. Валидация бизнес-правила: разрешенный город
	if !city.IsValid() {
		log.Warn("PVZ creation attempt in a non-allowed city")
		return nil, fmt.Errorf("%s: %w", op, domain.ErrPVZCityNotAllowed)
	}

	// 2. Создание доменного объекта PVZ
	pvz := &domain.PVZ{
		ID:               uuid.New(),       // Генерируем ID
		RegistrationDate: time.Now().UTC(), // Используем UTC
		City:             city,
	}

	// 3. Вызов репозитория для сохранения
	err := s.pvzRepo.Create(ctx, pvz)
	if err != nil {
		log.Error("Failed to create PVZ in repository", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseError, err)
	}

	// 4. Инкремент бизнесовой метрики
	s.metrics.IncPVZCreated()

	log.Info("PVZ created successfully", slog.String("pvz_id", pvz.ID.String()))
	return pvz, nil
}

// ListPVZs возвращает список ПВЗ с деталями приемок и товаров, с пагинацией и фильтрацией.
func (s *PVZService) ListPVZs(ctx context.Context, limit, page int, startDate, endDate *time.Time) ([]domain.PVZWithDetails, int, error) {
	const op = "PVZService.ListPVZs"
	reqID := middleware.GetRequestIDFromContext(ctx)
	log := s.log.With(
		slog.String("op", op),
		slog.String("request_id", reqID),
		slog.Int("limit", limit),
		slog.Int("page", page),
		slog.Any("startDate", startDate),
		slog.Any("endDate", endDate),
	)

	if limit <= 0 || page <= 0 {
		log.Error("Invalid pagination parameters received", slog.Int("limit", limit), slog.Int("page", page))
		// Это внутренняя ошибка, если хендлер пропустил
		return nil, 0, fmt.Errorf("%s: %w: invalid pagination parameters (limit/page must be positive)", op, domain.ErrInternalServer)
	}
	offset := (page - 1) * limit

	pvzIDs, total, err := s.pvzRepo.ListIDsAndTotal(ctx, limit, offset, startDate, endDate)
	if err != nil {
		log.Error("Failed to list PVZ IDs from repository", slog.String("error", err.Error()))
		return nil, 0, fmt.Errorf("%w: %v", domain.ErrDatabaseError, err)
	}

	if len(pvzIDs) == 0 {
		log.Info("No PVZs found for the given criteria/page")
		return []domain.PVZWithDetails{}, total, nil // Нет ПВЗ на этой странице
	}
	log.Debug("Fetched PVZ IDs for page", slog.Any("pvz_ids", pvzIDs), slog.Int("total", total))

	// 3. Получаем полные данные для найденных ID ПВЗ
	pvzs, err := s.pvzRepo.GetByIDs(ctx, pvzIDs)
	if err != nil {
		log.Error("Failed to get PVZ data by IDs from repository", slog.String("error", err.Error()))
		return nil, 0, fmt.Errorf("%w: %v", domain.ErrDatabaseError, err)
	}
	// Создаем мапу для быстрого доступа к ПВЗ по ID для шага 5
	pvzMap := make(map[uuid.UUID]domain.PVZ, len(pvzs))
	for _, p := range pvzs {
		pvzMap[p.ID] = p
	}
	log.Debug("Fetched PVZ data", slog.Int("count", len(pvzs)))

	// 4. Получаем связанные данные (Приемки и Товары) для этих ПВЗ
	receptionsMap, err := s.receptionRepo.ListByPVZIDsAndDate(ctx, pvzIDs, startDate, endDate)
	if err != nil {
		log.Error("Failed to get receptions and products details from repository", slog.String("error", err.Error()))
		return nil, 0, fmt.Errorf("%w: %v", domain.ErrDatabaseError, err)
	}
	log.Debug("Fetched related receptions and products", slog.Int("pvz_with_receptions_count", len(receptionsMap)))

	// 5. Собираем финальный результат, сохраняя порядок из pvzIDs/pvzs
	results := make([]domain.PVZWithDetails, 0, len(pvzIDs))
	for _, id := range pvzIDs { // Итерируемся по ID в правильном порядке
		pvzData, ok := pvzMap[id]
		if !ok {
			// Этого не должно произойти, если GetByIDs вернул данные для всех запрошенных ID
			log.Error("Consistency error: PVZ data not found for fetched ID", slog.String("pvz_id", id.String()))
			continue // Пропускаем этот ID
		}
		// Получаем приемки для этого ПВЗ из мапы (может быть nil)
		pvzReceptions := receptionsMap[id]
		results = append(results, domain.PVZWithDetails{
			PVZ:        pvzData,
			Receptions: pvzReceptions, // Добавляем приемки (или пустой/nil слайс)
		})
	}

	log.Info("PVZ list with details retrieved successfully", slog.Int("returned_count", len(results)), slog.Int("total_count", total))
	return results, total, nil
}
