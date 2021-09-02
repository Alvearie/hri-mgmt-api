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
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/path"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	es "github.com/IBM/event-streams-go-sdk-generator/build/generated"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const (
	//error codes returned by the IBM EventStreams Admin API
	topicAlreadyExists   int32 = 42236
	invalidCleanupPolicy int32 = 42240

	cleanupFailureMsg string = ", cleanup of associated topic [%s] also failed: %s"

	onePartition int64 = 1
)

func Create(
	args map[string]interface{},
	validator param.Validator,
	service eventstreams.Service) map[string]interface{} {

	logger := log.New(os.Stdout, "streams/create: ", log.Llongfile)

	// validate that required input params are present
	errResp := validator.Validate(
		args,
		param.Info{param.NumPartitions, reflect.Float64},
		param.Info{param.RetentionMs, reflect.Float64},
		param.Info{param.Validation, reflect.Bool},
	)
	if errResp != nil {
		logger.Printf("Bad input params: %s", errResp)
		return errResp
	}

	// validate that any provided optional input params are of the right type
	errResp = validator.ValidateOptional(
		args,
		param.Info{param.CleanupPolicy, reflect.String},
		param.Info{param.RetentionBytes, reflect.Float64},
		param.Info{param.SegmentMs, reflect.Float64},
		param.Info{param.SegmentBytes, reflect.Float64},
		param.Info{param.SegmentIndexBytes, reflect.Float64},
	)
	if errResp != nil {
		logger.Printf("Bad input optional params: %s", errResp)
		return errResp
	}

	// extract tenantId and streamId path params from URL
	tenantId, err := path.ExtractParam(args, param.TenantIndex)
	if err != nil {
		logger.Println(err.Error())
		return response.Error(http.StatusBadRequest, err.Error())
	}

	if err := param.TenantIdCheck(tenantId); err != nil {
		return response.Error(http.StatusBadRequest, err.Error())
	}

	streamId, err := path.ExtractParam(args, param.StreamIndex)
	if err != nil {
		logger.Println(err.Error())
		return response.Error(http.StatusBadRequest, err.Error())
	}

	if err := param.StreamIdCheck(streamId); err != nil {
		return response.Error(http.StatusBadRequest, err.Error())
	}

	inTopicName, notificationTopicName, outTopicName, invalidTopicName := eventstreams.CreateTopicNames(tenantId, streamId)

	//numPartitions and retentionTime configs are required, the rest are optional
	numPartitions := int64(args[param.NumPartitions].(float64))
	topicConfigs := setUpTopicConfigs(args)

	inTopicRequest := es.TopicCreateRequest{
		Name:           inTopicName,
		PartitionCount: numPartitions,
		Configs:        topicConfigs,
	}

	//create notification topic with only 1 Partition b/c of the small expected msg volume for this topic
	notificationTopicRequest := es.TopicCreateRequest{
		Name:           notificationTopicName,
		PartitionCount: onePartition,
		Configs:        topicConfigs,
	}

	// create the in and notification topics for the given tenant and stream pairing
	_, inResponse, inErr := service.CreateTopic(context.Background(), inTopicRequest)
	if inErr != nil {
		logger.Printf("Unable to create new topic [%s]. %s", inTopicName, inErr.Error())
		return getCreateResponseError(inResponse, service.HandleModelError(inErr), "")
	}

	_, notificationResponse, notificationErr := service.CreateTopic(context.Background(), notificationTopicRequest)
	if notificationErr != nil {
		logger.Printf("Unable to create new topic [%s]. %s", notificationTopicName, notificationErr.Error())
		logger.Printf("Deleting associated 'in' topic [%s]...", inTopicName)
		inTopicDeleted := deleteTopic(context.Background(), inTopicName, service, logger) //value will be bool, true if deleted successfully without errors

		var deleteTopicsErrorMsg = createDeleteErrorMsg(inTopicName, inTopicDeleted, true, true)
		return getCreateResponseError(notificationResponse, service.HandleModelError(notificationErr), deleteTopicsErrorMsg)
	}

	validation := args[param.Validation].(bool)
	logger.Printf("Value of validation is [%t]", validation)

	// if validation is enabled, create the out and invalid topics for the given tenant and stream pairing
	if validation {
		outTopicRequest := es.TopicCreateRequest{
			Name:           outTopicName,
			PartitionCount: numPartitions,
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
			logger.Printf("Unable to create new topic [%s]. %s", outTopicName, outErr.Error())

			//delete associated in and notification topics
			logger.Printf("Deleting associated 'in' topic [%s]...", inTopicName)
			inTopicDeleted := deleteTopic(context.Background(), inTopicName, service, logger)
			logger.Printf("Deleting associated 'notification' topic [%s]...", notificationTopicName)
			notificationTopicDeleted := deleteTopic(context.Background(), notificationTopicName, service, logger)

			var deleteTopicsErrorMsg = createDeleteErrorMsg(inTopicName, inTopicDeleted, notificationTopicDeleted, true)
			return getCreateResponseError(outResponse, service.HandleModelError(outErr), deleteTopicsErrorMsg)
		}

		_, invalidResponse, invalidErr := service.CreateTopic(context.Background(), invalidTopicRequest)
		if invalidErr != nil {
			logger.Printf("Unable to create new topic [%s]. %s", invalidTopicName, invalidErr.Error())

			//delete associated in, notification, and out topics
			logger.Printf("Deleting associated 'in' topic [%s]...", inTopicName)
			inTopicDeleted := deleteTopic(context.Background(), inTopicName, service, logger)
			logger.Printf("Deleting associated 'notification' topic [%s]...", notificationTopicName)
			notificationTopicDeleted := deleteTopic(context.Background(), notificationTopicName, service, logger)
			logger.Printf("Deleting associated 'out' topic [%s]...", outTopicName)
			outTopicDeleted := deleteTopic(context.Background(), outTopicName, service, logger)

			var deleteTopicsErrorMsg = createDeleteErrorMsg(inTopicName, inTopicDeleted, notificationTopicDeleted, outTopicDeleted)
			return getCreateResponseError(invalidResponse, service.HandleModelError(invalidErr), deleteTopicsErrorMsg)
		}
	}

	// return the name of the created stream
	respBody := map[string]interface{}{param.StreamId: streamId}
	return response.Success(http.StatusCreated, respBody)
}

