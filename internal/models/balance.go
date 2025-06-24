package models

import "time"

type Balance struct {
	Current   float64 `json:"current" db:"current"`
	Withdrawn float64 `json:"withdrawn" db:"withdrawn"`
}

type WithdrawalRequest struct {
	Order string  `json:"order" db:"order_number"`
	Sum   float64 `json:"sum" db:"sum"`
}

type Withdrawal struct {
	Order     string    `json:"order" db:"order_number"`
	Sum       float64   `json:"sum" db:"sum"`
	Processed time.Time `json:"processed_at" db:"processed_at"`
	UserID    int64     `json:"-" db:"user_id"`
}
