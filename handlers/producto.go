package handlers

import (
	"cotizador-productos-eml/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type ProductoHandler struct {
	db *bun.DB
}

func NewProductoHandler(db *bun.DB) *ProductoHandler {
	return &ProductoHandler{db: db}
}

func (h *ProductoHandler) CreateProducto(c *gin.Context) {

	var input struct {
		Nombre             string `json:"nombre"`
		PrecioVenta        int    `json:"precio_venta"`
		PrecioBruto        int    `json:"precio_bruto"`
		UltimaVezIngresado string `json:"ultima_vez_ingresado"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Parsear la fecha
	fecha, err := time.Parse("02/01/2006", input.UltimaVezIngresado)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de fecha inválido"})
		return
	}
	nuevoProducto := models.Producto{
		Nombre:             input.Nombre,
		PrecioVenta:        input.PrecioVenta,
		PrecioBruto:        input.PrecioBruto,
		UltimaVezIngresado: fecha,
		Disponible:         true,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Insertar en la base de datos
	_, err = h.db.NewInsert().Model(&nuevoProducto).Exec(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"producto": nuevoProducto,
	})

}
