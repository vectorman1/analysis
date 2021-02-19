package entity

import "time"

type Role int

const (
	AdminRole Role = iota
	UserRole
	ViewOnlyRole
)

type User struct {
	ID        uint
	Username  string
	Password  string
	Role      Role
	token     Token `gorm:"foreignKey:ID"`
	tokenID   uint
	CreatedAt time.Time
	UpdatedAt time.Time
}

type token interface {
	IsValid() (bool, error)
	SetValue() (string, error)
	GetValue() (string, error)
	Generate() (string, error)
}

type Token struct {
	token
	ID        uint
	value     string
	CreatedAt time.Time
	LastUsed  time.Time
}

func (t *Token) IsValid() (bool, error) {
	// throw error if no token present
	return false, nil
}

func (t *Token) SetValue() (string, error) {
	// validate
	return "", nil
}

func (t *Token) GetValue() (string, error) {
	// throw error if not valid
	return "", nil
}

func (t *Token) Generate() (string, error) {
	//generate token from permissions of user or throw error if has valid token
	return "", nil
}
