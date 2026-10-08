package main

import (
	"imgpipe/pkg/db"
	"imgpipe/pkg/queue"
	"log"
	"net/http"
)

func main() {
	mongoURI := "mongodb://localhost:27017"
	dbName := "imgpipe"

	database, err := db.ConnectMongo(mongoURI, dbName)
	if err != nil {
		log.Fatalf("error connecting to MongoDB: %v", err)
	}

	repo := db.NewJobRepository(database)
	log.Println("Connection with MongoDB established!")

	rabbitURI := "amqp://guest:guest@localhost:5672/"
	mq, err := queue.NewRabbitMQ(rabbitURI)
	if err != nil {
		log.Fatalf("error connecting to RabbitMQ: %v", err)
	}

	log.Println("Connection with RabbitMQ established!")
	defer mq.Conn.Close()
	defer mq.Channel.Close()

	jobHandler := NewJobHandler(repo, mq)

	http.HandleFunc("/api/v1/process", jobHandler.CreateJobHandler)
	http.HandleFunc("/api/v1/jobs", jobHandler.GetJobStatusHandler)

	log.Println("API Server runs on http://localhost:8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("error starting server: %v", err)
	}
}
