package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
)

type Client struct {
	conn   *KafkaConn
	config *KafkaConfig
}

type Middleware func(context.Context, string, []byte) error

func NewClientWithConfig(cfg KafkaConfig) (*Client, error) {
	conn, err := ConnectFromEnv(cfg)
	if err != nil {
		return nil, err
	}

	client := &Client{
		conn:   conn,
		config: LoadKafkaConfig(cfg),
	}

	if err := client.setupDLQ(); err != nil {
		conn.Close()
		return nil, err
	}

	return client, nil
}

func (c *Client) setupDLQ() error {
	topic := c.config.DeadLetterTopic
	err := c.conn.Admin.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     2,
		ReplicationFactor: 2,
	}, false)
	if err != nil {
		if err == sarama.ErrTopicAlreadyExists {
			return nil
		}
		return fmt.Errorf("failed to create DLQ topic %s: %w", topic, err)
	}
	return nil
}

// Publish sends a message to the specified topic with retry logic
func (c *Client) Publish(ctx context.Context, topic string, key, value []byte, middlewares ...Middleware) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
	}

	for _, mw := range middlewares {
		if err := mw(ctx, topic, value); err != nil {
			return fmt.Errorf("middleware failed: %w", err)
		}
	}

	var lastErr error
	for i := 0; i <= c.config.MaxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_, _, err := c.conn.Producer.SendMessage(msg)
			if err == nil {
				return nil
			}
			lastErr = err
			log.Printf("Failed to publish message, retry %d/%d: %v", i+1, c.config.MaxRetries, err)
			time.Sleep(time.Duration(c.config.RetryDelaySeconds) * time.Second)
		}
	}

	if dlqErr := c.sendToDLQ(ctx, topic, key, value); dlqErr != nil {
		return fmt.Errorf("max retries reached, last error: %v, DLQ error: %v", lastErr, dlqErr)
	}
	return fmt.Errorf("max retries reached, last error: %v", lastErr)

}

// sendToDLQ sends a failed message to the dead letter queue
func (c *Client) sendToDLQ(ctx context.Context, topic string, key, value []byte) error {
	dlqMsg := &sarama.ProducerMessage{
		Topic: c.config.DeadLetterTopic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
		Headers: []sarama.RecordHeader{
			{Key: []byte("original-topic"), Value: []byte(topic)},
			{Key: []byte("timestamp"), Value: []byte(time.Now().UTC().String())},
		},
	}

	_, _, err := c.conn.Producer.SendMessage(dlqMsg)
	if err != nil {
		return fmt.Errorf("failed to send message to DLQ: %w", err)
	}
	return nil
}

// ConsumerGroupHandler implements sarama.ConsumerGroupHandler for message processing
type ConsumerGroupHandler struct {
	handler     func(context.Context, *sarama.ConsumerMessage) error
	middlewares []Middleware
	topic       string
}

func (h *ConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := session.Context()
	for msg := range claim.Messages() {
		for _, mw := range h.middlewares {
			if err := mw(ctx, h.topic, msg.Value); err != nil {
				log.Printf("Middleware error for message: %v", err)
				return err
			}
		}

		if err := h.handler(ctx, msg); err != nil {
			log.Printf("Handler error for message: %v", err)
			return err
		}
		session.MarkMessage(msg, "")
	}
	return nil
}

// Subscribe consumes messages from the specified topics with middleware support
func (c *Client) Subscribe(ctx context.Context, topics []string, handler func(context.Context, *sarama.ConsumerMessage) error, middlewares ...Middleware) error {
	consumerHandler := &ConsumerGroupHandler{
		handler:     handler,
		middlewares: middlewares,
		topic:       topics[0],
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err := c.conn.Consumer.Consume(ctx, topics, consumerHandler)
			if err != nil {
				if err == sarama.ErrClosedConsumerGroup {
					return nil
				}
				log.Printf("Consumer error: %v", err)
				time.Sleep(time.Duration(c.config.RetryDelaySeconds) * time.Second)
				continue
			}
		}
	}
}

// Close gracefully shuts down the client
func (c *Client) Close() error {
	return c.conn.Close()
}
