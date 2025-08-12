package db

import (
	"github.com/Ajinx1/go-storage-config/src/db/kafka"
)

func InitKafka(theConfig kafka.KafkaConfig) (*kafka.Client, error) {
	return kafka.NewClientWithConfig(theConfig)
}
