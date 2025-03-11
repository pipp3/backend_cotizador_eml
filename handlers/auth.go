package handlers

import (
	"cotizador-productos-eml/models"
	"cotizador-productos-eml/utils"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db *bun.DB
}

func (h *AuthHandler) GenerateTokens(userID int, email string) (string, string, error) {

	// 2. Generar access token (vida más corta)
	accessToken, err := utils.GenerateJWT(
		userID, // Incluir userID
		email,
		os.Getenv("JWT_SECRET"),
		15*time.Minute, // 15 minutos (tiempo recomendado)
		"access",
	)
	if err != nil {
		return "", "", fmt.Errorf("error generando access token: %w", err)
	}

	// 3. Generar refresh token (vida más larga)
	refreshToken, err := utils.GenerateJWT(
		userID,
		email,
		os.Getenv("JWT_SECRET"),
		7*24*time.Hour, // 7 días (tiempo recomendado)
		"refresh",
	)
	if err != nil {
		return "", "", fmt.Errorf("error generando refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func NewAuthHandler(db *bun.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

func (h *AuthHandler) Register(c *gin.Context) {

	var input struct {
		Nombre   string `json:"nombre"`
		Apellido string `json:"apellido"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Ciudad   string `json:"ciudad"`
		Celular  string `json:"celular"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"succes": false,
			"error":  "Datos de registro invalidos: " + err.Error(),
		})
		return
	}

	// Verificar si el email ya está registrado
	var usuario models.Usuario
	err := h.db.NewSelect().Model(&usuario).Where("email = ?", input.Email).Scan(c)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"succes": false,
			"error":  "Email ya registrado",
		})
		return
	}
	//Encriptar contraseña
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"succes": false,
			"error":  "Error al encriptar contraseña: " + err.Error(),
		})
		return
	}
	input.Password = string(hash)

	// Crear nuevo usuario
	nuevoUsuario := models.Usuario{
		Nombre:    input.Nombre,
		Apellido:  input.Apellido,
		Email:     input.Email,
		Password:  input.Password,
		Rol:       "cliente",
		Ciudad:    input.Ciudad,
		Celular:   input.Celular,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	verificationToken, err := utils.GenerateJWT(
		nuevoUsuario.ID,
		nuevoUsuario.Email,
		os.Getenv("JWT_SECRET"), // Debes tener configurado el secreto
		24*time.Hour,
		"verification",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"succes": false,
			"error":  "Error al generar token de verificación: " + err.Error(),
		})
		return
	}
	// Enviar email de verificación
	if err := utils.SendVerificationEmail(nuevoUsuario.Email, verificationToken); err != nil {
		// Registrar el error pero no fallar el registro
		log.Printf("Error enviando email de verificación: %v", err)
	}

	// Insertar en la base de datos
	_, err = h.db.NewInsert().Model(&nuevoUsuario).Exec(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"succes": false,
			"error":  "Error al crear usuario" + err.Error(),
		})
		return
	}
	// Respuesta exitosa (sin datos sensibles)
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Registro exitoso. Verifica tu email para activar la cuenta.",
		"data": gin.H{
			"id":       nuevoUsuario.ID,
			"nombre":   nuevoUsuario.Nombre,
			"email":    nuevoUsuario.Email,
			"verified": nuevoUsuario.Verificado,
		},
	})

}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Token de verificación requerido",
		})
		return
	}

	// Parsear y validar el token JWT
	claims, err := utils.ParseJWT(token, os.Getenv("JWT_SECRET"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Token inválido o expirado",
		})
		return
	}

	// Verificar que el token es de tipo verificación
	if tokenType, ok := claims["type"].(string); !ok || tokenType != "verification" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Tipo de token inválido",
		})
		return
	}

	email, ok := claims["email"].(string)
	if !ok || email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Email inválido en el token",
		})
		return
	}

	// Actualizar usuario a verificado
	result, err := h.db.NewUpdate().
		Model((*models.Usuario)(nil)).
		Set("verificado = true").
		Where("email = ?", email).
		Exec(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Error al verificar el email",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Usuario no encontrado",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Email verificado exitosamente",
	})
}

func (h *AuthHandler) ResendVerificationEmail(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Email requerido: " + err.Error(),
		})
		return
	}

	// Verificar si el usuario existe y no está verificado
	var usuario models.Usuario
	err := h.db.NewSelect().
		Model(&usuario).
		Where("email = ?", input.Email).
		Scan(c)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Usuario no encontrado",
		})
		return
	}

	if usuario.Verificado {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "El email ya está verificado",
		})
		return
	}

	// Generar nuevo token
	verificationToken, err := utils.GenerateJWT(
		usuario.ID,
		usuario.Email,
		os.Getenv("JWT_SECRET"),
		24*time.Hour,
		"verification",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Error generando token de verificación",
		})
		return
	}

	// Enviar email
	if err := utils.SendVerificationEmail(usuario.Email, verificationToken); err != nil {
		log.Printf("Error enviando email de verificación: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Error al reenviar el email de verificación",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Email de verificación reenviado exitosamente",
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Datos inválidos: " + err.Error(),
		})
		return
	}

	var usuario models.Usuario
	err := h.db.NewSelect().
		Model(&usuario).
		Where("email = ?", input.Email).
		Scan(c)

	if err != nil || bcrypt.CompareHashAndPassword([]byte(usuario.Password), []byte(input.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Credenciales inválidas",
		})
		return
	}

	if !usuario.Verificado {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Cuenta no verificada",
		})
		return
	}

	accessToken, refreshToken, err := h.GenerateTokens(usuario.ID, usuario.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Error generando tokens",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"expires_in":    3600,
		},
	})
}