func setUpTopicConfigs(args map[string]interface{}) []es.ConfigCreate {
	retentionMs := int(args[param.RetentionMs].(float64))
	retentionMsConfig := es.ConfigCreate{
		Name:  "retention.ms",
		Value: strconv.Itoa(retentionMs),
	}
	configs := []es.ConfigCreate{retentionMsConfig}

	if args[param.RetentionBytes] != nil {
		retentionBytes := int(args[param.RetentionBytes].(float64))
		retentionBytesConfig := es.ConfigCreate{
			Name:  "retention.bytes",
			Value: strconv.Itoa(retentionBytes),
		}
		configs = append(configs, retentionBytesConfig)
	}

	if args[param.CleanupPolicy] != nil {
		cleanupPolicy := args[param.CleanupPolicy].(string)
		cleanupPolicyConfig := es.ConfigCreate{
			Name:  "cleanup.policy",
			Value: cleanupPolicy,
		}
		configs = append(configs, cleanupPolicyConfig)
	}

	if args[param.SegmentMs] != nil {
		segmentMs := int(args[param.SegmentMs].(float64))
		segmentMsConfig := es.ConfigCreate{
			Name:  "segment.ms",
			Value: strconv.Itoa(segmentMs),
		}
		configs = append(configs, segmentMsConfig)
	}

	if args[param.SegmentBytes] != nil {
		segmentBytes := int(args[param.SegmentBytes].(float64))
		segmentBytesConfig := es.ConfigCreate{
			Name:  "segment.bytes",
			Value: strconv.Itoa(segmentBytes),
		}
		configs = append(configs, segmentBytesConfig)
	}

	if args[param.SegmentIndexBytes] != nil {
		segmentIndexBytes := int(args[param.SegmentIndexBytes].(float64))
		segmentIndexBytesConfig := es.ConfigCreate{
			Name:  "segment.index.bytes",
			Value: strconv.Itoa(segmentIndexBytes),
		}
		configs = append(configs, segmentIndexBytesConfig)
	}

	return configs
}

func getCreateResponseError(resp *http.Response, err *es.ModelError, deleteError string) map[string]interface{} {

	//EventStreams Admin API gives us status 403 when provided bearer token is unauthorized
	//and status 401 when Authorization isn't provided or is nil
	if resp.StatusCode == http.StatusForbidden {
		return response.Error(http.StatusUnauthorized, eventstreams.UnauthorizedMsg+deleteError)
	} else if resp.StatusCode == http.StatusUnauthorized {
		return response.Error(http.StatusUnauthorized, eventstreams.MissingHeaderMsg+deleteError)
	} else if resp.StatusCode == http.StatusUnprocessableEntity && err.ErrorCode == topicAlreadyExists {
		return response.Error(http.StatusConflict, err.Message+deleteError)
	} else if resp.StatusCode == http.StatusUnprocessableEntity && err.ErrorCode == invalidCleanupPolicy {
		return response.Error(http.StatusBadRequest, err.Message+deleteError)
	}
	return response.Error(http.StatusInternalServerError, err.Message+deleteError)
}

func deleteTopic(ctx context.Context, topicName string, service eventstreams.Service, logger *log.Logger) bool {
	_, _, deleteErr := service.DeleteTopic(ctx, topicName)

	if deleteErr != nil {
		logger.Printf("Unable to delete topic [%s]. %s", topicName, deleteErr.Error())
		return false //topic was not deleted successfully and there was an error
	}
	return true //topic was deleted successfully
}

func createDeleteErrorMsg(inTopicName string, inTopicDeleted bool, notificationTopicDeleted bool, outTopicDeleted bool) string {
	if inTopicDeleted && notificationTopicDeleted && outTopicDeleted {
		return "" //if all topics deleted successfully, return empty error message
	}

	//if topics had errors when deleting, add them to the slice
	var deleteErrorTopics []string
	if !inTopicDeleted {
		deleteErrorTopics = append(deleteErrorTopics, "in")
	}
	if !notificationTopicDeleted {
		deleteErrorTopics = append(deleteErrorTopics, "notification")
	}
	if !outTopicDeleted {
		deleteErrorTopics = append(deleteErrorTopics, "out")
	}

	deleteError := fmt.Sprintf("failed to delete %s topic(s)", strings.Join(deleteErrorTopics, ","))

	//get topic name without in suffix
	inTopic := []rune(inTopicName)
	endIndex := len(inTopic) - len(eventstreams.InSuffix)
	topic := string(inTopic[0:endIndex])

	return fmt.Sprintf(cleanupFailureMsg, topic, deleteError)
}
