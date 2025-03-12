package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Genera token de verificación (puedes reutilizarlo para otros tipos de tokens)
func GenerateJWT(userID int, email, rol, secret string, expiresIn time.Duration, tokenType string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"rol":   rol,
		"exp":   time.Now().Add(expiresIn).Unix(),
		"type":  tokenType,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// Valida y parsea un token (útil para el endpoint de verificación)
func ParseJWT(tokenString, secret string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}
