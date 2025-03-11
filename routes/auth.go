package routes

import (
	"cotizador-productos-eml/handlers"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func AuthRoutes(router *gin.Engine, db *bun.DB) {
	handler := handlers.NewAuthHandler(db)
	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/register", handler.Register)
		authRoutes.GET("/verify-email", handler.VerifyEmail)
		authRoutes.POST("/resend-verification-email", handler.ResendVerificationEmail)
		authRoutes.POST("/login", handler.Login)
		authRoutes.POST("/refresh-token", handler.RefreshToken)
		authRoutes.GET("/logout", handler.Logout)
	}
}
