package entity

import "time"

type Role int

const (
	AdminRole Role = iota
	UserRole
)

type User struct {
	ID        uint
	Username  string
	Password  string
	Role      Role
	CreatedAt time.Time
	UpdatedAt time.Time
}
