package rabbitmq

import (
	"context"
	"log"

	"github.com/rabbitmq/amqp091-go"
)

func publishToDLQ(ctx context.Context, ch *amqp091.Channel, queue string, msg amqp091.Delivery) {
	dlqName := queue + ".dlq"

	err := ch.PublishWithContext(ctx,
		"",
		dlqName,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        msg.Body,
		},
	)

	if err != nil {
		log.Printf("[Worker] manual DLQ publish failed: %v", err)
	}
}

func alreadyRetried(msg amqp091.Delivery) bool {
	if msg.Headers == nil {
		return false
	}

	if _, ok := msg.Headers["x-death"]; ok {
		return true
	}

	return false
}
