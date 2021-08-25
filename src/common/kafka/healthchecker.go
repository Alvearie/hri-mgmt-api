/**
 * (C) Copyright IBM Corp. 2021
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package kafka

import (
	"errors"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"ibm.com/watson/health/foundation/hri/common/config"
	"strings"
)

// HealthChecker Public interface
type HealthChecker interface {
	Check() error
	Close()
}

// internal type that meets the HealthChecker interface
type confluentHealthChecker struct {
	confluentAdminClient
}

// internal interface for unit testing
type confluentAdminClient interface {
	GetMetadata(*string, bool, int) (*kafka.Metadata, error)
	Close()
}

func NewHealthChecker(config config.Config) (HealthChecker, error) {
	kafkaConfig := &kafka.ConfigMap{"bootstrap.servers": strings.Join(config.KafkaBrokers, ",")}
	for key, value := range config.KafkaProperties {
		kafkaConfig.SetKey(key, value)
	}

	client, err := kafka.NewAdminClient(kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("error constructing Kafka admin client: %w", err)
	}

	return confluentHealthChecker{client}, nil
}

func (chc confluentHealthChecker) Check() error {
	metadata, err := chc.GetMetadata(nil, true, 1000)
	if err != nil {
		return fmt.Errorf("error getting Kafka topics: %w", err)
	}

	if metadata == nil || len(metadata.Brokers) == 0 {
		return errors.New("error getting Kafka topics; returned metadata or list of brokers was empty")
	}

	// We can't assume that there are any topics, so if `GetMetadata()` succeeds without an error
	// then we can assume Kafka is up and we can connect
	return nil
}
