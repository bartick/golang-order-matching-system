package db

import (
	"log"

	"github.com/bartick/golang-order-matching-system/internals"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func ConnectDatabase(c internals.Config) *sqlx.DB {
	db, err := sqlx.Connect("postgres", "user="+c.DBUser+" dbname="+c.DBName+" sslmode=disable password="+c.DBPassword+" host="+c.DBHost+" port="+c.DBPort)
	if err != nil {
		log.Fatalln(err)
	}

	// Test the connection to the database
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	} else {
		log.Println("Successfully Connected")
	}

	return db
}
