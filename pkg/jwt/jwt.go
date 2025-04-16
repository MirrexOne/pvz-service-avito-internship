package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5" // Используем актуальную версию v5
	"github.com/google/uuid"

	// Используем относительные пути в рамках проекта
	"pvz-service-avito-internship/internal/domain"
)

// Claims определяет структуру данных (полезную нагрузку), хранимую в JWT токене.
// Встраиваем стандартные RegisteredClaims и добавляем свои поля.
type Claims struct {
	jwt.RegisteredClaims                 // Стандартные поля (iss, sub, aud, exp, nbf, iat, jti)
	UserID               uuid.UUID       `json:"user_id"` // Идентификатор пользователя
	Role                 domain.UserRole `json:"role"`    // Роль пользователя
}

// GenerateToken создает новый JWT токен с указанными данными пользователя, секретом и временем жизни (TTL).
func GenerateToken(userID uuid.UUID, role domain.UserRole, secret string, ttl time.Duration) (string, error) {
	const op = "jwt.GenerateToken" // Операция для контекста ошибок

	// Определяем время истечения токена
	expirationTime := time.Now().Add(ttl)

	// Создаем объект Claims с нашими данными
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// Устанавливаем время истечения (exp) и время выдачи (iat)
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			// Можно добавить другие стандартные поля при необходимости (Issuer, Subject)
			// Issuer: "pvz-service",
		},
		UserID: userID,
		Role:   role,
	}

	// Создаем новый токен с указанием алгоритма подписи (HS256) и нашими claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подписываем токен секретным ключом
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		// Оборачиваем ошибку подписи
		return "", fmt.Errorf("%s: failed to sign token: %w", op, err)
	}

	return signedToken, nil
}

// ValidateToken проверяет подпись и валидность (например, срок действия) JWT токена.
// Возвращает расшифрованные Claims, если токен валиден, или ошибку в противном случае.
func ValidateToken(tokenString string, secret string) (*Claims, error) {
	const op = "jwt.ValidateToken" // Операция для контекста ошибок

	// Парсим токен, используя наши Claims как структуру для полезной нагрузки.
	// Также передаем функцию для проверки ключа подписи (keyFunc).
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем, что алгоритм подписи токена - HMAC (как мы и ожидаем)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%s: unexpected signing method: %v", op, token.Header["alg"])
		}
		// Возвращаем секретный ключ для проверки подписи
		return []byte(secret), nil
	})

	// Обрабатываем ошибки парсинга/валидации, которые возвращает библиотека jwt
	if err != nil {
		// Ошибки могут быть разными: ErrTokenExpired, ErrTokenNotValidYet, ErrSignatureInvalid и т.д.
		// Оборачиваем общую ошибку, чтобы скрыть детали от вышестоящего слоя.
		return nil, fmt.Errorf("%s: failed to parse or validate token: %w", op, err)
	}

	// Проверяем, что токен валиден и что claims успешно распарсились в нашу структуру Claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Токен валиден, возвращаем claims
		return claims, nil
	}

	// Если что-то пошло не так (не удалось привести тип claims или token.Valid == false)
	return nil, fmt.Errorf("%s: invalid token (claims type assertion failed or token invalid)", op)
}
