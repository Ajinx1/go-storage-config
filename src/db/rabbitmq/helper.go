package rabbitmq

import (
	"context"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

func (c *Client) getQueueArgs(queue string) amqp091.Table {
	if c.config.DeadLetterExchange == "" {
		return nil
	}

	return amqp091.Table{
		"x-dead-letter-exchange":    c.config.DeadLetterExchange,
		"x-dead-letter-routing-key": queue + ".dlq",
	}
}

func (c *Client) publishToDLQ(ctx context.Context, queue string, msg amqp091.Delivery) {
	dlqName := queue + ".dlq"

	ch, err := c.conn.Connection.Channel()
	if err != nil {
		log.Printf("[Worker] DLQ channel open failed: %v", err)
		return
	}
	defer ch.Close()

	err = ch.PublishWithContext(
		ctx,
		"",
		dlqName,
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         msg.Body,
			Timestamp:    time.Now(),
			DeliveryMode: amqp091.Persistent,
			Headers:      msg.Headers,
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

func (c *Client) declareQueueSafe(ch *amqp091.Channel, queue string) error {

	_, err := ch.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		c.getQueueArgs(queue),
	)

	if err == nil {
		return nil
	}

	log.Printf("[RabbitMQ] DLQ declare failed for %s, retrying without args: %v", queue, err)

	_, err = ch.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		nil,
	)

	return err
}
