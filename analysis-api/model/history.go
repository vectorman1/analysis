package model

import (
	"github.com/jackc/pgx/pgtype"
)

type History struct {
	ID       uint `json:"id"`
	SymbolID uint `json:"symbol_id"`

	Values pgtype.Float4Array `json:"values"`

	CreatedAt pgtype.Timestamptz `json:"created_at"`
	DeletedAt pgtype.Timestamptz `json:"deleted_at"`
}
