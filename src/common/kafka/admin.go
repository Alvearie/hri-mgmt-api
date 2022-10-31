/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
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
	"github.com/golang-jwt/jwt"
)

const (
	AdminTimeout = time.Second * 45
)

type KafkaAdmin interface {
	GetMetadata(topic *string, allTopics bool, timeoutMs int) (*kafka.Metadata, error)
	CreateTopics(ctx context.Context, topics []kafka.TopicSpecification, options ...kafka.CreateTopicsAdminOption) (result []kafka.TopicResult, err error)
	DeleteTopics(ctx context.Context, topics []string, options ...kafka.DeleteTopicsAdminOption) (result []kafka.TopicResult, err error)
}

func NewAdminClientFromConfig(config config.Config, bearerToken string) (KafkaAdmin, *response.ErrorDetailResponse) {

	kafkaConfig := &kafka.ConfigMap{"bootstrap.servers": strings.Join(config.KafkaBrokers, ",")}
	// We use oauthbearer auth for create/delete topic requests.
	//kafkaConfig.SetKey("security.protocol", config.KafkaProperties["security.protocol"])
	//kafkaConfig.SetKey("sasl.mechanism", "OAUTHBEARER")
	//kafkaConfig.SetKey("ssl.ca.location", "C:/manav-kafka/pem/client.cer.pem")

	for key, value := range config.KafkaProperties {
		kafkaConfig.SetKey(key, value)
	}
	kafkaConfig.SetKey("ssl.ca.location", "C:/manav-kafka/pem/client.cer.pem")
	admin, err := kafka.NewAdminClient(kafkaConfig)
	if err != nil {
		return nil, response.NewErrorDetailResponse(http.StatusInternalServerError, "n/a", fmt.Sprintf("error constructing Kafka admin client: %s", err.Error()))
	}

	/*re := regexp.MustCompile(`(?i)bearer `) // Remove bearer prefix, ignoring case.
	tokenValue := re.ReplaceAllString(bearerToken, "")

	exp, err := getExpFromToken(tokenValue)
	if err != nil {
		return nil, response.NewErrorDetailResponse(http.StatusUnauthorized, "n/a", err.Error())
	}

	err = admin.SetOAuthBearerToken(kafka.OAuthBearerToken{
		TokenValue: tokenValue,
		Expiration: exp,
	})
	if err != nil {
		return nil, response.NewErrorDetailResponse(http.StatusUnauthorized, "n/a", err.Error())
	}*/

	return admin, nil
}

// To use OAuth bearer authentication in the confluent-kafka-go library, we need to create an OAuthBearerToken object.
// Creating this object requires providing the actual expiration ("exp") value encoded in the jwt token. We therefore
// must decode the token that was passed in and extract the "exp" field to create the confluent token object.
func getExpFromToken(bearerToken string) (time.Time, error) {
	exp := time.Unix(0, 0)
	token, _, err := new(jwt.Parser).ParseUnverified(bearerToken, jwt.MapClaims{})
	if err != nil {
		return exp, fmt.Errorf("unexpected error parsing bearer token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return exp, fmt.Errorf("unexpected error parsing bearer token: %w", err)
	}

	if val, ok := claims["exp"].(float64); ok {
		exp = time.Unix(int64(val), 0)
	} else {
		return exp, fmt.Errorf("unexpected error parsing bearer token, could not parse expiration time")
	}

	return exp, nil
}
