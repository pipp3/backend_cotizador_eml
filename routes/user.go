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
	adminUserRoutes := router.Group("/user")
	adminUserRoutes.Use(utils.AuthMiddleware())
	adminUserRoutes.Use(utils.RoleMiddleware("admin"))
	{
		adminUserRoutes.GET("/get-users", handler.GetAllUsers)
		adminUserRoutes.PATCH("/update-user/:id", handler.UpdateUser)
		adminUserRoutes.DELETE("/delete-user/:id", handler.DeleteUser)
		adminUserRoutes.POST("/create-user", handler.CreateUser)
		adminUserRoutes.GET("/get-user/:id", handler.GetUserByID)
	}
}
