package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"pvz-service-avito-internship/internal/domain"
	"pvz-service-avito-internship/internal/middleware"
)

// ReceptionService реализует интерфейс domain.ReceptionService.
type ReceptionService struct {
	log           *slog.Logger
	pvzRepo       domain.PVZRepository       // Зависимость для проверки существования ПВЗ
	receptionRepo domain.ReceptionRepository // Зависимость для работы с приемками
	metrics       domain.MetricsCollector    // Зависимость для сбора метрик
}

// NewReceptionService создает новый экземпляр ReceptionService.
func NewReceptionService(
	log *slog.Logger,
	pvzRepo domain.PVZRepository,
	receptionRepo domain.ReceptionRepository,
	metrics domain.MetricsCollector,
) *ReceptionService {
	return &ReceptionService{
		log:           log,
		pvzRepo:       pvzRepo,
		receptionRepo: receptionRepo,
		metrics:       metrics,
	}
}

// CreateReception инициирует новую приемку для ПВЗ.
func (s *ReceptionService) CreateReception(ctx context.Context, pvzID uuid.UUID) (*domain.Reception, error) {
	const op = "ReceptionService.CreateReception"
	reqID := middleware.GetRequestIDFromContext(ctx)
	log := s.log.With(slog.String("op", op), slog.String("request_id", reqID), slog.String("pvz_id", pvzID.String()))

	_, err := s.pvzRepo.GetByID(ctx, pvzID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			log.Warn("Attempt to create reception for non-existent PVZ")
			return nil, fmt.Errorf("%s: %w: pvz with id %s not found", op, domain.ErrValidation, pvzID)
		}
		log.Error("Failed to get PVZ by ID before creating reception", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseError, err)
	}

	_, err = s.receptionRepo.FindOpenByPVZID(ctx, pvzID)
	if err == nil {
		log.Warn("Attempt to create reception while another is in progress")
		return nil, fmt.Errorf("%s: %w", op, domain.ErrReceptionInProgress)
	}
	if !errors.Is(err, domain.ErrNoOpenReception) {
		log.Error("Failed to check for open reception", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseError, err)
	}
	log.Debug("No open reception found, proceeding to create a new one")

	reception := &domain.Reception{
		ID:       uuid.New(),
		DateTime: time.Now().UTC(),
		PVZID:    pvzID,
		Status:   domain.StatusInProgress,
	}

	err = s.receptionRepo.Create(ctx, reception)
	if err != nil {
		log.Error("Failed to create reception in repository", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseError, err)
	}

	s.metrics.IncReceptionsCreated()

	log.Info("Reception created successfully", slog.String("reception_id", reception.ID.String()))
	return reception, nil
}

// CloseReception закрывает последнюю активную приемку для ПВЗ.
func (s *ReceptionService) CloseReception(ctx context.Context, pvzID uuid.UUID) (*domain.Reception, error) {
	const op = "ReceptionService.CloseReception"
	reqID := middleware.GetRequestIDFromContext(ctx)
	log := s.log.With(slog.String("op", op), slog.String("request_id", reqID), slog.String("pvz_id", pvzID.String()))

	reception, err := s.receptionRepo.FindOpenByPVZID(ctx, pvzID)
	if err != nil {
		if errors.Is(err, domain.ErrNoOpenReception) {
			log.Warn("No open reception found to close")
			return nil, fmt.Errorf("%s: %w", op, domain.ErrReceptionClosed)
		}
		log.Error("Failed to find open reception to close", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseError, err)
	}
	log = log.With(slog.String("reception_id", reception.ID.String()))

	err = s.receptionRepo.UpdateStatus(ctx, reception.ID, domain.StatusClosed)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		log.Error("Failed to update reception status to closed", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseError, err)
	}

	reception.Status = domain.StatusClosed
	log.Info("Reception closed successfully")
	return reception, nil
}
