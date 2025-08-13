package rabbitmq

import (
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type RabbitConn struct {
	Connection *amqp091.Connection
	Channel    *amqp091.Channel
}

func ConnectFromEnv(cfg RabbitMQConfig) (*RabbitConn, error) {
	config := LoadRabbitMQConfig(cfg)

	conn, err := amqp091.DialConfig(config.URL, amqp091.Config{
		Heartbeat: 10 * time.Second,
	})
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

func (c *Client) Reconnect() (*RabbitConn, error) {
	cfg := LoadRabbitMQConfig(*c.config)

	var conn *amqp091.Connection
	var ch *amqp091.Channel
	var err error

	for attempt := 1; attempt <= cfg.MaxRetries; attempt++ {
		conn, err = amqp091.DialConfig(cfg.URL, amqp091.Config{Heartbeat: 10 * time.Second})
		if err == nil {
			ch, err = conn.Channel()
			if err == nil {
				if err := ch.Qos(1, 0, false); err != nil {
					conn.Close()
					return nil, err
				}
				return &RabbitConn{Connection: conn, Channel: ch}, nil
			}
			conn.Close()
		}
		log.Printf("[RabbitMQ] Reconnect attempt %d/%d failed: %v. Retrying in %ds", attempt, cfg.MaxRetries, err, cfg.RetryDelaySeconds)
		time.Sleep(time.Duration(cfg.RetryDelaySeconds) * time.Second)
	}
	return nil, fmt.Errorf("failed to reconnect after %d attempts: %w", cfg.MaxRetries, err)
}
