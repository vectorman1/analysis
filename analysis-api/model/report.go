package model

import "github.com/jackc/pgx/pgtype"

type Report struct {
	ID       uint   `json:"id"`
	SymbolID uint   `json:"symbol_id"`
	Symbol   Symbol `json:"-"`

	ExponentialMovingAverages pgtype.Float4Array `json:"exponential_moving_averages"`
	SimpleMovingAverages      pgtype.Float4Array `json:"simple_moving_averages"`
	MACD                      pgtype.Float4Array `json:"macd"`
	RSI                       pgtype.Float4Array `json:"rsi"`

	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
	DeletedAt pgtype.Timestamptz `json:"deleted_at"`
}
