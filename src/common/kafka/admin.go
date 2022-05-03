/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package kafka

import (
	"context"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/config"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"strings"
	"time"
)

type confluentKafkaAdminClient struct {
	KafkaAdmin
}

type KafkaAdmin interface {
	GetMetadata(topic *string, allTopics bool, timeoutMs int) (*kafka.Metadata, error)
	CreateTopics(ctx context.Context, topics []kafka.TopicSpecification, options ...kafka.CreateTopicsAdminOption) (result []kafka.TopicResult, err error)
	DeleteTopics(ctx context.Context, topics []string, options ...kafka.DeleteTopicsAdminOption) (result []kafka.TopicResult, err error)
}

func NewAdminClientFromConfig(config config.Config, bearerToken string) (KafkaAdmin, error) {
	kafkaConfig := &kafka.ConfigMap{"bootstrap.servers": strings.Join(config.KafkaBrokers, ",")}
	for key, value := range config.KafkaProperties {
		if !strings.HasPrefix(key, "sasl.") {
			kafkaConfig.SetKey(key, value)
		}
	}
	kafkaConfig.SetKey("security.protocol", "SASL_SSL")
	kafkaConfig.SetKey("sasl.mechanism", "OAUTHBEARER")

	admin, err := kafka.NewAdminClient(kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("error constructing Kafka admin client: %w", err)
	}

	err = admin.SetOAuthBearerToken(kafka.OAuthBearerToken{
		TokenValue: bearerToken,
		Expiration: time.Now().Add(time.Minute),
	})

	if err != nil {
		return nil, fmt.Errorf("error setting oauth bearer token: %w", err)
	}

	return confluentKafkaAdminClient{
		admin,
	}, nil
}
