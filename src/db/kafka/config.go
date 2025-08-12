package kafka

type KafkaConfig struct {
	BootstrapServers  string
	ClientID          string
	GroupID           string
	MaxRetries        int
	RetryDelaySeconds int
	DeadLetterTopic   string
	RequiredAcks      int    // Maps to Sarama's RequiredAcks (e.g., WaitForAll, WaitForLocal)
	Idempotent        bool   // Enable idempotent producer
	AutoOffsetReset   string // Consumer offset reset policy (earliest, latest)
}

func DefaultConfig() *KafkaConfig {
	return &KafkaConfig{
		BootstrapServers:  "localhost:9092",
		ClientID:          "go-kafka-client",
		GroupID:           "go-kafka-group",
		MaxRetries:        5,
		RetryDelaySeconds: 5,
		DeadLetterTopic:   "dlq.topic",
		RequiredAcks:      -1,
		Idempotent:        true,
		AutoOffsetReset:   "earliest",
	}
}

func LoadKafkaConfig(cfg KafkaConfig) *KafkaConfig {
	defaultCfg := DefaultConfig()
	if cfg.BootstrapServers != "" {
		defaultCfg.BootstrapServers = cfg.BootstrapServers
	}
	if cfg.ClientID != "" {
		defaultCfg.ClientID = cfg.ClientID
	}
	if cfg.GroupID != "" {
		defaultCfg.GroupID = cfg.GroupID
	}
	if cfg.MaxRetries > 0 {
		defaultCfg.MaxRetries = cfg.MaxRetries
	}
	if cfg.RetryDelaySeconds > 0 {
		defaultCfg.RetryDelaySeconds = cfg.RetryDelaySeconds
	}
	if cfg.DeadLetterTopic != "" {
		defaultCfg.DeadLetterTopic = cfg.DeadLetterTopic
	}
	if cfg.RequiredAcks != 0 {
		defaultCfg.RequiredAcks = cfg.RequiredAcks
	}
	if cfg.Idempotent {
		defaultCfg.Idempotent = cfg.Idempotent
	}
	if cfg.AutoOffsetReset != "" {
		defaultCfg.AutoOffsetReset = cfg.AutoOffsetReset
	}
	return defaultCfg
}
