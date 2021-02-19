package entity

import (
	"github.com/jackc/pgx/pgtype"
)

type Symbol struct {
	ID         uint `gorm:"not null;"`
	CurrencyID uint `gorm:"not null"`
	Currency   Currency
	HistoryID  uint
	History    History

	ISIN                 string        `gorm:"not null;"`
	Identifier           string        `gorm:"not null"`
	Name                 string        `gorm:"not null"`
	MinimumOrderQuantity pgtype.Float4 `gorm:"not null"`
	MarketName           string        `gorm:"not null"`
	MarketHoursGMT       string        `gorm:"not null"`

	Reports []Report `gorm:"foreignKey:SymbolID"`
	Signals []Signal `gorm:"foreignKey:SymbolID"`

	CreatedAt pgtype.Timestamptz `gorm:"not null"`
	UpdatedAt pgtype.Timestamptz `gorm:"not null"`
	DeletedAt pgtype.Timestamptz `gorm:"index"`
}
