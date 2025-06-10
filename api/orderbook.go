package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/bartick/golang-order-matching-system/models"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func AddOrderBookRoute(r *gin.Engine, db *sqlx.DB) {
	r.GET("/orderbook", func(c *gin.Context) {
		symbol := strings.ToUpper(c.Query("symbol"))
		if symbol == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Symbol parameter is required"})
			return
		}

		orderBook := models.OrderBook{
			Symbol: symbol,
			Bids:   []models.OrderBookLevel{},
			Asks:   []models.OrderBookLevel{},
		}

		// Get bids (buy orders) - highest price first
		bidsQuery := `
		SELECT price, SUM(remaining_quantity) as total_quantity, COUNT(*) as order_count
		FROM orders 
		WHERE symbol = $1 AND side = 'buy' AND status IN ('open', 'partially_filled')
		GROUP BY price 
		ORDER BY price DESC
		LIMIT 10`

		rows, err := db.Query(bidsQuery, symbol)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch bids: %v", err)})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var level models.OrderBookLevel
			var price sql.NullFloat64
			err := rows.Scan(&price, &level.TotalQuantity, &level.OrderCount)
			if err != nil {
				continue
			}
			if price.Valid {
				level.Price = price.Float64
				orderBook.Bids = append(orderBook.Bids, level)
			}
		}

		// Get asks (sell orders) - lowest price first
		asksQuery := `
		SELECT price, SUM(remaining_quantity) as total_quantity, COUNT(*) as order_count
		FROM orders 
		WHERE symbol = $1 AND side = 'sell' AND status IN ('open', 'partially_filled')
		GROUP BY price 
		ORDER BY price ASC
		LIMIT 10`

		rows, err = db.Query(asksQuery, symbol)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch asks: %v", err)})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var level models.OrderBookLevel
			var price sql.NullFloat64
			err := rows.Scan(&price, &level.TotalQuantity, &level.OrderCount)
			if err != nil {
				continue
			}
			if price.Valid {
				level.Price = price.Float64
				orderBook.Asks = append(orderBook.Asks, level)
			}
		}

		c.JSON(http.StatusOK, orderBook)
	})
}
