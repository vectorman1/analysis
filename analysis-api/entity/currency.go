package entity

type Currency struct {
	ID       uint
	Code     string `gorm:"unique;not null"`
	LongName string `gorm:"not null"`
}
