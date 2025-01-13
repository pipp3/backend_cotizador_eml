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
		productoRoutes.POST("", handler.CreateProducto)
		productoRoutes.GET("", handler.GetAllProductos)
		productoRoutes.DELETE("/:id", handler.DeleteProducto)
		productoRoutes.PATCH("/:id", handler.UpdateProducto)
		productoRoutes.GET("/:id", handler.GetProductoById)
	}
}
