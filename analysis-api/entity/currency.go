package entity

type Currency struct {
	ID      uint
	Code    string `gorm:"unique;not null"`
	Name    string `gorm:"not null"`
	Symbols []Symbol
}
