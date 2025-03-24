package routes

import (
	"cotizador-productos-eml/handlers"
	"cotizador-productos-eml/utils"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func RegisterProductoRoutes(router *gin.Engine, db *bun.DB) {
	handler := handlers.NewProductoHandler(db)

	// Grupo de rutas protegidas con middleware de autenticación y rol de admin
	productoRoutes := router.Group("/productos")
	productoRoutes.Use(utils.AuthMiddleware())
	productoRoutes.Use(utils.RoleMiddleware("admin"))
	{
		productoRoutes.POST("", handler.CreateProducto)
		productoRoutes.GET("", handler.GetAllProductos)
		productoRoutes.DELETE("/:id", handler.DeleteProducto)
		productoRoutes.PUT("/:id", handler.UpdateProducto)
		productoRoutes.GET("/:id", handler.GetProductoById)
	}

	// Grupo de rutas que solo requieren autenticación (sin permisos de admin)
	clientProductoRoutes := router.Group("/productos")
	clientProductoRoutes.Use(utils.AuthMiddleware())
	{
		clientProductoRoutes.GET("/get-products-clients", handler.GetProductosForClientes)
	}
}
