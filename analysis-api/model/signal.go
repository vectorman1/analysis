package model

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
	ID         uint       `json:"id"`
	SymbolID   uint       `json:"symbol_id"`
	Symbol     Symbol     `json:"-"`
	StrategyID uint       `json:"strategy_id"`
	Type       SignalType `json:"type"`

	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
	DeletedAt pgtype.Timestamptz `json:"deleted_at"`
}
