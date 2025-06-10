package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bartick/golang-order-matching-system/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type OrderResponse struct {
	Order  models.Order   `json:"order"`
	Trades []models.Trade `json:"trades,omitempty"`
}

type OrderRequest struct {
	Symbol   string   `json:"symbol" binding:"required"`
	Side     string   `json:"side" binding:"required"`
	Type     string   `json:"type" binding:"required"`
	Price    *float64 `json:"price"`
	Quantity int      `json:"quantity" binding:"required,min=1"`
}

func NewOrderResponse(order models.Order, trades []models.Trade) *OrderResponse {
	return &OrderResponse{
		order,
		trades,
	}
}

func AddOrderRoute(r *gin.Engine, db *sqlx.DB) {
	r.POST("/orders", func(c *gin.Context) {
		placeOrder(c, db)
	})

	r.GET("/orders/:id", func(c *gin.Context) {
		getOrderStatus(c, db)
	})

	r.DELETE("/orders/:id", func(c *gin.Context) {
		cancelOrder(c, db)
	})
}

func placeOrder(c *gin.Context, db *sqlx.DB) {
	var req OrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate input
	if err := validateOrderRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Insert order
	order, err := insertOrder(tx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to insert order: %v", err)})
		return
	}

	// Match the order
	trades, err := matchOrder(tx, order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to match order: %v", err)})
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Return response
	response := OrderResponse{
		Order:  *order,
		Trades: trades,
	}
	c.JSON(http.StatusCreated, response)
}

func getOrderStatus(c *gin.Context, db *sqlx.DB) {
	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID format"})
		return
	}

	var order models.Order
	query := `SELECT id, symbol, side, type, price, initial_quantity, remaining_quantity, status, created_at, updated_at 
			  FROM orders WHERE id = $1`
	err = db.QueryRow(query, orderID).Scan(
		&order.ID, &order.Symbol, &order.Side, &order.Type, &order.Price,
		&order.InitialQuantity, &order.RemainingQuantity, &order.Status,
		&order.CreatedAt, &order.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Database error: %v", err)})
		return
	}

	c.JSON(http.StatusOK, order)
}

