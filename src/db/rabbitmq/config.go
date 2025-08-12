package rabbitmq

type RabbitMQConfig struct {
	URL                string
	MaxRetries         int    // Maximum retry attempts for operations
	RetryDelaySeconds  int    // Delay between retries in seconds
	DeadLetterExchange string // Dead letter exchange name
	DeadLetterQueue    string // Dead letter queue name
}

func DefaultConfig() *RabbitMQConfig {
	return &RabbitMQConfig{
		MaxRetries:         5,
		RetryDelaySeconds:  5,
		DeadLetterExchange: "dlx.exchange",
		DeadLetterQueue:    "dlx.queue",
	}
}

func LoadRabbitMQConfig(cfg RabbitMQConfig) *RabbitMQConfig {
	defaultCfg := DefaultConfig()
	if cfg.URL != "" {
		defaultCfg.URL = cfg.URL
	}
	if cfg.MaxRetries > 0 {
		defaultCfg.MaxRetries = cfg.MaxRetries
	}
	if cfg.RetryDelaySeconds > 0 {
		defaultCfg.RetryDelaySeconds = cfg.RetryDelaySeconds
	}
	if cfg.DeadLetterExchange != "" {
		defaultCfg.DeadLetterExchange = cfg.DeadLetterExchange
	}
	if cfg.DeadLetterQueue != "" {
		defaultCfg.DeadLetterQueue = cfg.DeadLetterQueue
	}
	return defaultCfg
}
