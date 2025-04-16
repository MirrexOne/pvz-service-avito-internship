package http

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"                              // Используем стандартный UUID в домене
	openapi_types "github.com/oapi-codegen/runtime/types" // Импортируем типы OpenAPI

	"pvz-service-avito-internship/internal/domain"
	"pvz-service-avito-internship/internal/handler/http/api" // Используем сгенерированные типы
	"pvz-service-avito-internship/internal/handler/http/response"
	mw "pvz-service-avito-internship/internal/middleware"
)

type BaseHandler struct {
	log *slog.Logger
}

func NewBaseHandler(log *slog.Logger) *BaseHandler {
	return &BaseHandler{log: log}
}

func (h *BaseHandler) mapError(err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, "Resource not found"
	case errors.Is(err, domain.ErrValidation):
		rootCause := getRootCause(err)
		// Можно добавить более детальную обработку ошибок валидации от Gin binding
		return http.StatusBadRequest, fmt.Sprintf("Invalid request data: %s", rootCause.Error())
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, "Access forbidden"
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized, "Unauthorized"
	case errors.Is(err, domain.ErrPVZCityNotAllowed):
		return http.StatusBadRequest, domain.ErrPVZCityNotAllowed.Error()
	case errors.Is(err, domain.ErrReceptionInProgress):
		return http.StatusBadRequest, domain.ErrReceptionInProgress.Error()
	case errors.Is(err, domain.ErrNoOpenReception):
		return http.StatusBadRequest, domain.ErrNoOpenReception.Error()
	case errors.Is(err, domain.ErrReceptionClosed):
		return http.StatusBadRequest, domain.ErrReceptionClosed.Error()
	case errors.Is(err, domain.ErrProductDeletionOrder):
		return http.StatusBadRequest, domain.ErrProductDeletionOrder.Error()
	case errors.Is(err, domain.ErrNoProductsToDelete):
		return http.StatusBadRequest, domain.ErrNoProductsToDelete.Error()
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict, "Resource conflict"
	case errors.Is(err, domain.ErrDatabaseError):
		h.log.Error("Database error occurred", slog.String("original_error", err.Error()))
		return http.StatusInternalServerError, "Internal server error (database operation failed)"
	default:
		h.log.Error("Unknown internal error occurred", slog.String("error_type", fmt.Sprintf("%T", err)), slog.String("error", err.Error()))
		return http.StatusInternalServerError, "Internal server error"
	}
}

func (h *BaseHandler) handleError(c *gin.Context, op string, err error) {
	reqID := mw.GetRequestIDFromContext(c)
	log := h.log.With(slog.String("op", op), slog.String("request_id", reqID))
	statusCode, message := h.mapError(err)
	if statusCode >= http.StatusInternalServerError {
		log.Error("Internal server error mapped", slog.Int("status", statusCode), slog.String("message", message), slog.String("original_error", err.Error()))
	} else {
		log.Warn("Client error mapped", slog.Int("status", statusCode), slog.String("message", message), slog.String("original_error", err.Error()))
	}
	response.SendError(c, statusCode, message)
}

func (h *BaseHandler) parseUUID(c *gin.Context, paramName string) (uuid.UUID, error) {
	idStr := c.Param(paramName)
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: invalid UUID format for path parameter '%s'", domain.ErrValidation, paramName)
	}
	return id, nil
}

// parseIntQuery - без изменений
func (h *BaseHandler) parseIntQuery(c *gin.Context, paramName string, defaultValue int) (int, error) {
	valueStr := c.Query(paramName)
	if valueStr == "" {
		return defaultValue, nil
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid integer format for query parameter '%s'", domain.ErrValidation, paramName)
	}
	return value, nil
}

// parseDateTimeQuery - без изменений
func (h *BaseHandler) parseDateTimeQuery(c *gin.Context, paramName string) (*time.Time, error) {
	valueStr := c.Query(paramName)
	if valueStr == "" {
		return nil, nil
	}
	value, err := time.Parse(time.RFC3339, valueStr)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid date-time format (RFC3339 required) for query parameter '%s'", domain.ErrValidation, paramName)
	}
	t := value.UTC()
	return &t, nil
}

