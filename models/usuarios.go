package models

import (
	"time"

	"github.com/uptrace/bun"
)

type Usuario struct {
	bun.BaseModel `bun:"usuarios"`
	ID            int       `bun:"id,pk,autoincrement"`
	Nombre        string    `bun:"nombre"`
	Apellido      string    `bun:"apellido"`
	Email         string    `bun:"email"`
	Password      string    `bun:"password"`
	Rol           string    `bun:"rol,default:'cliente'"`
	Ciudad        string    `bun:"ciudad"`
	Celular       string    `bun:"celular"`
	Verificado    bool      `bun:"verificado,default:false"`
	CreatedAt     time.Time `bun:"created_at"`
	UpdatedAt     time.Time `bun:"updated_at"`
}
