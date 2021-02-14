package entity

import (
	"github.com/lib/pq"
	"time"
)

type Technical struct {
	ID       uint
	SymbolID uint `gorm:"not null"`

	ExponentialMovingAverages pq.Float32Array `gorm:"not null;type:text"`
	SimpleMovingAverages      pq.Float32Array `gorm:"not null;type:text"`
	MACD                      pq.Float32Array `gorm:"not null;type:text"`
	RSI                       pq.Float32Array `gorm:"not null;type:text"`

	CreatedAt time.Time `gorm:"not null"`
}
