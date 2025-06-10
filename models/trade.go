package models

import (
	"time"

	"github.com/google/uuid"
)

type Trade struct {
	ID          uuid.UUID `json:"id" db:"id"`
	BuyOrderID  uuid.UUID `json:"buy_order_id" db:"buy_order_id"`
	SellOrderID uuid.UUID `json:"sell_order_id" db:"sell_order_id"`
	Symbol      string    `json:"symbol" db:"symbol"`
	Price       float64   `json:"price" db:"price"`
	Quantity    int       `json:"quantity" db:"quantity"`
	ExecutedAt  time.Time `json:"executed_at" db:"executed_at"`
}
