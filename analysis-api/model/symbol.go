package model

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/jackc/pgx/pgtype"
	"github.com/vectorman1/analysis/analysis-api/generated/proto_models"
)

type Symbol struct {
	ID         uint        `json:"-"`
	Uuid       pgtype.UUID `json:"uuid"`
	CurrencyID uint        `json:"currency_id"`
	Currency   Currency    `json:"-"`

	Isin                 string        `json:"isin"`
	Identifier           string        `json:"identifier"`
	Name                 string        `json:"name"`
	MinimumOrderQuantity pgtype.Float4 `json:"minimum_order_quantity"`
	MarketName           string        `json:"market_name"`
	MarketHoursGmt       string        `json:"market_hours_gmt"`

	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
	DeletedAt pgtype.Timestamptz `json:"deleted_at"`
}

func (s *Symbol) ToProtoObject() *proto_models.Symbol {
	createdAt, _ := ptypes.TimestampProto(s.CreatedAt.Time)
	updatedAt, _ := ptypes.TimestampProto(s.UpdatedAt.Time)
	deletedAt, _ := ptypes.TimestampProto(s.DeletedAt.Time)
	if s.DeletedAt.Status == pgtype.Null {
		deletedAt = nil
	}

	// db constraint
	var u string
	s.Uuid.AssignTo(&u)

	return &proto_models.Symbol{
		Id: uint64(s.ID),
		Currency: &proto_models.Currency{
			Id:       uint64(s.Currency.ID),
			Code:     s.Currency.Code,
			LongName: s.Currency.LongName,
		},
		Isin:                 s.Isin,
		Uuid:                 u,
		Identifier:           s.Identifier,
		Name:                 s.Name,
		MinimumOrderQuantity: s.MinimumOrderQuantity.Float,
		MarketName:           s.MarketName,
		MarketHoursGmt:       s.MarketHoursGmt,
		CreatedAt:            createdAt,
		UpdatedAt:            updatedAt,
		DeletedAt:            deletedAt,
	}
}
