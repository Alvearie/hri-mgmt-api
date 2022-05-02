/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package streams

import (
	"context"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/kafka"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	cfk "github.com/confluentinc/confluent-kafka-go/kafka"
	"net/http"
	"strconv"
	"time"
)

const (
	onePartition      int = 1
	replicationFactor int = 2
)

func Create(
	request model.CreateStreamsRequest,
	tenantId string,
	streamId string,
	validationEnabled bool,
	requestId string,
	adminClient kafka.KafkaAdmin) ([]string, int, error) {

	prefix := "streams/create"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Streams Create")

	inTopicName, notificationTopicName, outTopicName, invalidTopicName := kafka.CreateTopicNames(
		tenantId, streamId)

	//numPartitions and retentionTime configs are required, the rest are optional
	topicConfig := setUpTopicConfig(request)

	// Build up topic names and partition counts arrays. Orders are:
	// { inTopicName, notificationTopicName, outTopicName, invalidTopicName }
	// { <fromRequest>, 1, <fromRequest>, 1 }
	topicNames := make([]string, 0, 4)
	partitionCounts := make([]int, 0, 4)

	topicNames = append(topicNames, inTopicName, notificationTopicName)
	partitionCounts = append(partitionCounts, int(*request.NumPartitions), onePartition)
	if validationEnabled {
		topicNames = append(topicNames, outTopicName, invalidTopicName)
		partitionCounts = append(partitionCounts, int(*request.NumPartitions), onePartition)
	}
	topicSpecs := buildTopicSpecifications(topicNames, partitionCounts, topicConfig)

	ctx := context.Background()

	results, err := adminClient.CreateTopics(ctx, topicSpecs, cfk.SetAdminRequestTimeout(time.Second*10))
	if err != nil {
		return make([]string, 0), http.StatusInternalServerError, fmt.Errorf("unexpected error creating Kafka topics: %w", err)
	}

	createdTopics := make([]string, 0, 4)
	errs := make([]cfk.Error, 0, 4)
	for _, res := range results {
		if res.Error.Code() != cfk.ErrNoError {
			logger.Errorf(res.Error.Error())
			errs = append(errs, res.Error)
		} else {
			createdTopics = append(createdTopics, res.Topic)
		}
	}

	returnCode := http.StatusCreated
	err = nil
	if len(errs) > 0 {
		err = errs[0]
		returnCode = getResponseCodeAndErrorMessage(errs[0])
	}

	return createdTopics, returnCode, err
}

func buildTopicSpecifications(topicNames []string, partitionCounts []int, config map[string]string) []cfk.TopicSpecification {

	topicSpecs := make([]cfk.TopicSpecification, len(topicNames))
	for i, _ := range topicNames {
		topicSpecs[i] = cfk.TopicSpecification{
			Topic:             topicNames[i],
			NumPartitions:     partitionCounts[i],
			ReplicationFactor: replicationFactor,
			Config:            config,
		}
	}
	return topicSpecs

}

func setUpTopicConfig(request model.CreateStreamsRequest) map[string]string {

	config := map[string]string{
		"retention.ms": strconv.Itoa(*request.RetentionMs),
	}

	if request.RetentionBytes != nil {
		config["retention.bytes"] = strconv.Itoa(*request.RetentionBytes)
	}

	if request.CleanupPolicy != nil {
		config["cleanup.policy"] = *request.CleanupPolicy
	}

	if request.SegmentMs != nil {
		config["segment.ms"] = strconv.Itoa(*request.SegmentMs)
	}

	if request.SegmentBytes != nil {
		config["segment.bytes"] = strconv.Itoa(*request.SegmentBytes)
	}

	if request.SegmentIndexBytes != nil {
		config["segment.index.bytes"] = strconv.Itoa(*request.SegmentIndexBytes)
	}

	return config
}

func getResponseCodeAndErrorMessage(err cfk.Error) int {
	code := err.Code()
	// In confluent-kafka-go library, Auth errors seem to get logged
	// as auth errors, but returned with very generic error messages so
	// we can't distinguish them
	if code == cfk.ErrTopicAlreadyExists {
		return http.StatusConflict
	} else if code == cfk.ErrPolicyViolation || code == cfk.ErrInvalidConfig {
		// invalid int values return policy violation, invalid cleanup policy gives invalid config.
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}
