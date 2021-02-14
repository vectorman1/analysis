package entity

type Symbol struct {
	ID                   uint
	Identifier           string `gorm:"not null"`
	Name                 string `gorm:"not null"`
	CurrencyID           uint   `gorm:"not null"`
	Currency             Currency
	ISIN                 string  `gorm:"not null"`
	MinimumOrderQuantity float32 `gorm:"not null"`
	MarketName           string  `gorm:"not null"`
	MarketHoursGMT       string  `gorm:"not null"`
	TechnicalID          uint
	Technical            Technical
}
