/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package kafka

import (
	"encoding/json"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/config"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"strings"
)

type Writer interface {
	Write(topic string, key string, val map[string]interface{}) error
	Close()
}

// internal type that meets the Writer interface
type confluentKafkaWriter struct {
	confluentProducer
}

// internal interface for unit testing
type confluentProducer interface {
	Produce(*kafka.Message, chan kafka.Event) error
	Events() chan kafka.Event
	Flush(int) int
	Close()
}

func NewWriterFromConfig(config config.Config) (Writer, error) {
	kafkaConfig := &kafka.ConfigMap{"bootstrap.servers": strings.Join(config.KafkaBrokers, ",")}
	for key, value := range config.KafkaProperties {
		kafkaConfig.SetKey(key, value)
	}

	producer, err := kafka.NewProducer(kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("error constructing Kafka producer: %w", err)
	}

	return confluentKafkaWriter{
		producer,
	}, nil
}

// This method is not thread safe. Each thread needs it's own ConfluentKafkaWriter instance
func (cfk confluentKafkaWriter) Write(topic string, key string, val map[string]interface{}) error {
	jsonVal, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("error marshaling kafka message: %w", err)
	}

	err = cfk.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          jsonVal,
	}, nil) // nil uses the default producer channel

	if err != nil {
		return fmt.Errorf("kafka producer error: %w", err)
	}

	// wait upto 1 second for messages to be written
	cfk.Flush(1000)

	// The Confluent Kafka library sends the messages in a separate thread and
	// a channel is used to communicate a result back. This call blocks until
	// that threads sends the result back over the channel, and let's us verify
	// there weren't any issues. This pattern was taken from the library's examples:
	// https://github.com/confluentinc/confluent-kafka-go/blob/master/examples/producer_example/producer_example.go
	e := <-cfk.Events()
	m := e.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		return fmt.Errorf("kafka producer error: %w", m.TopicPartition.Error)
	}

	return nil
}
