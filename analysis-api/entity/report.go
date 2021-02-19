package entity

import (
	"github.com/jackc/pgx/pgtype"
)

type Report struct {
	ID       uint
	SymbolID uint
	Symbol   Symbol

	ExponentialMovingAverages pgtype.Float4Array `gorm:"not null;type:text"`
	SimpleMovingAverages      pgtype.Float4Array `gorm:"not null;type:text"`
	MACD                      pgtype.Float4Array `gorm:"not null;type:text"`
	RSI                       pgtype.Float4Array `gorm:"not null;type:text"`

	CreatedAt pgtype.Timestamptz `gorm:"not null"`
	UpdatedAt pgtype.Timestamptz `gorm:"not null"`
	DeletedAt pgtype.Timestamptz `gorm:"index"`
}
