package main

import (
	"imgpipe/pkg/db"
	"imgpipe/pkg/queue"
	"log"
	"net/http"
	"os"
)

func main() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017" // Fallback for local `go run` execution
	}

	rabbitURI := os.Getenv("RABBITMQ_URI")
	if rabbitURI == "" {
		rabbitURI = "amqp://guest:guest@localhost:5672/" // Fallback for local `go run` execution
	}

	database, err := db.ConnectMongo(mongoURI, "imgpipe")
	if err != nil {
		log.Fatalf("error connecting to MongoDB: %v", err)
	}

	repo := db.NewJobRepository(database)
	log.Println("Connection with MongoDB established!")

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
