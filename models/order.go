package models

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID                uuid.UUID `json:"id" db:"id"`
	Symbol            string    `json:"symbol" db:"symbol"`
	Side              string    `json:"side" db:"side"`
	Type              string    `json:"type" db:"type"`
	Price             *float64  `json:"price" db:"price"`
	InitialQuantity   int       `json:"initial_quantity" db:"initial_quantity"`
	RemainingQuantity int       `json:"remaining_quantity" db:"remaining_quantity"`
	Status            string    `json:"status" db:"status"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}
