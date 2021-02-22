package model

import "github.com/jackc/pgx/pgtype"

type Strategy struct {
	ID uint `json:"id"`

	Signals []Signal `json:"-"`

	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
	DeletedAt pgtype.Timestamptz `json:"deleted_at"`
}
