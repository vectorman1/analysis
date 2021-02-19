package entity

import (
	"github.com/jackc/pgx/pgtype"
)

type SignalType int

const (
	Ignore SignalType = iota
	Buy
	Sell
)

type Signal struct {
	ID         uint `gorm:"primarykey"`
	SymbolID   uint
	Symbol     Symbol
	StrategyID uint
	Strategy   Strategy

	Type      SignalType
	HistoryID uint `gorm:"type:text"`
	ReportID  uint `gorm:"type:text"`

	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz `gorm:"index"`
}
