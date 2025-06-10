package service

import (
	"net/http"

	"github.com/bartick/golang-order-matching-system/api"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type WebServer struct {
	Addr         string
	router       *gin.Engine
	srv          *http.Server
	dbConnection *sqlx.DB
}

type WebServerInterface interface {
	Start() error
}

func NewWebServer(addr string, db *sqlx.DB) *WebServer {
	return &WebServer{
		Addr:         addr,
		router:       gin.Default(),
		dbConnection: db,
	}
}

func (ws *WebServer) Start() {

	api.AddPingRoute(ws.router)
	api.AddOrderRoute(ws.router, ws.dbConnection)
	api.AddOrderBookRoute(ws.router, ws.dbConnection)
	api.AddTradeRoute(ws.router, ws.dbConnection)

	ws.srv = &http.Server{
		Addr:    ws.Addr,
		Handler: ws.router.Handler(),
	}

	go func() {
		if err := ws.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic("listen: " + err.Error())
		}
	}()
}

func (ws *WebServer) Shutdown() error {
	if ws.srv != nil {
		return ws.srv.Shutdown(nil)
	}
	return nil
}
