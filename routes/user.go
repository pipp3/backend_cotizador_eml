package routes

import (
	"cotizador-productos-eml/handlers"
	"cotizador-productos-eml/utils"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func UserRoutes(router *gin.Engine, db *bun.DB) {
	handler := handlers.NewUserHandler(db)
	userRoutes := router.Group("/user")
	userRoutes.Use(utils.AuthMiddleware())
	{
		userRoutes.GET("/me", handler.Me)

	}
}
