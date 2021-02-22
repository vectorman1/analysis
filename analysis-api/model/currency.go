package model

type Currency struct {
	ID       uint   `json:"id"`
	Code     string `json:"code"`
	LongName string `json:"long_name"`
}
