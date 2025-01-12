package models

import (
	"time"

	"github.com/uptrace/bun"
)

type Producto struct {
	bun.BaseModel      `bun:"productos"`
	ID                 int       `bun:"id,pk,autoincrement"`
	Nombre             string    `bun:"nombre"`
	PrecioVenta        int       `bun:"precio_venta"`
	PrecioBruto        int       `bun:"precio_bruto"`
	Disponible         bool      `bun:"disponible"`
	UltimaVezIngresado time.Time `bun:"ultima_vez_ingresado"`
	CreatedAt          time.Time `bun:"created_at"`
	UpdatedAt          time.Time `bun:"updated_at"`
}
