package http

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	openapitypes "github.com/oapi-codegen/runtime/types"

	"pvz-service-avito-internship/internal/domain"
	"pvz-service-avito-internship/internal/handler/http/api"
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
		rootCause := GetRootCause(err)
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

func (h *BaseHandler) parseBoolQuery(c *gin.Context, s string, b bool) (interface{}, interface{}) {
	valueStr := c.Query(s)
	if valueStr == "" {
		return b, nil
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid boolean format for query parameter '%s'", domain.ErrValidation, s)
	}
	return value, nil
}

func GetRootCause(err error) error {
	for {
		unwrapped := errors.Unwrap(err)
		if unwrapped == nil {
			return err
		}
		err = unwrapped
	}
}

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
	apiID := product.ID
	receptionID := product.ReceptionID
	return api.Product{
		Id:          &apiID,
		DateTime:    &dateTime,
		Type:        api.ProductType(product.Type),
		ReceptionId: receptionID,
	}
}

func toReceptionResponse(reception domain.Reception) api.Reception {
	dateTime := reception.DateTime.UTC()
	apiID := reception.ID
	pvzID := reception.PVZID
	return api.Reception{
		Id:       &apiID,
		DateTime: dateTime,
		PvzId:    pvzID,
		Status:   api.ReceptionStatus(reception.Status),
	}
}

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
	apiID := user.ID
	apiEmail := openapitypes.Email(user.Email)
	return api.User{
		Id:    &apiID,
		Email: apiEmail,
		Role:  api.UserRole(user.Role),
	}
}
