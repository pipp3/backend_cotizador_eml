package utils

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware es un middleware que protege las rutas privadas
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var secret = []byte(os.Getenv("JWT_SECRET"))
		var tokenString string

		// Primero intenta obtener el token desde la cookie
		cookieToken, err := c.Cookie("access_token")
		if err == nil {
			tokenString = cookieToken
		} else {
			// Si no está en la cookie, intenta obtenerlo desde la cabecera Authorization
			authHeader := c.GetHeader("Authorization")
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		// Si no se encontró el token en ninguno de los dos lugares, retorna error
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Token requerido"})
			c.Abort()
			return
		}

		// Verificar el token
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("método de firma inválido")
			}
			return secret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Token inválido o expirado"})
			c.Abort()
			return
		}

		// Extraer los datos del usuario
		userID, userOk := claims["sub"].(float64)
		email, emailOk := claims["email"].(string)
		if !userOk || !emailOk {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Token inválido o no se encontro el usuario"})
			c.Abort()
			return
		}

		// Agregar los datos del usuario al contexto
		c.Set("userID", int(userID))
		c.Set("email", email)

		c.Next()
	}
}
