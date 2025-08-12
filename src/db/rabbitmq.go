package db

import (
	"github.com/Ajinx1/go-storage-config/src/db/rabbitmq"
)

func InitRabbitMQ(theConfig rabbitmq.RabbitMQConfig) (*rabbitmq.Client, error) {
	return rabbitmq.NewClientWithConfig(theConfig)
}
