package http

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"pvz-service-avito-internship/pkg/validator"

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

func (h *AuthHandler) PostDummyLogin(c *gin.Context) {
	const op = "AuthHandler.PostDummyLogin"
	reqID := mw.GetRequestIDFromContext(c)
	log := h.log.With(slog.String("op", op), slog.String("request_id", reqID))

	var reqBody api.PostDummyLoginJSONRequestBody

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		log.Warn("Failed to bind request", slog.String("error", err.Error()))
		response.SendError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %s", err.Error()))
		return
	}

	domainRole := domain.UserRole(reqBody.Role)
	log = log.With(slog.String("role", string(domainRole)))

	token, err := h.authService.DummyLogin(c.Request.Context(), domainRole)
	if err != nil {
		h.handleError(c, op, err)
		return
	}

	log.Info("Dummy token generated successfully")
	response.SendSuccess(c, http.StatusOK, token)
}

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

	type loginValidation struct {
		Password string `validate:"required,password"`
	}

	validateStruct := loginValidation{
		Password: reqBody.Password,
	}

	customValidator := validator.NewCustomValidator()
	if err := customValidator.Validate(&validateStruct); err != nil {
		log.Warn("Failed to validate request", slog.String("error", err.Error()))
		response.SendError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request data: %s", err.Error()))
		return
	}

	emailStr := string(reqBody.Email)
	domainRole := domain.UserRole(reqBody.Role)
	log = log.With(slog.String("email", emailStr), slog.String("role", string(domainRole)))

	user, err := h.authService.Register(c.Request.Context(), emailStr, reqBody.Password, domainRole)
	if err != nil {
		h.handleError(c, op, err)
		return
	}

	log.Info("User registered successfully", slog.String("user_id", user.ID.String()))
	response.SendSuccess(c, http.StatusCreated, toUserResponse(*user))
}

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

	type loginValidation struct {
		Password string `validate:"required,password"`
	}

	validateStruct := loginValidation{
		Password: reqBody.Password,
	}

	customValidator := validator.NewCustomValidator()
	if err := customValidator.Validate(&validateStruct); err != nil {
		log.Warn("Failed to validate request", slog.String("error", err.Error()))
		response.SendError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request data: %s", err.Error()))
		return
	}

	emailStr := string(reqBody.Email)
	log = log.With(slog.String("email", emailStr))

	token, err := h.authService.Login(c.Request.Context(), emailStr, reqBody.Password)
	if err != nil {
		if errors.Is(err, domain.ErrPassIsRequired) {
			response.SendError(c, http.StatusBadRequest, domain.ErrPassIsRequired.Error())
			return
		}

		h.handleError(c, op, err)
		return
	}

	log.Info("User logged in successfully")
	response.SendSuccess(c, http.StatusOK, token)
}
