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
)

const (
	//error codes returned by the IBM EventStreams Admin API
	topicAlreadyExists   int32 = 42236
	invalidCleanupPolicy int32 = 42240

	cleanupFailureMsg string = ", cleanup of associated topic [%s] also failed: %s"
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

	inTopicName, notificationTopicName := eventstreams.CreateTopicNames(tenantId, streamId)

	//numPartitions and retentionTime configs are required, the rest are optional
	numPartitions := int64(args[param.NumPartitions].(float64))
	topicConfigs := setUpTopicConfigs(args)

	inTopicRequest := es.TopicCreateRequest{
		Name:           inTopicName,
		PartitionCount: numPartitions,
		Configs:        topicConfigs,
	}
	notificationTopicRequest := es.TopicCreateRequest{
		Name:           notificationTopicName,
		PartitionCount: numPartitions,
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
		_, _, deleteErr := service.DeleteTopic(context.Background(), inTopicName)
		var deleteErrorMsg = ""
		if deleteErr != nil {
			logger.Printf("Unable to delete topic [%s]. %s", inTopicName, deleteErr.Error())
			deleteErrorMsg = fmt.Sprintf(cleanupFailureMsg, inTopicName, deleteErr.Error())
		}
		return getCreateResponseError(notificationResponse, service.HandleModelError(notificationErr), deleteErrorMsg)
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
