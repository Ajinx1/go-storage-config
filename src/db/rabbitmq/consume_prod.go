package rabbitmq

import (
	"context"
	"encoding/json"
	"log"
	"reflect"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

func (c *Client) ConsumeWithMiddleware(
	ctx context.Context,
	queue string,
	handler func(context.Context, interface{}) error,
	target interface{},
	middlewares ...Middleware,
) error {

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
			if c.conn.Connection == nil || c.conn.Connection.IsClosed() {
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

			ch, err := c.conn.Connection.Channel()
			if err != nil {
				log.Printf("[Worker] Failed to open channel: %v", err)
				time.Sleep(backoff)
				continue
			}
			c.conn.Channel = ch

			// 🔥 QoS (controls concurrency per consumer)
			prefetch := 5
			if err := ch.Qos(prefetch, 0, false); err != nil {
				log.Printf("[Worker] Failed to set QoS: %v", err)
				ch.Close()
				continue
			}

			// Declare queue
			_, err = ch.QueueDeclare(
				queue,
				true,
				false,
				false,
				false,
				amqp091.Table{
					"x-dead-letter-exchange":    c.config.DeadLetterExchange,
					"x-dead-letter-routing-key": queue + ".dlq",
				},
			)
			if err != nil {
				log.Printf("[Worker] Failed to declare queue %s: %v", queue, err)
				ch.Close()
				continue
			}

			// Start consuming
			msgs, err := ch.Consume(
				queue,
				"",
				false, // manual ack
				false,
				false,
				false,
				nil,
			)
			if err != nil {
				log.Printf("[Worker] Failed to consume queue %s: %v", queue, err)
				ch.Close()
				continue
			}

			log.Printf("[Worker] Started consumer for queue: %s (prefetch=%d)", queue, prefetch)

			// 🔥 Worker pool (limits goroutines to prefetch)
			sem := make(chan struct{}, prefetch)

			for msg := range msgs {
				sem <- struct{}{}

				go func(msg amqp091.Delivery) {
					defer func() { <-sem }()

					// Deserialize
					data := reflect.New(reflect.TypeOf(target).Elem()).Interface()
					if err := json.Unmarshal(msg.Body, data); err != nil {
						log.Printf("[Worker] JSON unmarshal failed: %v", err)
						msg.Nack(false, false) // send to DLQ
						return
					}

					// Middleware execution
					for _, mw := range middlewares {
						if err := mw(ctx, queue, msg.Body); err != nil {
							log.Printf("[Worker] Middleware failed: %v", err)
							msg.Nack(false, true) // requeue
							return
						}
					}

					// Handler retry logic
					const maxHandlerRetries = 3
					var handlerErr error

					for attempt := 1; attempt <= maxHandlerRetries; attempt++ {
						handlerErr = handler(ctx, data)
						if handlerErr == nil {
							break
						}

						log.Printf("[Worker] Handler failed (attempt %d/%d): %v",
							attempt, maxHandlerRetries, handlerErr)

						time.Sleep(time.Duration(attempt) * time.Second)
					}

					if handlerErr != nil {
						log.Printf("[Worker] Sending message to DLQ after retries")
						msg.Nack(false, false)
						return
					}

					// ACK
					if err := msg.Ack(false); err != nil {
						log.Printf("[Worker] Failed to ack message: %v", err)
					} else {
						log.Printf("[Worker] Message processed and acked")
					}

				}(msg)
			}

			log.Printf("[Worker] Channel closed for queue %s, reconnecting in %v...", queue, backoff)
			ch.Close()
			time.Sleep(backoff)

			if backoff < maxBackoff {
				backoff *= 2
			}
		}
	}()

	return nil
}
