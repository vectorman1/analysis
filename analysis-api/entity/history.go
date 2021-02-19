package entity

import (
	"github.com/jackc/pgx/pgtype"
)

type History struct {
	ID       uint `gorm:"primarykey"`
	SymbolID uint

	RequestUuid pgtype.UUID        `gorm:"not null;unique;type:text"`
	Value       pgtype.Float4Array `gorm:"not null;type:text"`

	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz `gorm:"index"`
}
