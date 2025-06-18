package models

import "time"

type Order struct {
	Number     string    `json:"number" db:"number"`
	Status     string    `json:"status" db:"status"`
	Accrual    *float64  `json:"accrual,omitempty" db:"accrual"`
	UploadedAt time.Time `json:"uploaded_at" db:"uploaded_at"`
	UserID     int64     `json:"-" db:"user_id"`
}
