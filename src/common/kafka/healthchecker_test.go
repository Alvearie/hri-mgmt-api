/**
 * (C) Copyright IBM Corp. 2021
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package kafka

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Alvearie/hri-mgmt-api/common/config"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/assert"
)

func TestNewHealthChecker(t *testing.T) {
	tests := []struct {
		name   string
		config config.Config
		expErr error
	}{
		{
			name:   "successful construction",
			config: config.Config{AzKafkaBrokers: []string{"broker1", "broker2"}, AzKafkaProperties: config.StringMap{"message.max.bytes": "10000"}},
			expErr: nil,
		},
		{
			name:   "bad config",
			config: config.Config{AzKafkaBrokers: []string{"broker1", "broker2"}, AzKafkaProperties: config.StringMap{"message.max.bytes": "bad_value"}},
			expErr: fmt.Errorf("error constructing Kafka admin client: %w",
				kafka.NewError(-186, "Invalid value for configuration property \"message.max.bytes\"", false)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := HriHealthChecker(tt.config)

			assert.Equal(t, tt.expErr, err)
		})
	}
}

type fakeAdminClient struct {
	t        *testing.T
	metadata *kafka.Metadata
	err      error
}

func (fac fakeAdminClient) GetMetadata(topics *string, allTopics bool, timeout int) (*kafka.Metadata, error) {
	assert.Nil(fac.t, topics)
	assert.True(fac.t, allTopics)
	assert.Equal(fac.t, 1000, timeout)

	return fac.metadata, fac.err
}

func (fac fakeAdminClient) Close() {}

func TestConfluentHealthChecker_Check(t *testing.T) {
	tests := []struct {
		name   string
		client fakeAdminClient
		expErr error
	}{
		{
			name: "successful health check",
			client: fakeAdminClient{
				t: t,
				metadata: &kafka.Metadata{
					Brokers:           []kafka.BrokerMetadata{{1, "broker-host", 9093}},
					Topics:            map[string]kafka.TopicMetadata{"topic": {Topic: "topic"}},
					OriginatingBroker: kafka.BrokerMetadata{},
				},
				err: nil,
			},
			expErr: nil,
		},
		{
			name: "error on health check",
			client: fakeAdminClient{
				t:        t,
				metadata: nil,
				err:      errors.New("connection timeout"),
			},
			expErr: fmt.Errorf("error getting Kafka topics: %w", errors.New("connection timeout")),
		},
		{
			name: "health check response missing metadata",
			client: fakeAdminClient{
				t:        t,
				metadata: nil,
				err:      nil,
			},
			expErr: errors.New("error getting Kafka topics; returned metadata or list of brokers was empty"),
		},
		{
			name: "health check response missing brokers",
			client: fakeAdminClient{
				t: t,
				metadata: &kafka.Metadata{
					Brokers:           []kafka.BrokerMetadata{},
					Topics:            map[string]kafka.TopicMetadata{"topic": {Topic: "topic"}},
					OriginatingBroker: kafka.BrokerMetadata{},
				},
				err: nil,
			},
			expErr: errors.New("error getting Kafka topics; returned metadata or list of brokers was empty"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			healthChecker := confluentHealthChecker{tt.client}
			err := healthChecker.Check()

			assert.Equal(t, tt.expErr, err)
		})
	}
}
