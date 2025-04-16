package http

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"pvz-service-avito-internship/internal/domain"
	"pvz-service-avito-internship/internal/handler/http/api"
	"pvz-service-avito-internship/internal/handler/http/response"
	mw "pvz-service-avito-internship/internal/middleware"
)

type AuthHandler struct {
	BaseHandler
	authService domain.AuthService
}

func NewAuthHandler(log *slog.Logger, authService domain.AuthService) *AuthHandler {
	return &AuthHandler{
		BaseHandler: *NewBaseHandler(log),
		authService: authService,
	}
}

// PostDummyLogin - имя изменено согласно генератору
func (h *AuthHandler) PostDummyLogin(c *gin.Context) {
	const op = "AuthHandler.PostDummyLogin"
	reqID := mw.GetRequestIDFromContext(c)
	log := h.log.With(slog.String("op", op), slog.String("request_id", reqID))

	// Используем тип-обертку для тела запроса
	var reqBody api.PostDummyLoginJSONRequestBody

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		log.Warn("Failed to bind request", slog.String("error", err.Error()))
		response.SendError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %s", err.Error()))
		return
	}

	// Проверка на nil не нужна, т.к. role обязательна в DTO

	// Конвертируем api.PostDummyLoginJSONBodyRole в domain.UserRole
	domainRole := domain.UserRole(reqBody.Role)
	log = log.With(slog.String("role", string(domainRole)))

	token, err := h.authService.DummyLogin(c.Request.Context(), domainRole)
	if err != nil {
		h.handleError(c, op, err)
		return
	}

	log.Info("Dummy token generated successfully")
	response.SendSuccess(c, http.StatusOK, api.Token(token))
}

// PostRegister - имя изменено согласно генератору
func (h *AuthHandler) PostRegister(c *gin.Context) {
	const op = "AuthHandler.PostRegister"
	reqID := mw.GetRequestIDFromContext(c)
	log := h.log.With(slog.String("op", op), slog.String("request_id", reqID))

	var reqBody api.PostRegisterJSONRequestBody

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		log.Warn("Failed to bind request", slog.String("error", err.Error()))
		response.SendError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %s", err.Error()))
		return
	}

	// Конвертируем openapi_types.Email в string
	emailStr := string(reqBody.Email)
	// Конвертируем api.PostRegisterJSONBodyRole в domain.UserRole
	domainRole := domain.UserRole(reqBody.Role)
	log = log.With(slog.String("email", emailStr), slog.String("role", string(domainRole)))

	user, err := h.authService.Register(c.Request.Context(), emailStr, reqBody.Password, domainRole)
	if err != nil {
		h.handleError(c, op, err)
		return
	}

	log.Info("User registered successfully", slog.String("user_id", user.ID.String()))
	response.SendSuccess(c, http.StatusCreated, toUserResponse(*user)) // Используем маппер
}

// PostLogin - имя изменено согласно генератору
func (h *AuthHandler) PostLogin(c *gin.Context) {
	const op = "AuthHandler.PostLogin"
	reqID := mw.GetRequestIDFromContext(c)
	log := h.log.With(slog.String("op", op), slog.String("request_id", reqID))

	var reqBody api.PostLoginJSONRequestBody

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		log.Warn("Failed to bind request", slog.String("error", err.Error()))
		response.SendError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %s", err.Error()))
		return
	}

	emailStr := string(reqBody.Email)
	log = log.With(slog.String("email", emailStr))

	token, err := h.authService.Login(c.Request.Context(), emailStr, reqBody.Password)
	if err != nil {
		h.handleError(c, op, err)
		return
	}

	log.Info("User logged in successfully")
	response.SendSuccess(c, http.StatusOK, api.Token(token))
}
