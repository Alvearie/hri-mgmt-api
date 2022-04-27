/**
 * (C) Copyright IBM Corp. 2022
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package kafka

import (
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/config"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"strings"
	"time"
)

type Reader interface {
	Read(topic string, key string, val map[string]interface{}) error
	Close() error
}

// internal type that meets the Writer interface
type confluentKafkaReader struct {
	confluentConsumer
}

// internal interface for unit testing
type confluentConsumer interface {
	Subscribe(string, kafka.RebalanceCb) error
	ReadMessage(time.Duration) (*kafka.Message, error)
	Close() error
}

func NewReaderFromConfig(config config.Config) (Reader, error) {
	kafkaConfig := &kafka.ConfigMap{"bootstrap.servers": strings.Join(config.KafkaBrokers, ",")}
	for key, value := range config.KafkaProperties {
		kafkaConfig.SetKey(key, value)
	}

	consumer, err := kafka.NewConsumer(kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("error constructing Kafka producer: %w", err)
	}

	return confluentKafkaReader{
		consumer,
	}, nil
}

// This method is not thread safe. Each thread needs it's own ConfluentKafkaWriter instance
func (cfk confluentKafkaReader) Read(topic string, key string, val map[string]interface{}) error {

	return nil
}
