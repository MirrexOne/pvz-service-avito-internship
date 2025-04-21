package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"pvz-service-avito-internship/internal/domain"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID       `json:"user_id"`
	Role   domain.UserRole `json:"role"`
}

func GenerateToken(userID uuid.UUID, role domain.UserRole, secret string, ttl time.Duration) (string, error) {
	const op = "jwt.GenerateToken"

	expirationTime := time.Now().Add(ttl)

	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: userID,
		Role:   role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("%s: failed to sign token: %w", op, err)
	}

	return signedToken, nil
}

func ValidateToken(tokenString string, secret string) (*Claims, error) {
	const op = "jwt.ValidateToken"

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%s: unexpected signing method: %v", op, token.Header["alg"])
		}

		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("%s: failed to parse or validate token: %w", op, err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("%s: invalid token (claims type assertion failed or token invalid)", op)
}
