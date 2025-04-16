package middleware

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	// Используем относительные пути в рамках проекта
	"pvz-service-avito-internship/internal/domain"
	"pvz-service-avito-internship/internal/handler/http/response" // Путь к хелперу ответов
	"pvz-service-avito-internship/pkg/jwt"                        // Путь к JWT утилитам
)

// contextKey - неэкспортируемый тип для ключей контекста, чтобы избежать коллизий.
type contextKey string

// Константы для ключей контекста, используемых для передачи информации о пользователе.
const (
	UserIDKey   contextKey = "userID"
	UserRoleKey contextKey = "userRole"
)

// AuthMiddleware содержит зависимости и методы для middleware аутентификации/авторизации.
type AuthMiddleware struct {
	log       *slog.Logger
	jwtSecret string
}

// NewAuthMiddleware создает новый экземпляр AuthMiddleware.
func NewAuthMiddleware(log *slog.Logger, jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{log: log, jwtSecret: jwtSecret}
}

// Authorize - это Gin middleware функция, которая:
//  1. Извлекает JWT токен из заголовка Authorization.
//  2. Валидирует токен с использованием jwtSecret.
//  3. В случае успеха, помещает UserID и UserRole в контекст Gin (`c.Set`).
//  4. В случае ошибки (отсутствие заголовка, неверный формат, невалидный токен),
//     прерывает цепочку обработки (`c.Abort()`) и отправляет ответ 401 Unauthorized.
func (m *AuthMiddleware) Authorize(c *gin.Context) {
	const op = "Middleware.Authorize"
	// Получаем request_id из контекста (должен быть установлен предыдущим middleware)
	reqID := GetRequestIDFromContext(c) // Используем хелпер
	log := m.log.With(slog.String("op", op), slog.String("request_id", reqID))

	// Получаем заголовок Authorization
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		log.Warn("Authorization header is missing")
		response.SendError(c, http.StatusUnauthorized, "Authorization header required")
		c.Abort() // Прерываем обработку
		return
	}

	// Проверяем формат заголовка "Bearer <token>"
	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 || !strings.EqualFold(headerParts[0], "Bearer") { // Сравнение без учета регистра
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

	// Валидируем токен
	claims, err := jwt.ValidateToken(tokenString, m.jwtSecret)
	if err != nil {
		log.Warn("Invalid or expired token", slog.String("error", err.Error()))
		response.SendError(c, http.StatusUnauthorized, "Invalid or expired token")
		c.Abort()
		return
	}

	// Токен валиден, извлекаем данные и добавляем в контекст Gin
	userID := claims.UserID
	userRole := claims.Role

	// Дополнительная проверка валидности роли из токена
	if !userRole.IsValid() {
		log.Error("Invalid user role found in token claims", slog.String("role", string(userRole)), slog.String("user_id", userID.String()))
		response.SendError(c, http.StatusUnauthorized, "Invalid token claims (role)") // Считаем токен невалидным
		c.Abort()
		return
	}

	// Сохраняем ID и роль в контексте Gin для доступа в последующих хендлерах/сервисах
	c.Set(string(UserIDKey), userID)
	c.Set(string(UserRoleKey), userRole)

	// Добавляем информацию о пользователе в логгер запроса для последующих логов
	log = log.With(slog.String("user_id", userID.String()), slog.String("role", string(userRole)))
	log.Debug("User authorized successfully") // Используем Debug, т.к. это штатная ситуация

	// Передаем управление следующему middleware или хендлеру
	c.Next()
}

