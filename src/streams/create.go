/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package streams

import (
	"context"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/eventstreams"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	es "github.com/IBM/event-streams-go-sdk-generator/build/generated"
	"net/http"
	"strconv"
)

const (
	//error codes returned by the IBM EventStreams Admin API
	topicAlreadyExists   int32 = 42236
	invalidCleanupPolicy int32 = 42240
	onePartition         int64 = 1
)

func Create(
	request model.CreateStreamsRequest,
	tenantId string,
	streamId string,
	validationEnabled bool,
	requestId string,
	service eventstreams.Service) ([]string, int, error) {

	prefix := "streams/create"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Streams Create")

	inTopicName, notificationTopicName, outTopicName, invalidTopicName := eventstreams.CreateTopicNames(
		tenantId, streamId)

	//numPartitions and retentionTime configs are required, the rest are optional
	topicConfigs := setUpTopicConfigs(request)

	inTopicRequest := es.TopicCreateRequest{
		Name:           inTopicName,
		PartitionCount: *request.NumPartitions,
		Configs:        topicConfigs,
	}

	//create notification topic with only 1 Partition b/c of the small expected msg volume for this topic
	notificationTopicRequest := es.TopicCreateRequest{
		Name:           notificationTopicName,
		PartitionCount: onePartition,
		Configs:        topicConfigs,
	}

	createdTopics := make([]string, 0, 4)

	// create the input and notification topics for the given tenant and stream pairing
	_, inResponse, inErr := service.CreateTopic(context.Background(), inTopicRequest)
	if inErr != nil {
		logger.Errorf("Unable to create new topic [%s]. %s", inTopicName, inErr.Error())
		responseCode, errorMessage := getResponseCodeAndErrorMessage(inResponse, service.HandleModelError(inErr))
		return createdTopics, responseCode, fmt.Errorf(errorMessage)
	}
	createdTopics = append(createdTopics, inTopicName)

	_, notificationResponse, notificationErr := service.CreateTopic(context.Background(), notificationTopicRequest)
	if notificationErr != nil {
		logger.Errorf("Unable to create new topic [%s]. %s", notificationTopicName, notificationErr.Error())
		responseCode, errorMessage := getResponseCodeAndErrorMessage(notificationResponse, service.HandleModelError(notificationErr))
		return createdTopics, responseCode, fmt.Errorf(errorMessage)
	}
	createdTopics = append(createdTopics, notificationTopicName)

	// if validation is enabled, create the out and invalid topics for the given tenant and stream pairing
	if validationEnabled {
		outTopicRequest := es.TopicCreateRequest{
			Name:           outTopicName,
			PartitionCount: *request.NumPartitions,
			Configs:        topicConfigs,
		}

		//create invalid topic with only 1 Partition b/c of the small expected msg volume for this topic
		invalidTopicRequest := es.TopicCreateRequest{
			Name:           invalidTopicName,
			PartitionCount: onePartition,
			Configs:        topicConfigs,
		}

		_, outResponse, outErr := service.CreateTopic(context.Background(), outTopicRequest)
		if outErr != nil {
			logger.Errorf("Unable to create new topic [%s]. %s", outTopicName, outErr.Error())
			responseCode, errorMessage := getResponseCodeAndErrorMessage(outResponse, service.HandleModelError(outErr))
			return createdTopics, responseCode, fmt.Errorf(errorMessage)
		}
		createdTopics = append(createdTopics, outTopicName)

		_, invalidResponse, invalidErr := service.CreateTopic(context.Background(), invalidTopicRequest)
		if invalidErr != nil {
			logger.Errorf("Unable to create new topic [%s]. %s", invalidTopicName, invalidErr.Error())
			responseCode, errorMessage := getResponseCodeAndErrorMessage(invalidResponse, service.HandleModelError(invalidErr))
			return createdTopics, responseCode, fmt.Errorf(errorMessage)
		}
		createdTopics = append(createdTopics, invalidTopicName)
	}

	return createdTopics, http.StatusCreated, nil
}

func setUpTopicConfigs(request model.CreateStreamsRequest) []es.ConfigCreate {
	retentionMsConfig := es.ConfigCreate{
		Name:  "retention.ms",
		Value: strconv.Itoa(*request.RetentionMs),
	}
	configs := []es.ConfigCreate{retentionMsConfig}

	if request.RetentionBytes != nil {
		retentionBytesConfig := es.ConfigCreate{
			Name:  "retention.bytes",
			Value: strconv.Itoa(*request.RetentionBytes),
		}
		configs = append(configs, retentionBytesConfig)
	}

	if request.CleanupPolicy != nil {
		cleanupPolicyConfig := es.ConfigCreate{
			Name:  "cleanup.policy",
			Value: *request.CleanupPolicy,
		}
		configs = append(configs, cleanupPolicyConfig)
	}

	if request.SegmentMs != nil {
		segmentMsConfig := es.ConfigCreate{
			Name:  "segment.ms",
			Value: strconv.Itoa(*request.SegmentMs),
		}
		configs = append(configs, segmentMsConfig)
	}

	if request.SegmentBytes != nil {
		segmentBytesConfig := es.ConfigCreate{
			Name:  "segment.bytes",
			Value: strconv.Itoa(*request.SegmentBytes),
		}
		configs = append(configs, segmentBytesConfig)
	}

	if request.SegmentIndexBytes != nil {
		segmentIndexBytesConfig := es.ConfigCreate{
			Name:  "segment.index.bytes",
			Value: strconv.Itoa(*request.SegmentIndexBytes),
		}
		configs = append(configs, segmentIndexBytesConfig)
	}

	return configs
}

func getResponseCodeAndErrorMessage(resp *http.Response, err *es.ModelError) (int, string) {
	//EventStreams Admin API gives us status 403 when provided bearer token is unauthorized
	//and status 401 when Authorization isn't provided or is nil
	if resp.StatusCode == http.StatusForbidden {
		return http.StatusUnauthorized, eventstreams.UnauthorizedMsg
	} else if resp.StatusCode == http.StatusUnauthorized {
		return http.StatusUnauthorized, eventstreams.MissingHeaderMsg
	} else if resp.StatusCode == http.StatusUnprocessableEntity && err.ErrorCode == topicAlreadyExists {
		return http.StatusConflict, err.Message
	} else if resp.StatusCode == http.StatusUnprocessableEntity && err.ErrorCode == invalidCleanupPolicy {
		return http.StatusBadRequest, err.Message
	}
	return http.StatusInternalServerError, err.Message
}
