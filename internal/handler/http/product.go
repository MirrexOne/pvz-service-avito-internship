package http

import (
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

type ProductHandler struct {
	BaseHandler
	productService domain.ProductService
}

func NewProductHandler(log *slog.Logger, productService domain.ProductService) *ProductHandler {
	return &ProductHandler{
		BaseHandler:    *NewBaseHandler(log),
		productService: productService,
	}
}

func (h *ProductHandler) PostProducts(c *gin.Context) {
	const op = "ProductHandler.PostProducts"
	reqID := mw.GetRequestIDFromContext(c)
	log := h.log.With(slog.String("op", op), slog.String("request_id", reqID))

	var reqBody api.PostProductsJSONRequestBody

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

	if reqBody.Type == "" {
		log.Warn("type is missing or empty in request body")
		response.SendError(c, http.StatusBadRequest, "type is required and cannot be empty")
		return
	}

	validTypes := map[string]bool{"электроника": true, "одежда": true, "обувь": true}
	if !validTypes[string(reqBody.Type)] {
		log.Warn("invalid product type in request body", slog.String("type", string(reqBody.Type)))
		response.SendError(c, http.StatusBadRequest, "invalid product type")
		return
	}

	domainPvzID := reqBody.PvzId
	domainProductType := domain.ProductType(reqBody.Type)
	log = log.With(slog.String("pvz_id", domainPvzID.String()), slog.String("type", string(domainProductType)))

	product, err := h.productService.AddProduct(c.Request.Context(), domainPvzID, domainProductType)
	if err != nil {
		h.handleError(c, op, err)
		return
	}

	log.Info("Product added successfully", slog.String("product_id", product.ID.String()))
	response.SendSuccess(c, http.StatusCreated, toProductResponse(*product))
}
