package rabbitmq

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

type RabbitConn struct {
	Connection *amqp091.Connection
	Channel    *amqp091.Channel
}

func ConnectFromEnv(cfg RabbitMQConfig) (*RabbitConn, error) {
	config := LoadRabbitMQConfig(cfg)

	conn, err := amqp091.Dial(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	if err := ch.Qos(1, 0, false); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	return &RabbitConn{
		Connection: conn,
		Channel:    ch,
	}, nil
}

func (c *RabbitConn) Close() error {
	if err := c.Channel.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %w", err)
	}
	if err := c.Connection.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}
	return nil
}
