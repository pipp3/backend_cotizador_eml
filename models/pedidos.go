package models

import (
	"time"

	"github.com/uptrace/bun"
)

type Pedido struct {
	bun.BaseModel    `bun:"pedidos"`
	ID               int        `bun:"id,pk,autoincrement"`
	UsuarioId        int        `bun:"usuario_id"`
	Usuario          *Usuario   `bun:"rel:belongs-to,join:usuario_id=id"`
	Total            int        `bun:"total"`
	Estado           string     `bun:"estado"`
	FechaEnvio       *time.Time `bun:"fecha_envio"`
	CiudadDestino    string     `bun:"ciudad_destino"`
	DireccionDestino string     `bun:"direccion_destino"`
	RutDestinatario  string     `bun:"rut_destinatario"`
	Company          string     `bun:"company"`
	TipoEnvio        string     `bun:"tipo_envio"`
	MetodoPago       string     `bun:"metodo_pago"`
	TipoDocumento    string     `bun:"tipo_documento"`
	CreatedAt        time.Time  `bun:"created_at"`
	UpdatedAt        time.Time  `bun:"updated_at"`
}
