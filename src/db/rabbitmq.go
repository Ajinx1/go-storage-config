package db

import (
	"go-storage-config/src/db/rabbitmq"
)

// var RabbitClient *rabbitmq.Client

// func InitRabbitMQ(theConfig rabbitmq.RabbitMQConfig) {
// 	client, err := rabbitmq.NewClientWithConfig(theConfig)
// 	if err != nil {
// 		log.Fatalf("RabbitMQ initialization failed: %v", err)
// 	}
// 	RabbitClient = client
// }

func InitRabbitMQ(theConfig rabbitmq.RabbitMQConfig) (*rabbitmq.Client, error) {
	return rabbitmq.NewClientWithConfig(theConfig)
}
