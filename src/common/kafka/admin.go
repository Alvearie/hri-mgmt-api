package kafka

import (
	"context"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/config"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"strings"
)

type confluentKafkaAdminClient struct {
	KafkaAdmin
}

type KafkaAdmin interface {
	GetMetadata(topic *string, allTopics bool, timeoutMs int) (*kafka.Metadata, error)
	CreateTopics(ctx context.Context, topics []kafka.TopicSpecification, options ...kafka.CreateTopicsAdminOption) (result []kafka.TopicResult, err error)
	DeleteTopics(ctx context.Context, topics []string, options ...kafka.DeleteTopicsAdminOption) (result []kafka.TopicResult, err error)
}

func NewAdminClientFromConfig(config config.Config) (KafkaAdmin, error) {
	kafkaConfig := &kafka.ConfigMap{"bootstrap.servers": strings.Join(config.KafkaBrokers, ",")}
	for key, value := range config.KafkaProperties {
		kafkaConfig.SetKey(key, value)
	}
	admin, err := kafka.NewAdminClient(kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("error constructing Kafka admin client: %w", err)
	}

	return confluentKafkaAdminClient{
		admin,
	}, nil
}