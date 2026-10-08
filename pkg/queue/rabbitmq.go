package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"imgpipe/pkg/models"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
	Queue   amqp.Queue
}

func NewRabbitMQ(amqpURL string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("can't connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("can't open amqp channel: %w", err)
	}

	q, err := ch.QueueDeclare(
		"image_jobs",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("can't declare queue: %w", err)
	}

	log.Println("API is connected to RabbitMQ!")

	return &RabbitMQ{
		Conn:    conn,
		Channel: ch,
		Queue:   q,
	}, nil
}

func (r *RabbitMQ) PublishJob(ctx context.Context, job *models.Job) error {
	body, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("can't serialize job to JSON: %w", err)
	}

	err = r.Channel.PublishWithContext(
		ctx,
		"",
		r.Queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("can't publish message to RabbitMQ: %w", err)
	}

	log.Printf("Job %s successfully sent to RabbitMQ!", job.ID)
	return nil
}
