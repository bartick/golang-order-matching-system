package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bartick/golang-order-matching-system/models"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type TradeResponse struct {
	TradeID     string `json:"trade_id"`
	Price       string `json:"price"`
	Quantity    string `json:"quantity"`
	TradeTime   string `json:"trade_time"`
	TradeSymbol string `json:"trade_symbol"`
}

func NewTradeResponse(tradeId, price, quantity, tradeTime, tradeSymbol string) *TradeResponse {
	return &TradeResponse{
		TradeID:     tradeId,
		Price:       price,
		Quantity:    quantity,
		TradeTime:   tradeTime,
		TradeSymbol: tradeSymbol,
	}
}

func AddTradeRoute(r *gin.Engine, db *sqlx.DB) {

	r.GET("/trades", func(c *gin.Context) {
		symbol := strings.ToUpper(c.Query("symbol"))
		if symbol == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Symbol parameter is required"})
			return
		}

		query := `SELECT id, buy_order_id, sell_order_id, symbol, price, quantity, executed_at 
			  FROM trades WHERE symbol = $1 ORDER BY executed_at DESC LIMIT 100`

		rows, err := db.Query(query, symbol)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch trades: %v", err)})
			return
		}
		defer rows.Close()

		var trades []models.Trade
		for rows.Next() {
			var trade models.Trade
			err := rows.Scan(&trade.ID, &trade.BuyOrderID, &trade.SellOrderID,
				&trade.Symbol, &trade.Price, &trade.Quantity, &trade.ExecutedAt)
			if err != nil {
				continue
			}
			trades = append(trades, trade)
		}

		c.JSON(http.StatusOK, gin.H{"trades": trades})

	})
}
