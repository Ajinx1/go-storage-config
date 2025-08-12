package rabbitmq

import (
	"context"
	"encoding/json"
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

func (c *Client) ConsumeWithMiddleware(queue string, handler func(context.Context, interface{}) error, target interface{}, middlewares ...Middleware) error {
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

	msgs, err := c.conn.Channel.Consume(
		queue,
		"",
		false, // Manual ack
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			ctx := context.Background()
			var data interface{}
			if err := json.Unmarshal(msg.Body, &data); err != nil {
				msg.Nack(false, false)
				continue
			}

			for _, mw := range middlewares {
				if err := mw(ctx, queue, msg.Body); err != nil {
					msg.Nack(false, true)
					continue
				}
			}

			if err := handler(ctx, data); err != nil {
				msg.Nack(false, true)
				continue
			}

			msg.Ack(false)
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

func JSONToStruct(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func StructToJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
