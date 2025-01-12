package routes

import (
	"cotizador-productos-eml/handlers"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func RegisterProductoRoutes(router *gin.Engine, db *bun.DB) {
	handler := handlers.NewProductoHandler(db)
	productoRoutes := router.Group("/productos")
	{
		productoRoutes.POST("/", handler.CreateProducto)
		// Aquí puedes añadir más rutas como POST, PUT, DELETE, etc.
	}
}
