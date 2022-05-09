/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package kafka

import (
	"context"
	"errors"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/config"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/golang-jwt/jwt"
	"net/http"
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

func NewAdminClientFromConfig(config config.Config, bearerToken string) (adminClient KafkaAdmin, errCode int, err error) {

	kafkaConfig := &kafka.ConfigMap{"bootstrap.servers": strings.Join(config.KafkaBrokers, ",")}
	// We use oauthbearer auth for create/delete topic requests.
	kafkaConfig.SetKey("security.protocol", "SASL_SSL")
	kafkaConfig.SetKey("sasl.mechanism", "OAUTHBEARER")

	admin, err := kafka.NewAdminClient(kafkaConfig)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("error constructing Kafka admin client: %w", err)
	}

	exp, err := getExpFromToken(bearerToken)
	if err != nil {
		return nil, http.StatusUnauthorized, err
	}

	err = admin.SetOAuthBearerToken(kafka.OAuthBearerToken{
		TokenValue: bearerToken,
		Expiration: exp,
	})
	if err != nil {
		return nil, http.StatusUnauthorized, err
	}

	return confluentKafkaAdminClient{
		admin,
	}, -1, nil
}

func getExpFromToken(bearerToken string) (time.Time, error) {
	exp := time.Unix(0, 0)
	token, _, err := new(jwt.Parser).ParseUnverified(bearerToken, jwt.MapClaims{})
	if err != nil {
		return exp, errors.New("unexpected error parsing bearer token, could not parse jwt token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return exp, errors.New("unexpected error parsing bearer token, could not extract claims")
	}

	if val, ok := claims["exp"].(float64); ok {
		exp = time.Unix(int64(val), 0)
	} else if val, ok := claims["exp"].(int64); ok {
		exp = time.Unix(val, 0)
	} else {
		return exp, errors.New("unexpected error parsing bearer token, field 'exp' was not of expected type")
	}

	return exp, nil
}
