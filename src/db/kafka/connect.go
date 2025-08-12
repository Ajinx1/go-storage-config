package kafka

import (
	"fmt"
	"time"

	"github.com/IBM/sarama"
)

type KafkaConn struct {
	Producer     sarama.SyncProducer
	Consumer     sarama.ConsumerGroup
	Admin        sarama.ClusterAdmin
	ConsumerDone chan struct{}
}

func ConnectFromEnv(cfg KafkaConfig) (*KafkaConn, error) {
	config := LoadKafkaConfig(cfg)

	// Sarama configuration
	saramaConfig := sarama.NewConfig()
	saramaConfig.ClientID = config.ClientID
	saramaConfig.Producer.RequiredAcks = sarama.RequiredAcks(config.RequiredAcks)
	saramaConfig.Producer.Retry.Max = config.MaxRetries
	saramaConfig.Producer.Retry.Backoff = time.Duration(config.RetryDelaySeconds) * time.Second
	saramaConfig.Producer.Idempotent = config.Idempotent
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Net.MaxOpenRequests = 5 // Required for idempotence
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	saramaConfig.Consumer.Offsets.Initial = parseOffset(config.AutoOffsetReset)
	saramaConfig.Consumer.Offsets.AutoCommit.Enable = true
	saramaConfig.Consumer.Offsets.AutoCommit.Interval = 5 * time.Second
	saramaConfig.Consumer.Group.Session.Timeout = 6 * time.Second
	saramaConfig.Consumer.Group.Heartbeat.Interval = 2 * time.Second
	saramaConfig.Consumer.MaxProcessingTime = 300 * time.Millisecond

	if config.Idempotent {
		saramaConfig.Producer.Transaction.ID = config.ClientID + "-txn"
	}

	// Create producer
	producer, err := sarama.NewSyncProducer([]string{config.BootstrapServers}, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	// Create consumer group
	consumerGroup, err := sarama.NewConsumerGroup([]string{config.BootstrapServers}, config.GroupID, saramaConfig)
	if err != nil {
		producer.Close()
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	// Create admin client
	admin, err := sarama.NewClusterAdmin([]string{config.BootstrapServers}, saramaConfig)
	if err != nil {
		producer.Close()
		consumerGroup.Close()
		return nil, fmt.Errorf("failed to create admin client: %w", err)
	}

	return &KafkaConn{
		Producer:     producer,
		Consumer:     consumerGroup,
		Admin:        admin,
		ConsumerDone: make(chan struct{}),
	}, nil
}

func (c *KafkaConn) Close() error {
	var errs []error

	if err := c.Consumer.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close consumer group: %w", err))
	}
	close(c.ConsumerDone)

	if err := c.Producer.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close producer: %w", err))
	}

	if err := c.Admin.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close admin client: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}
	return nil
}

func parseOffset(offset string) int64 {
	switch offset {
	case "earliest":
		return sarama.OffsetOldest
	case "latest":
		return sarama.OffsetNewest
	default:
		return sarama.OffsetOldest
	}
}
