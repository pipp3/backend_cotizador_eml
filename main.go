package main

import (
	"cotizador-productos-eml/db"
	"cotizador-productos-eml/routes"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	// Cargar el archivo .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error al cargar el archivo .env ", err)
	}

	db := db.ConnectDB()
	defer db.Close()
	// Configuración del router de Gin
	r := gin.Default()

	frontendUrl := os.Getenv("FRONTEND_URL")

	// Configurar CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{frontendUrl},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Registrar rutas
	routes.RegisterProductoRoutes(r, db)

	// Iniciar el servidor
	if err := r.Run(":4000"); err != nil {
		log.Fatal("Error al iniciar el servidor:", err)
	}
}
