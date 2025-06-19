package models

type User struct {
	ID       int64  `json:"-" db:"id"`
	Login    string `json:"login" db:"login"`
	Password string `json:"password,omitempty" db:"password_hash"`
}
