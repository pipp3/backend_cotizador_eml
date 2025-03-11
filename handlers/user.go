package handlers

import (
	"cotizador-productos-eml/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type UserHandler struct {
	db *bun.DB
}

func NewUserHandler(db *bun.DB) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) Me(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"error": "No Autorizado"})
		return
	}
	var usuario models.Usuario
	err := h.db.NewSelect().
		Model(&usuario).
		Where("id = ?", userID).
		Scan(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo usuario"})
		return
	}

	// Responder con los datos del usuario (sin incluir la contraseña)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":       usuario.ID,
			"email":    usuario.Email,
			"nombre":   usuario.Nombre,
			"apellido": usuario.Apellido,
			"celular":  usuario.Celular,
			"ciudad":   usuario.Ciudad,
		},
	})

}
