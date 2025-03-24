package routes

import (
	"cotizador-productos-eml/handlers"
	"cotizador-productos-eml/utils"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func OrderRoutes(router *gin.Engine, db *bun.DB) {
	handler := handlers.NewOrderHandler(db)

	orderRoutes := router.Group("/orders")
	orderRoutes.Use(utils.AuthMiddleware())
	{
		orderRoutes.POST("/create-order", handler.CreateOrder)
		orderRoutes.GET("/get-user-orders", handler.GetUserOrders)
		orderRoutes.GET("/get-order-detail", handler.GetOrderDetail)
		orderRoutes.PATCH("/update-order-client", handler.UpdateOrderClient)
	}

	// Grupo de rutas protegidas con middleware de autenticación y rol de admin
	adminOrderRoutes := router.Group("/orders")
	adminOrderRoutes.Use(utils.AuthMiddleware())
	adminOrderRoutes.Use(utils.RoleMiddleware("admin"))
	{
		adminOrderRoutes.GET("/get-orders", handler.GetOrders)
		adminOrderRoutes.PATCH("/update-order-admin", handler.UpdateOrderAdmin)
	}
}
