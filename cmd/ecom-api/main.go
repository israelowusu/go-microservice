package main

import (
	"log"

	"github.com/israelowusu/go-microservice.git/db"
	"github.com/israelowusu/go-microservice.git/ecom-api/handler"
	"github.com/israelowusu/go-microservice.git/ecom-api/server"
	"github.com/israelowusu/go-microservice.git/ecom-api/storer"
)

func main() {
	db, err := db.NewDatabase()
	if err != nil {
		log.Fatalf("error opening database: %v", err)
	}

	defer db.Close()
	log.Println("successfully connected to database")

	// Do something with the database
	st := storer.NewPostgresStorer(db.GetDB())
	srv := server.NewServer(st)
	hdl := handler.NewHandler(srv)
	handler.ResgisterRoutes(hdl)
	handler.Start(":8080")
}
