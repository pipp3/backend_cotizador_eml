package db

import (
	"context"
	"cotizador-productos-eml/models"
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func ConnectDB() *bun.DB {

	// Obtener las credenciales de las variables de entorno

	// Obtener la URL de la base de datos desde las variables de entorno
	dsn := os.Getenv("DATABASE_PUBLIC_URL")
	if dsn == "" {
		log.Fatal("La variable de entorno DATABASE_PUBLIC_URL no está configurada")
	}

	// Crear la conexión usando pgdriver y la URL completa
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	// Validar si la conexión es exitosa
	if err := sqldb.Ping(); err != nil {
		log.Fatal("Error al verificar la conexión a la base de datos: ", err)
	}
	// Crear la instancia de Bun con la base de datos
	db := bun.NewDB(sqldb, pgdialect.New())

	// Ejecutar las migraciones automáticamente
	err := runMigrations(db)
	if err != nil {
		log.Fatal("Error al ejecutar migraciones: ", err)
	}

	return db
}

// runMigrations ejecuta las migraciones de la base de datos
func runMigrations(db *bun.DB) error {
	// Aquí especificamos los modelos que queremos migrar (crear las tablas)
	// Se ejecuta solo si las tablas no existen
	_, err := db.NewCreateTable().Model((*models.Producto)(nil)).IfNotExists().Exec(context.Background())
	if err != nil {
		return err
	}

	// Aquí se pueden agregar más migraciones si tienes más tablas o cambios
	// También puedes usar migraciones más complejas si las tienes predefinidas en archivos SQL.

	return nil
}
