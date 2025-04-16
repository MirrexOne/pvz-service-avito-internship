package http

import (
	"fmt"
	"log/slog"
	"net/http"
	"pvz-service-avito-internship/internal/handler/http/api"

	"github.com/gin-gonic/gin"

	"pvz-service-avito-internship/internal/domain"

	"pvz-service-avito-internship/internal/handler/http/response"
	mw "pvz-service-avito-internship/internal/middleware"
)

// Структура PVZHandler остается прежней
type PVZHandler struct {
	BaseHandler
	pvzService       domain.PVZService
	receptionService domain.ReceptionService
	productService   domain.ProductService
}

// Конструктор NewPVZHandler остается прежним
func NewPVZHandler(log *slog.Logger, pvzService domain.PVZService, receptionService domain.ReceptionService, productService domain.ProductService) *PVZHandler {
	return &PVZHandler{
		BaseHandler:      *NewBaseHandler(log),
		pvzService:       pvzService,
		receptionService: receptionService,
		productService:   productService,
	}
}

// GetPvz - ИСПРАВЛЕННАЯ ВЕРСИЯ
// Сигнатура теперь func(c *gin.Context)
// Парсинг параметров перенесен внутрь метода.
func (h *PVZHandler) GetPvz(c *gin.Context) {
	const op = "PVZHandler.GetPvz"
	reqID := mw.GetRequestIDFromContext(c)
	log := h.log.With(slog.String("op", op), slog.String("request_id", reqID))

	// --- Парсинг и валидация Query параметров ВНУТРИ хендлера ---
	page, err := h.parseIntQuery(c, "page", 1) // Используем хелпер из BaseHandler
	if err != nil {
		h.handleError(c, op, err) // Обрабатываем ошибку парсинга
		return
	}
	if page < 1 {
		page = 1 // Корректируем до минимума
	}

	limit, err := h.parseIntQuery(c, "limit", 10) // Используем хелпер
	if err != nil {
		h.handleError(c, op, err)
		return
	}
	if limit < 1 {
		limit = 1 // Корректируем до минимума
	}
	if limit > 30 { // Применяем максимум из спецификации
		limit = 30
	}

	startDate, err := h.parseDateTimeQuery(c, "startDate") // Используем хелпер
	if err != nil {
		h.handleError(c, op, err)
		return
	}
	endDate, err := h.parseDateTimeQuery(c, "endDate") // Используем хелпер
	if err != nil {
		h.handleError(c, op, err)
		return
	}
	// --- Конец парсинга параметров ---

	log = log.With(slog.Int("page", page), slog.Int("limit", limit))
	if startDate != nil {
		log = log.With(slog.Time("startDate", *startDate))
	}
	if endDate != nil {
		log = log.With(slog.Time("endDate", *endDate))
	}

	// Вызов сервиса для получения списка ПВЗ
	pvzs, total, err := h.pvzService.ListPVZs(c.Request.Context(), limit, page, startDate, endDate)
	if err != nil {
		h.handleError(c, op, err) // Обработка ошибок сервиса/репозитория
		return
	}

	log.Info("PVZs listed successfully", slog.Int("count", len(pvzs)), slog.Int("total", total))
	// Маппим результат в структуру ответа и отправляем 200 OK
	// Используем функцию маппинга из handler.go
	response.SendSuccess(c, http.StatusOK, toPVZListResponse(pvzs, total, page, limit))
}

// --- Остальные методы хендлера (PostPvz, CloseLastReception, DeleteLastProduct) остаются без изменений ---

// PostPvz ... (код без изменений)
func (h *PVZHandler) PostPvz(c *gin.Context) {
	const op = "PVZHandler.PostPvz"
	reqID := mw.GetRequestIDFromContext(c)
	log := h.log.With(slog.String("op", op), slog.String("request_id", reqID))
	var reqBody api.PostPvzJSONRequestBody
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		log.Warn("Failed to bind request", slog.String("error", err.Error()))
		response.SendError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %s", err.Error()))
		return
	}
	log = log.With(slog.String("city", string(reqBody.City)))
	pvz, err := h.pvzService.CreatePVZ(c.Request.Context(), domain.City(reqBody.City))
	if err != nil {
		h.handleError(c, op, err)
		return
	}
	log.Info("PVZ created successfully", slog.String("pvz_id", pvz.ID.String()))
	response.SendSuccess(c, http.StatusCreated, toPVZResponse(*pvz))
}

// CloseLastReception ... (код без изменений)
func (h *PVZHandler) CloseLastReception(c *gin.Context) {
	const op = "PVZHandler.CloseLastReception"
	reqID := mw.GetRequestIDFromContext(c)
	log := h.log.With(slog.String("op", op), slog.String("request_id", reqID))
	pvzID, err := h.parseUUID(c, "pvzId")
	if err != nil {
		h.handleError(c, op, err)
		return
	}
	log = log.With(slog.String("pvz_id", pvzID.String()))
	reception, err := h.receptionService.CloseReception(c.Request.Context(), pvzID)
	if err != nil {
		h.handleError(c, op, err)
		return
	}
	log.Info("Reception closed successfully", slog.String("reception_id", reception.ID.String()))
	response.SendSuccess(c, http.StatusOK, toReceptionResponse(*reception))
}

// DeleteLastProduct ... (код без изменений)
func (h *PVZHandler) DeleteLastProduct(c *gin.Context) {
	const op = "PVZHandler.DeleteLastProduct"
	reqID := mw.GetRequestIDFromContext(c)
	log := h.log.With(slog.String("op", op), slog.String("request_id", reqID))
	pvzID, err := h.parseUUID(c, "pvzId")
	if err != nil {
		h.handleError(c, op, err)
		return
	}
	log = log.With(slog.String("pvz_id", pvzID.String()))
	err = h.productService.DeleteLastProduct(c.Request.Context(), pvzID)
	if err != nil {
		h.handleError(c, op, err)
		return
	}
	log.Info("Last product deleted successfully")
	response.SendSuccess(c, http.StatusOK, nil)
}
