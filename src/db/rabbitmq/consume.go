package rabbitmq

import (
	"context"
	"encoding/json"
	"log"
	"reflect"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

func (c *Client) ConsumeWithMiddleware(ctx context.Context, queue string,
	handler func(context.Context, interface{}) error, target interface{},
	middlewares ...Middleware) error {

	go func() {
		backoff := time.Duration(c.config.RetryDelaySeconds) * time.Second
		maxBackoff := 60 * time.Second

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			// Reconnect if connection closed
			if c.conn.Connection.IsClosed() {
				log.Printf("[Worker] RabbitMQ connection closed, reconnecting...")
				newConn, err := c.Reconnect()
				if err != nil {
					log.Printf("[Worker] Failed to reconnect: %v. Retrying in %v", err, backoff)
					time.Sleep(backoff)
					if backoff < maxBackoff {
						backoff *= 2
					}
					continue
				}
				c.conn = newConn
				backoff = time.Duration(c.config.RetryDelaySeconds) * time.Second
			}

			// Declare queue
			_, err := c.conn.Channel.QueueDeclare(
				queue,
				true,
				false,
				false,
				false,
				amqp091.Table{"x-dead-letter-exchange": c.config.DeadLetterExchange},
			)
			if err != nil {
				log.Printf("[Worker] Failed to declare queue %s: %v. Reconnecting...", queue, err)
				c.conn.Connection.Close()
				continue
			}

			// Start consuming
			msgs, err := c.conn.Channel.Consume(queue, "", false, false, false, false, nil)
			if err != nil {
				log.Printf("[Worker] Failed to consume queue %s: %v. Reconnecting...", queue, err)
				c.conn.Connection.Close()
				continue
			}

			log.Printf("[Worker] Started consumer for queue: %s", queue)

			for msg := range msgs {
				data := reflect.New(reflect.TypeOf(target).Elem()).Interface()
				if err := json.Unmarshal(msg.Body, data); err != nil {
					log.Printf("[Worker] JSON unmarshal failed: %v", err)
					msg.Nack(false, false)
					continue
				}

				// Middleware
				for _, mw := range middlewares {
					if err := mw(ctx, queue, msg.Body); err != nil {
						log.Printf("[Worker] Middleware failed: %v", err)
						msg.Nack(false, true)
						continue
					}
				}

				// Handler retry
				const maxHandlerRetries = 3
				var handlerErr error
				for attempt := 1; attempt <= maxHandlerRetries; attempt++ {
					handlerErr = handler(ctx, data)
					if handlerErr == nil {
						break
					}
					log.Printf("[Worker] Handler failed (attempt %d/%d): %v", attempt, maxHandlerRetries, handlerErr)
					time.Sleep(time.Second * time.Duration(attempt))
				}

				if handlerErr != nil {
					msg.Nack(false, true)
					continue
				}

				if err := msg.Ack(false); err != nil {
					log.Printf("[Worker] Failed to ack message: %v", err)
				} else {
					log.Printf("[Worker] Message processed and acked")
				}
			}

			log.Printf("[Worker] Channel closed for queue %s, reconnecting in %v...", queue, backoff)
			c.conn.Connection.Close()
			time.Sleep(backoff)
			if backoff < maxBackoff {
				backoff *= 2
			}
		}
	}()

	return nil
}
