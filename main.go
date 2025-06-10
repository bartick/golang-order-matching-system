package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	internalDb "github.com/bartick/golang-order-matching-system/db"
	"github.com/bartick/golang-order-matching-system/internals"
	"github.com/bartick/golang-order-matching-system/service"
)

func main() {
	log.Println("Starting the application...")

	environmentConfig := internals.GetConfig()

	dbConnection := internalDb.ConnectDatabase(environmentConfig)
	if dbConnection == nil {
		log.Fatal("Failed to connect to the database")
	}
	log.Println("Database connection established successfully.")

	srv := service.NewWebServer(":"+environmentConfig.ServerPort, dbConnection)
	srv.Start()

	fmt.Println("Application is running...")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	signal.Notify(sig, syscall.SIGINT)
	<-sig
	log.Println("Shutting down the application gracefully...")

	srv.Shutdown()
	dbConnection.Close()
	log.Println("Application has been shut down.")
}
