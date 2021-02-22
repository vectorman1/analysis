package model

import "github.com/dgrijalva/jwt-go"

type Token struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.StandardClaims
}
