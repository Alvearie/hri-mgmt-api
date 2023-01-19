package kafka

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Alvearie/hri-mgmt-api/common/config"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

const (
	AdminTimeout = time.Second * 45
)

type KafkaAdmin interface {
	GetMetadata(topic *string, allTopics bool, timeoutMs int) (*kafka.Metadata, error)
	CreateTopics(ctx context.Context, topics []kafka.TopicSpecification, options ...kafka.CreateTopicsAdminOption) (result []kafka.TopicResult, err error)
	DeleteTopics(ctx context.Context, topics []string, options ...kafka.DeleteTopicsAdminOption) (result []kafka.TopicResult, err error)
}

func AzNewAdminClientFromConfig(config config.Config, bearerToken string) (KafkaAdmin, *response.ErrorDetailResponse) {

	kafkaConfig := &kafka.ConfigMap{"bootstrap.servers": strings.Join(config.AzKafkaBrokers, ",")}
	for key, value := range config.AzKafkaProperties {
		kafkaConfig.SetKey(key, value)
	}
	//Add CA location
	kafkaConfig.SetKey("ssl.ca.location", config.SslCALocation)

	admin, err := kafka.NewAdminClient(kafkaConfig)
	if err != nil {
		return nil, response.NewErrorDetailResponse(http.StatusInternalServerError, "n/a", fmt.Sprintf("error constructing Kafka admin client: %s", err.Error()))
	}
	return admin, nil
}