func cancelOrder(c *gin.Context, db *sqlx.DB) {
	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID format"})
		return
	}

	// Check if order exists and can be canceled
	var order models.Order
	query := `SELECT id, symbol, side, type, price, initial_quantity, remaining_quantity, status, created_at, updated_at 
			  FROM orders WHERE id = $1`
	err = db.QueryRow(query, orderID).Scan(
		&order.ID, &order.Symbol, &order.Side, &order.Type, &order.Price,
		&order.InitialQuantity, &order.RemainingQuantity, &order.Status,
		&order.CreatedAt, &order.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Database error: %v", err)})
		return
	}

	// Check if order can be canceled
	if order.Status == "filled" || order.Status == "canceled" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot cancel filled or already canceled order"})
		return
	}

	// Cancel the order
	updateQuery := `UPDATE orders SET status = 'canceled', updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err = db.Exec(updateQuery, orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel order"})
		return
	}

	order.Status = "canceled"
	order.UpdatedAt = time.Now()
	c.JSON(http.StatusOK, gin.H{"message": "Order canceled successfully", "order": order})
}

func validateOrderRequest(req *OrderRequest) error {
	// Validate side
	if req.Side != "buy" && req.Side != "sell" {
		return fmt.Errorf("side must be 'buy' or 'sell'")
	}

	// Validate type
	if req.Type != "limit" && req.Type != "market" {
		return fmt.Errorf("type must be 'limit' or 'market'")
	}

	// Validate price for limit orders
	if req.Type == "limit" {
		if req.Price == nil || *req.Price <= 0 {
			return fmt.Errorf("limit orders must have a positive price")
		}
	}

	// Validate symbol (basic format check)
	if len(req.Symbol) < 1 || len(req.Symbol) > 10 {
		return fmt.Errorf("symbol must be between 1 and 10 characters")
	}
	req.Symbol = strings.ToUpper(req.Symbol)

	return nil
}

func insertOrder(tx *sql.Tx, req OrderRequest) (*models.Order, error) {
	query := `INSERT INTO orders (symbol, side, type, price, initial_quantity, remaining_quantity, status) 
			  VALUES ($1, $2, $3, $4, $5, $6, 'open') RETURNING id, created_at, updated_at`

	order := &models.Order{
		Symbol:            strings.ToUpper(req.Symbol),
		Side:              req.Side,
		Type:              req.Type,
		Price:             req.Price,
		InitialQuantity:   req.Quantity,
		RemainingQuantity: req.Quantity,
		Status:            "open",
	}

	err := tx.QueryRow(query, order.Symbol, order.Side, order.Type, order.Price,
		order.InitialQuantity, order.RemainingQuantity).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert order: %w", err)
	}

	return order, nil
}

func matchOrder(tx *sql.Tx, order *models.Order) ([]models.Trade, error) {
	var trades []models.Trade

	for order.RemainingQuantity > 0 {
		// Find matching orders
		var matchingOrders []*models.Order
		var err error

		if order.Side == "buy" {
			matchingOrders, err = findMatchingSellOrders(tx, order)
		} else {
			matchingOrders, err = findMatchingBuyOrders(tx, order)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to find matching orders: %w", err)
		}

		if len(matchingOrders) == 0 {
			// No more matches
			if order.Type == "market" {
				// Cancel remaining quantity for market orders
				order.RemainingQuantity = 0
				order.Status = "filled"
			}
			break
		}

		// Execute trades with matching orders
		for _, matchingOrder := range matchingOrders {
			if order.RemainingQuantity == 0 {
				break
			}

			// Determine trade quantity and price
			tradeQuantity := min(order.RemainingQuantity, matchingOrder.RemainingQuantity)

			// Use the resting order's price for limit/limit matches
			// Use the limit price for market/limit matches
			var tradePrice float64
			if order.Type == "market" {
				tradePrice = *matchingOrder.Price
			} else if matchingOrder.Type == "market" {
				tradePrice = *order.Price
			} else {
				// Both are limit orders - use the resting (existing) order's price
				tradePrice = *matchingOrder.Price
			}

			// Create trade
			trade, err := createTrade(tx, order, matchingOrder, tradePrice, tradeQuantity)
			if err != nil {
				return nil, fmt.Errorf("failed to create trade: %w", err)
			}
			trades = append(trades, *trade)

			// Update order quantities
			order.RemainingQuantity -= tradeQuantity
			matchingOrder.RemainingQuantity -= tradeQuantity

			// Update orders in database
			err = updateOrderQuantity(tx, order)
			if err != nil {
				return nil, fmt.Errorf("failed to update order quantity: %w", err)
			}
			err = updateOrderQuantity(tx, matchingOrder)
			if err != nil {
				return nil, fmt.Errorf("failed to update matching order quantity: %w", err)
			}
		}
	}

	return trades, nil
}

func findMatchingSellOrders(tx *sql.Tx, buyOrder *models.Order) ([]*models.Order, error) {
	var query string
	var args []interface{}

	if buyOrder.Type == "market" {
		// Market buy order matches any sell order (lowest price first)
		query = `SELECT id, symbol, side, type, price, initial_quantity, remaining_quantity, status, created_at, updated_at 
				 FROM orders 
				 WHERE symbol = $1 AND side = 'sell' AND status IN ('open', 'partially_filled')
				 ORDER BY price ASC, created_at ASC`
		args = []interface{}{buyOrder.Symbol}
	} else {
		// Limit buy order matches sell orders at or below the buy price
		query = `SELECT id, symbol, side, type, price, initial_quantity, remaining_quantity, status, created_at, updated_at 
				 FROM orders 
				 WHERE symbol = $1 AND side = 'sell' AND status IN ('open', 'partially_filled') AND price <= $2
				 ORDER BY price ASC, created_at ASC`
		args = []interface{}{buyOrder.Symbol, *buyOrder.Price}
	}

	return queryOrders(tx, query, args...)
}

func findMatchingBuyOrders(tx *sql.Tx, sellOrder *models.Order) ([]*models.Order, error) {
	var query string
	var args []interface{}

	if sellOrder.Type == "market" {
		// Market sell order matches any buy order (highest price first)
		query = `SELECT id, symbol, side, type, price, initial_quantity, remaining_quantity, status, created_at, updated_at 
				 FROM orders 
				 WHERE symbol = $1 AND side = 'buy' AND status IN ('open', 'partially_filled')
				 ORDER BY price DESC, created_at ASC`
		args = []interface{}{sellOrder.Symbol}
	} else {
		// Limit sell order matches buy orders at or above the sell price
		query = `SELECT id, symbol, side, type, price, initial_quantity, remaining_quantity, status, created_at, updated_at 
				 FROM orders 
				 WHERE symbol = $1 AND side = 'buy' AND status IN ('open', 'partially_filled') AND price >= $2
				 ORDER BY price DESC, created_at ASC`
		args = []interface{}{sellOrder.Symbol, *sellOrder.Price}
	}

	return queryOrders(tx, query, args...)
}

func queryOrders(tx *sql.Tx, query string, args ...interface{}) ([]*models.Order, error) {
	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		order := &models.Order{}
		err := rows.Scan(
			&order.ID, &order.Symbol, &order.Side, &order.Type, &order.Price,
			&order.InitialQuantity, &order.RemainingQuantity, &order.Status,
			&order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func createTrade(tx *sql.Tx, order1, order2 *models.Order, price float64, quantity int) (*models.Trade, error) {
	var buyOrderID, sellOrderID uuid.UUID

	if order1.Side == "buy" {
		buyOrderID = order1.ID
		sellOrderID = order2.ID
	} else {
		buyOrderID = order2.ID
		sellOrderID = order1.ID
	}

	query := `INSERT INTO trades (buy_order_id, sell_order_id, symbol, price, quantity) 
			  VALUES ($1, $2, $3, $4, $5) RETURNING id, executed_at`

	trade := &models.Trade{
		BuyOrderID:  buyOrderID,
		SellOrderID: sellOrderID,
		Symbol:      order1.Symbol,
		Price:       price,
		Quantity:    quantity,
	}

	err := tx.QueryRow(query, trade.BuyOrderID, trade.SellOrderID, trade.Symbol, trade.Price, trade.Quantity).
		Scan(&trade.ID, &trade.ExecutedAt)
	if err != nil {
		return nil, err
	}

	return trade, nil
}

func updateOrderQuantity(tx *sql.Tx, order *models.Order) error {
	var status string
	if order.RemainingQuantity == 0 {
		status = "filled"
	} else if order.RemainingQuantity < order.InitialQuantity {
		status = "partially_filled"
	} else {
		status = "open"
	}

	query := `UPDATE orders SET remaining_quantity = $1, status = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3`
	_, err := tx.Exec(query, order.RemainingQuantity, status, order.ID)
	if err != nil {
		return err
	}

	order.Status = status
	return nil
}