// getRootCause - без изменений
func getRootCause(err error) error {
	for {
		unwrapped := errors.Unwrap(err)
		if unwrapped == nil {
			return err
		}
		err = unwrapped
	}
}

// --- Функции маппинга Domain -> API DTO (Адаптированы под api.* типы) ---

func toPVZResponse(pvz domain.PVZ) api.PVZ {
	regDate := pvz.RegistrationDate.UTC()
	apiID := pvz.ID
	return api.PVZ{
		Id:               &apiID,
		RegistrationDate: &regDate,
		City:             api.PVZCity(pvz.City),
	}
}

func toProductResponse(product domain.Product) api.Product {
	dateTime := product.DateTime.UTC()
	apiID := openapi_types.UUID(product.ID)
	// Конвертируем стандартный UUID в openapi_types.UUID
	receptionID := openapi_types.UUID(product.ReceptionID)
	return api.Product{
		Id:          &apiID,    // Теперь указатель
		DateTime:    &dateTime, // Теперь указатель
		Type:        api.ProductType(product.Type),
		ReceptionId: receptionID, // Не указатель в сгенерированном типе
	}
}

func toReceptionResponse(reception domain.Reception) api.Reception {
	dateTime := reception.DateTime.UTC()
	apiID := openapi_types.UUID(reception.ID)
	pvzID := openapi_types.UUID(reception.PVZID)
	return api.Reception{
		Id:       &apiID,   // Теперь указатель
		DateTime: dateTime, // Не указатель в сгенерированном типе
		PvzId:    pvzID,    // Не указатель
		Status:   api.ReceptionStatus(reception.Status),
	}
}

// Структура для ответа GET /pvz (адаптирована под api.*)
type PVZListResponseItem struct {
	Pvz        api.PVZ                         `json:"pvz"`
	Receptions []ReceptionWithProductsResponse `json:"receptions"`
}
type ReceptionWithProductsResponse struct {
	Reception api.Reception `json:"reception"`
	Products  []api.Product `json:"products"`
}
type ListPVZResponse struct {
	Items []PVZListResponseItem `json:"items"`
	Total int                   `json:"total"`
	Page  int                   `json:"page"`
	Limit int                   `json:"limit"`
}

func toReceptionWithProductsResponse(rwp domain.ReceptionWithProducts) ReceptionWithProductsResponse {
	products := make([]api.Product, 0, len(rwp.Products))
	for _, p := range rwp.Products {
		products = append(products, toProductResponse(p))
	}
	return ReceptionWithProductsResponse{
		Reception: toReceptionResponse(rwp.Reception),
		Products:  products,
	}
}

func toPVZListResponseItem(pvzd domain.PVZWithDetails) PVZListResponseItem {
	receptions := make([]ReceptionWithProductsResponse, 0, len(pvzd.Receptions))
	for _, rwp := range pvzd.Receptions {
		receptions = append(receptions, toReceptionWithProductsResponse(rwp))
	}
	return PVZListResponseItem{
		Pvz:        toPVZResponse(pvzd.PVZ),
		Receptions: receptions,
	}
}

func toPVZListResponse(items []domain.PVZWithDetails, total, page, limit int) ListPVZResponse {
	respItems := make([]PVZListResponseItem, 0, len(items))
	for _, item := range items {
		respItems = append(respItems, toPVZListResponseItem(item))
	}
	return ListPVZResponse{
		Items: respItems,
		Total: total,
		Page:  page,
		Limit: limit,
	}
}

func toUserResponse(user domain.User) api.User {
	apiID := openapi_types.UUID(user.ID)
	// Конвертируем string email в openapi_types.Email
	apiEmail := openapi_types.Email(user.Email)
	return api.User{
		Id:    &apiID, // Теперь указатель
		Email: apiEmail,
		Role:  api.UserRole(user.Role),
	}
}