// RequireRole возвращает Gin middleware, которое проверяет, имеет ли пользователь,
// аутентифицированный предыдущим Authorize middleware, одну из разрешенных ролей.
// Если роль не соответствует, прерывает обработку и возвращает 403 Forbidden.
func RequireRole(allowedRoles ...domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "Middleware.RequireRole"
		reqID := GetRequestIDFromContext(c)
		log := slog.Default().With(slog.String("op", op), slog.String("request_id", reqID)) // Используем дефолтный логгер, т.к. логгер запроса может быть не настроен здесь

		// Пытаемся получить роль пользователя из контекста Gin
		roleValue, exists := c.Get(string(UserRoleKey))
		if !exists {
			// Эта ситуация критическая - Authorize middleware должен был установить роль
			log.Error("User role not found in context. Authorize middleware might be missing or failed.")
			response.SendError(c, http.StatusInternalServerError, "Internal server error (auth context missing)")
			c.Abort()
			return
		}

		// Проверяем тип значения из контекста
		userRole, ok := roleValue.(domain.UserRole)
		if !ok {
			log.Error("Invalid user role type in context", slog.Any("role_value", roleValue))
			response.SendError(c, http.StatusInternalServerError, "Internal server error (invalid auth context)")
			c.Abort()
			return
		}

		// Проверяем, есть ли роль пользователя в списке разрешенных
		isAllowed := false
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				isAllowed = true
				break
			}
		}

		// Если роль не разрешена
		if !isAllowed {
			userID, _ := c.Get(string(UserIDKey)) // Получаем ID для лога (игнорируем ошибку, т.к. роль важнее)
			log.Warn("User role not allowed for this endpoint",
				slog.String("required_roles", fmt.Sprintf("%v", allowedRoles)),
				slog.String("user_role", string(userRole)),
				slog.Any("user_id", userID), // Логируем ID пользователя
			)
			response.SendError(c, http.StatusForbidden, "Access forbidden: required role not met")
			c.Abort()
			return
		}

		// Роль разрешена, передаем управление дальше
		log.Debug("User role check passed", slog.String("user_role", string(userRole)))
		c.Next()
	}
}

// --- Вспомогательные функции для получения данных из контекста ---

// GetUserIDFromContext извлекает UserID из контекста запроса (Gin или стандартного).
// Возвращает ошибку, если ID отсутствует или имеет неверный тип.
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	// Сначала пробуем получить из стандартного контекста
	userIDVal := ctx.Value(UserIDKey)
	if userIDVal != nil {
		userID, ok := userIDVal.(uuid.UUID)
		if !ok {
			return uuid.Nil, errors.New("invalid user ID type in context value")
		}
		return userID, nil
	}

	// Если нет в стандартном контексте, пробуем получить из Gin контекста (если он передан)
	if gCtx, ok := ctx.(*gin.Context); ok {
		userIDVal, exists := gCtx.Get(string(UserIDKey))
		if !exists {
			return uuid.Nil, errors.New("user ID not found in gin context")
		}
		userID, ok := userIDVal.(uuid.UUID)
		if !ok {
			return uuid.Nil, errors.New("invalid user ID type in gin context")
		}
		return userID, nil
	}

	// Если не нашли ни там, ни там
	return uuid.Nil, errors.New("user ID not found in context")
}

// GetUserRoleFromContext извлекает UserRole из контекста запроса (Gin или стандартного).
// Возвращает ошибку, если роль отсутствует или имеет неверный тип.
func GetUserRoleFromContext(ctx context.Context) (domain.UserRole, error) {
	// Сначала пробуем получить из стандартного контекста
	userRoleVal := ctx.Value(UserRoleKey)
	if userRoleVal != nil {
		userRole, ok := userRoleVal.(domain.UserRole)
		if !ok {
			return "", errors.New("invalid user role type in context value")
		}
		return userRole, nil
	}

	// Если нет в стандартном контексте, пробуем получить из Gin контекста (если он передан)
	if gCtx, ok := ctx.(*gin.Context); ok {
		userRoleVal, exists := gCtx.Get(string(UserRoleKey))
		if !exists {
			return "", errors.New("user role not found in gin context")
		}
		userRole, ok := userRoleVal.(domain.UserRole)
		if !ok {
			return "", errors.New("invalid user role type in gin context")
		}
		return userRole, nil
	}

	// Если не нашли ни там, ни там
	return "", errors.New("user role not found in context")
}
