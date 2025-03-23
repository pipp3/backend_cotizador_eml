package models

import (
	"time"

	"github.com/uptrace/bun"
)

type DetallePedido struct {
	bun.BaseModel  `bun:"detalle_pedido"`
	ID             int       `bun:"id,pk,autoincrement"`
	PedidoID       int       `bun:"pedido_id"`
	Pedido         *Pedido   `bun:"rel:belongs-to,join:pedido_id=id"`
	ProductoID     int       `bun:"producto_id"`
	Producto       *Producto `bun:"rel:belongs-to,join:producto_id=id"`
	Cantidad       int       `bun:"cantidad"`
	PrecioUnitario int       `bun:"precio_unitario"`
	PrecioTotal    int       `bun:"precio_total"`
	CreatedAt      time.Time `bun:"created_at"`
	UpdatedAt      time.Time `bun:"updated_at"`
}
