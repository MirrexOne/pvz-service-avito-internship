package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"pvz-service-avito-internship/internal/domain"
	"pvz-service-avito-internship/internal/handler/http/response"
	"pvz-service-avito-internship/pkg/jwt"
)

type contextKey string

const (
	UserIDKey   contextKey = "userID"
	UserRoleKey contextKey = "userRole"
)

type AuthMiddleware struct {
	log       *slog.Logger
	jwtSecret string
}

func NewAuthMiddleware(log *slog.Logger, jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{log: log, jwtSecret: jwtSecret}
}

func (m *AuthMiddleware) Authorize(c *gin.Context) {
	const op = "Middleware.Authorize"
	reqID := GetRequestIDFromContext(c)
	log := m.log.With(slog.String("op", op), slog.String("request_id", reqID))

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		log.Warn("Authorization header is missing")
		response.SendError(c, http.StatusUnauthorized, "Authorization header required")
		c.Abort()
		return
	}

	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 || !strings.EqualFold(headerParts[0], "Bearer") {
		log.Warn("Invalid Authorization header format", slog.String("header", authHeader))
		response.SendError(c, http.StatusUnauthorized, "Invalid Authorization header format (Bearer token expected)")
		c.Abort()
		return
	}

	tokenString := headerParts[1]
	if tokenString == "" {
		log.Warn("Token is empty")
		response.SendError(c, http.StatusUnauthorized, "Token is missing")
		c.Abort()
		return
	}

	claims, err := jwt.ValidateToken(tokenString, m.jwtSecret)
	if err != nil {
		log.Warn("Invalid or expired token", slog.String("error", err.Error()))
		response.SendError(c, http.StatusUnauthorized, "Invalid or expired token")
		c.Abort()
		return
	}

	userID := claims.UserID
	userRole := claims.Role

	if !userRole.IsValid() {
		log.Error("Invalid user role found in token claims", slog.String("role", string(userRole)), slog.String("user_id", userID.String()))
		response.SendError(c, http.StatusUnauthorized, "Invalid token claims (role)")
		c.Abort()
		return
	}

	c.Set(string(UserIDKey), userID)
	c.Set(string(UserRoleKey), userRole)

	log = log.With(slog.String("user_id", userID.String()), slog.String("role", string(userRole)))
	log.Debug("User authorized successfully")

	c.Next()
}

func RequireRole(allowedRoles ...domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "Middleware.RequireRole"
		reqID := GetRequestIDFromContext(c)
		log := slog.Default().With(slog.String("op", op), slog.String("request_id", reqID))

		roleValue, exists := c.Get(string(UserRoleKey))
		if !exists {
			log.Error("User role not found in context. Authorize middleware might be missing or failed.")
			response.SendError(c, http.StatusInternalServerError, "Internal server error (auth context missing)")
			c.Abort()
			return
		}

		userRole, ok := roleValue.(domain.UserRole)
		if !ok {
			log.Error("Invalid user role type in context", slog.Any("role_value", roleValue))
			response.SendError(c, http.StatusInternalServerError, "Internal server error (invalid auth context)")
			c.Abort()
			return
		}

		isAllowed := false
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			userID, _ := c.Get(string(UserIDKey))
			log.Warn("User role not allowed for this endpoint",
				slog.String("required_roles", fmt.Sprintf("%v", allowedRoles)),
				slog.String("user_role", string(userRole)),
				slog.Any("user_id", userID),
			)
			response.SendError(c, http.StatusForbidden, "Access forbidden: required role not met")
			c.Abort()
			return
		}

		log.Debug("User role check passed", slog.String("user_role", string(userRole)))
		c.Next()
	}
}
