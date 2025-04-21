package http

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"pvz-service-avito-internship/internal/domain"
	"pvz-service-avito-internship/internal/handler/http/api"
	"pvz-service-avito-internship/internal/handler/http/response"
	mw "pvz-service-avito-internship/internal/middleware"
)

type ReceptionHandler struct {
	BaseHandler
	receptionService domain.ReceptionService
}

func NewReceptionHandler(log *slog.Logger, receptionService domain.ReceptionService) *ReceptionHandler {
	return &ReceptionHandler{
		BaseHandler:      *NewBaseHandler(log),
		receptionService: receptionService,
	}
}

func (h *ReceptionHandler) PostReceptions(c *gin.Context) {
	const op = "ReceptionHandler.PostReceptions"
	reqID := mw.GetRequestIDFromContext(c)
	log := h.log.With(slog.String("op", op), slog.String("request_id", reqID))

	var reqBody api.PostReceptionsJSONRequestBody

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		log.Warn("Failed to bind request", slog.String("error", err.Error()))
		response.SendError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %s", err.Error()))
		return
	}

	if reqBody.PvzId == uuid.Nil {
		log.Warn("pvzId is missing or invalid in request body")
		response.SendError(c, http.StatusBadRequest, "pvzId is required and must be a valid UUID")
		return
	}

	domainPvzID := reqBody.PvzId
	log = log.With(slog.String("pvz_id", domainPvzID.String()))

	reception, err := h.receptionService.CreateReception(c.Request.Context(), domainPvzID)
	if err != nil {
		h.handleError(c, op, err)
		return
	}

	log.Info("Reception created successfully", slog.String("reception_id", reception.ID.String()))
	response.SendSuccess(c, http.StatusCreated, toReceptionResponse(*reception))
}

func (h *ReceptionHandler) CloseLastReception(c *gin.Context) {
	pvzIDParam := c.Param("pvzId")
	pvzID, err := uuid.Parse(pvzIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid UUID format"})
		return
	}

	reception, err := h.receptionService.CloseReception(c.Request.Context(), pvzID)
	if err != nil {
		if errors.Is(err, domain.ErrPVZNotFound) || errors.Is(err, domain.ErrReceptionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":     reception.ID.String(),
		"pvzId":  reception.PVZID.String(),
		"status": reception.Status,
	})
}
