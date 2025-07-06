package jwt_tokens

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
)

// VerifyToken verifies a JWT token and returns its claims
func VerifyToken(tokenString string, secretKey string) (jwt.MapClaims, error) {
	// Парсим токен
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем алгоритм подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}
	// Проверяем, валиден ли токен
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	// Извлекаем claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	return claims, nil
}
