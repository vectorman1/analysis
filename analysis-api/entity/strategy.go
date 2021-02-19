package entity

import (
	"github.com/jackc/pgx/pgtype"
)

type Strategy struct {
	ID uint `gorm:"primarykey"`

	Signals []Signal `gorm:"foreignKey:StrategyID"`

	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz `gorm:"index"`
}
