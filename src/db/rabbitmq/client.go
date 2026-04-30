package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

	return client, nil
}

func (c *Client) setupDLXForQueue(queue string) error {

	ch, err := c.conn.Connection.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	dlqName := queue + ".dlq"

	if err := ch.ExchangeDeclare(
		c.config.DeadLetterExchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	_, err = ch.QueueDeclare(
		dlqName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	return ch.QueueBind(
		dlqName,
		dlqName,
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

	ch, err := c.conn.Connection.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := c.declareQueueSafe(ch, queue); err != nil {
		return err
	}

	return ch.PublishWithContext(
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
