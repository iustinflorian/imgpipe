package main

import (
	"imgpipe/pkg/db"
	"log"
)

func main() {
	mongoURI := "mongodb://localhost:27017"
	dbName := "imgpipe"

	_, err := db.ConnectMongo(mongoURI, dbName)
	if err != nil {
		log.Fatalf("critical error at running the api: %v", err)
	}

	log.Println("API started successfully and is connected to MongoDB!")
}
