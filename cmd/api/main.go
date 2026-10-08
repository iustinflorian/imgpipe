package main

import (
	"imgpipe/pkg/db"
	"imgpipe/pkg/queue"
	"log"
)

func main() {
	mongoURI := "mongodb://localhost:27017"
	dbName := "imgpipe"

	_, err := db.ConnectMongo(mongoURI, dbName)
	if err != nil {
		log.Fatalf("error connecting to MongoDB: %v", err)
	}

	log.Println("Connection with MongoDB established!")

	rabbitURI := "amqp://guest:guest@localhost:5672/"
	mq, err := queue.NewRabbitMQ(rabbitURI)
	if err != nil {
		log.Fatalf("error connecting to RabbitMQ: %v", err)
	}

	log.Println("Connection with RabbitMQ established!")
	defer mq.Conn.Close()
	defer mq.Channel.Close()

}
