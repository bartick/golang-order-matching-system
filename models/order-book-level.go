package models

type OrderBookLevel struct {
	Price         float64 `json:"price"`
	TotalQuantity int     `json:"total_quantity"`
	OrderCount    int     `json:"order_count"`
}
