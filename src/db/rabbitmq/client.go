package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn   *RabbitConn
	config *RabbitMQConfig
}

type Middleware func(context.Context, string, []byte) error

func NewClientWithConfig(cfg RabbitMQConfig) (*Client, error) {
	conn, err := ConnectFromEnv(cfg)
	if err != nil {
		return nil, err
	}

	client := &Client{
		conn:   conn,
		config: LoadRabbitMQConfig(cfg),
	}

	if err := client.setupDLX(); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) setupDLX() error {
	err := c.conn.Channel.ExchangeDeclare(
		c.config.DeadLetterExchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	_, err = c.conn.Channel.QueueDeclare(
		c.config.DeadLetterQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	return c.conn.Channel.QueueBind(
		c.config.DeadLetterQueue,
		"",
		c.config.DeadLetterExchange,
		false,
		nil,
	)
}

func (c *Client) PublishWithMiddleware(ctx context.Context, queue string, body interface{}, middlewares ...Middleware) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	for _, mw := range middlewares {
		if err := mw(ctx, queue, data); err != nil {
			return err
		}
	}

	return c.retryOperation(ctx, func() error {
		return c.publish(queue, data)
	})
}

func (c *Client) publish(queue string, body []byte) error {
	_, err := c.conn.Channel.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		amqp091.Table{
			"x-dead-letter-exchange": c.config.DeadLetterExchange,
		},
	)
	if err != nil {
		return err
	}

	return c.conn.Channel.PublishWithContext(
		context.Background(),
		"",
		queue,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now(),
		},
	)
}

func (c *Client) ConsumeWithMiddleware(ctx context.Context, queue string, handler func(context.Context, interface{}) error, target interface{},
	middlewares ...Middleware) error {
	declareQueue := func() error {
		_, err := c.conn.Channel.QueueDeclare(
			queue,
			true,
			false,
			false,
			false,
			amqp091.Table{
				"x-dead-letter-exchange": c.config.DeadLetterExchange,
			},
		)
		return err
	}

	if err := declareQueue(); err != nil {
		return fmt.Errorf("failed to declare queue %s: %w", queue, err)
	}

	// Retry loop for consumer startup
	var msgs <-chan amqp091.Delivery
	var err error
	for retries := 0; retries < 5; retries++ {
		msgs, err = c.conn.Channel.Consume(
			queue,
			"",
			false, // Manual ack
			false,
			false,
			false,
			nil,
		)
		if err == nil {
			break
		}
		log.Printf("[Worker] Failed to consume queue %s: %v (retrying in %ds)", queue, err, retries+1)
		time.Sleep(time.Duration(retries+1) * time.Second)
	}
	if err != nil {
		return fmt.Errorf("failed to consume queue %s after retries: %w", queue, err)
	}

	go func() {
		log.Printf("[Worker] Started consumer for queue: %s", queue)

		for {
			select {
			case <-ctx.Done():
				log.Printf("[Worker] Stopping consumer for queue: %s", queue)
				return
			case msg, ok := <-msgs:
				if !ok {
					log.Printf("[Worker] Channel closed for queue: %s", queue)
					return
				}

				log.Printf("[Worker] Received message from %s", queue)

				// Create a new instance of the target type
				data := reflect.New(reflect.TypeOf(target).Elem()).Interface()
				if err := json.Unmarshal(msg.Body, data); err != nil {
					log.Printf("[Worker] JSON unmarshal failed: %v", err)
					msg.Nack(false, false) // send to DLX
					continue
				}

				// Middleware execution
				for _, mw := range middlewares {
					if err := mw(ctx, queue, msg.Body); err != nil {
						log.Printf("[Worker] Middleware failed: %v", err)
						msg.Nack(false, true) // requeue
						continue
					}
				}

				// Handler execution with retry
				const maxHandlerRetries = 3
				var handlerErr error
				for attempt := 1; attempt <= maxHandlerRetries; attempt++ {
					handlerErr = handler(ctx, data)
					if handlerErr == nil {
						break
					}
					log.Printf("[Worker] Handler failed (attempt %d/%d): %v", attempt, maxHandlerRetries, handlerErr)
					time.Sleep(time.Second * time.Duration(attempt)) // backoff
				}

				if handlerErr != nil {
					msg.Nack(false, true) // requeue
					continue
				}

				// Ack
				if err := msg.Ack(false); err != nil {
					log.Printf("[Worker] Failed to ack message: %v", err)
				} else {
					log.Printf("[Worker] Message processed and acked")
				}
			}
		}
	}()

	return nil
}
func (c *Client) retryOperation(ctx context.Context, operation func() error) error {
	var lastErr error
	for i := 0; i < c.config.MaxRetries; i++ {
		if err := operation(); err != nil {
			lastErr = err
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(c.config.RetryDelaySeconds) * time.Second):
				continue
			}
		}
		return nil
	}
	return lastErr
}

func (c *Client) Close() error {
	var errs []error

	if c.conn != nil && c.conn.Channel != nil {
		if err := c.conn.Channel.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close channel: %w", err))
		}
	}

	if c.conn != nil && c.conn.Connection != nil {
		if err := c.conn.Connection.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close connection: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors while closing RabbitMQ: %v", errs)
	}

	log.Println("[RabbitMQ] Connection and channel closed")
	return nil
}

func JSONToStruct(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func StructToJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
