package main

import (
	"context"
	"encoding/json"
	"imgpipe/pkg/db"
	"imgpipe/pkg/models"
	"imgpipe/pkg/processor"
	"imgpipe/pkg/queue"
	"log"
	"time"
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

	deliveries, err := mq.ConsumeJobs()
	if err != nil {
		log.Fatalf("failed to consume jobs: %v", err)
	}

	log.Println("[Worker] Worker service running... Waiting for jobs from RabbitMQ.")

	for msg := range deliveries {
		var job models.Job

		if err := json.Unmarshal(msg.Body, &job); err != nil {
			log.Printf("[Worker] Corrupted message JSON: %v. Dropping message.", err)
			msg.Nack(false, false)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		outputURL, err := processor.ProcessImage(ctx, &job)
		if err != nil {
			log.Printf("[Worker] Failed to process job %s: %v", job.ID, err)

			_ = repo.UpdateJobStatus(ctx, job.ID, "Failed", "", err.Error())

			msg.Nack(false, false)
			cancel()
			continue
		}

		err = repo.UpdateJobStatus(ctx, job.ID, "Completed", outputURL, "")
		if err != nil {
			log.Printf("[Worker] Failed to update DB status for job %s: %v", job.ID, err)

			msg.Nack(false, true)
			cancel()
			continue
		}

		msg.Ack(false)
		log.Printf("[Worker] Job %s fully processed and ACKed.", job.ID)
		cancel()
	}
}
