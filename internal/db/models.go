package db

import "time"

type User struct {
	ID         int
	TelegramID int64
	CreatedAt  time.Time
}

type Payment struct {
	ID                int
	UserID            int
	YookassaPaymentID string
	Amount            float64
	Status            string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
