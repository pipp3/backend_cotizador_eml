package main

import (
	"cotizador-productos-eml/db"
	"cotizador-productos-eml/routes"
	"log"

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

	// Registrar rutas
	routes.RegisterProductoRoutes(r, db)

	// Iniciar el servidor
	if err := r.Run(":4000"); err != nil {
		log.Fatal("Error al iniciar el servidor:", err)
	}
}
