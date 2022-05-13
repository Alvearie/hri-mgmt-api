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
	"github.com/Alvearie/hri-mgmt-api/common/test"
	cfk "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const (
	numPartitions                 int64  = 2
	retentionMs                   int    = 86400000
	topicAlreadyExistsMsgTemplate string = "Topic '%s' already exists."
	timedOutMessage               string = "Failed while waiting for controller: Local: Timed out"
)

var StatusForbidden = http.Response{StatusCode: 403}
var StatusUnauthorized = http.Response{StatusCode: 401}

var StatusUnprocessableEntity = http.Response{StatusCode: 422}

var getInt64Pointer = func(num int64) *int64 {
	return &num
}

var getIntPointer = func(num int) *int {
	return &num
}

var getStringPointer = func(str string) *string {
	return &str
}

func TestCreate(t *testing.T) {
	logwrapper.Initialize("error", os.Stdout)

	tenantId := "tenant123"
	streamId1 := "data-integrator123.qualifier_123"
	requestId := "request-id"
	baseTopicName := strings.Join([]string{tenantId, streamId1}, ".")
	inTopicName := kafka.TopicPrefix + baseTopicName + kafka.InSuffix
	notificationTopicName := kafka.TopicPrefix + baseTopicName + kafka.NotificationSuffix
	outTopicName := kafka.TopicPrefix + baseTopicName + kafka.OutSuffix
	invalidTopicName := kafka.TopicPrefix + baseTopicName + kafka.InvalidSuffix

	validStreamsRequest := model.CreateStreamsRequest{
		NumPartitions: getInt64Pointer(numPartitions),
		RetentionMs:   getIntPointer(retentionMs),
	}

	testCases := []struct {
		name               string
		streamsRequest     model.CreateStreamsRequest
		tenantId           string
		streamId           string
		topic              string
		createResults      []cfk.TopicResult
		createError        error
		validationEnabled  bool
		expectedTopics     []string
		expectedReturnCode int
		expectedError      error
	}{
		{
			name:           "not-authorized",
			streamsRequest: validStreamsRequest,
			tenantId:       tenantId,
			streamId:       streamId1,
			createResults: []cfk.TopicResult{
				{Topic: "in", Error: cfk.NewError(cfk.ErrTopicAuthorizationFailed, unauthorizedMessage, false)},
			},
			expectedReturnCode: http.StatusUnauthorized,
			expectedError:      cfk.NewError(cfk.ErrTopicAuthorizationFailed, unauthorizedMessage, false),
		},
		{
			name:               "timed-out",
			streamsRequest:     validStreamsRequest,
			tenantId:           tenantId,
			streamId:           streamId1,
			createResults:      []cfk.TopicResult{},
			createError:        cfk.NewError(cfk.ErrTimedOut, timedOutMessage, false),
			expectedReturnCode: http.StatusInternalServerError,
			expectedError:      fmt.Errorf("unexpected error creating Kafka topics: %w", cfk.NewError(cfk.ErrTimedOut, timedOutMessage, false)),
		},
		{
			name:           "in-topic-already-exists",
			streamsRequest: validStreamsRequest,
			tenantId:       tenantId,
			streamId:       streamId1,
			createResults: []cfk.TopicResult{
				{Topic: inTopicName, Error: cfk.NewError(cfk.ErrTopicAlreadyExists, fmt.Sprintf(topicAlreadyExistsMsgTemplate, inTopicName), false)},
				{Topic: notificationTopicName},
			},
			expectedReturnCode: http.StatusConflict,
			expectedTopics:     []string{notificationTopicName},
			expectedError:      cfk.NewError(cfk.ErrTopicAlreadyExists, fmt.Sprintf(topicAlreadyExistsMsgTemplate, inTopicName), false),
		},
		{
			name:           "notification-topic-already-exists",
			streamsRequest: validStreamsRequest,
			tenantId:       tenantId,
			streamId:       streamId1,
			createResults: []cfk.TopicResult{
				{Topic: inTopicName},
				{Topic: notificationTopicName, Error: cfk.NewError(cfk.ErrTopicAlreadyExists, fmt.Sprintf(topicAlreadyExistsMsgTemplate, notificationTopicName), false)},
			},
			expectedReturnCode: http.StatusConflict,
			expectedTopics:     []string{inTopicName},
			expectedError:      cfk.NewError(cfk.ErrTopicAlreadyExists, fmt.Sprintf(topicAlreadyExistsMsgTemplate, notificationTopicName), false),
		},
		{
			name:              "out-topic-already-exists",
			streamsRequest:    validStreamsRequest,
			tenantId:          tenantId,
			streamId:          streamId1,
			validationEnabled: true,
			createResults: []cfk.TopicResult{
				{Topic: inTopicName},
				{Topic: notificationTopicName},
				{Topic: outTopicName, Error: cfk.NewError(cfk.ErrTopicAlreadyExists, fmt.Sprintf(topicAlreadyExistsMsgTemplate, outTopicName), false)},
				{Topic: invalidTopicName},
			},
			expectedReturnCode: http.StatusConflict,
			expectedTopics:     []string{inTopicName, notificationTopicName, invalidTopicName},
			expectedError:      cfk.NewError(cfk.ErrTopicAlreadyExists, fmt.Sprintf(topicAlreadyExistsMsgTemplate, outTopicName), false),
		},
		{
			name:              "two-topics-already-exist",
			streamsRequest:    validStreamsRequest,
			tenantId:          tenantId,
			streamId:          streamId1,
			validationEnabled: true,
			createResults: []cfk.TopicResult{
				{Topic: inTopicName},
				{Topic: notificationTopicName, Error: cfk.NewError(cfk.ErrTopicAlreadyExists, fmt.Sprintf(topicAlreadyExistsMsgTemplate, notificationTopicName), false)},
				{Topic: outTopicName, Error: cfk.NewError(cfk.ErrTopicAlreadyExists, fmt.Sprintf(topicAlreadyExistsMsgTemplate, outTopicName), false)},
				{Topic: invalidTopicName},
			},
			expectedReturnCode: http.StatusConflict,
			expectedTopics:     []string{inTopicName, invalidTopicName},
			expectedError:      cfk.NewError(cfk.ErrTopicAlreadyExists, fmt.Sprintf(topicAlreadyExistsMsgTemplate, notificationTopicName), false),
		},
		{
			name:           "invalid-cleanup-policy",
			streamsRequest: validStreamsRequest,
			tenantId:       tenantId,
			streamId:       streamId1,
			createResults: []cfk.TopicResult{
				{Topic: inTopicName, Error: cfk.NewError(cfk.ErrInvalidConfig, "Invalid value abcd for configuration cleanup.policy: String must be one of: compact, delete", false)},
				{Topic: notificationTopicName},
			},
			expectedReturnCode: http.StatusBadRequest,
			expectedTopics:     []string{notificationTopicName},
			expectedError:      cfk.NewError(cfk.ErrInvalidConfig, "Invalid value abcd for configuration cleanup.policy: String must be one of: compact, delete", false),
		},
		{
			name:           "invalid-retention-ms",
			streamsRequest: validStreamsRequest,
			tenantId:       tenantId,
			streamId:       streamId1,
			createResults: []cfk.TopicResult{
				{Topic: inTopicName, Error: cfk.NewError(cfk.ErrPolicyViolation, "Invalid retention.ms specified. The allowed range is [3600000..2592000000]", false)},
				{Topic: notificationTopicName},
			},
			expectedReturnCode: http.StatusBadRequest,
			expectedTopics:     []string{notificationTopicName},
			expectedError:      cfk.NewError(cfk.ErrPolicyViolation, "Invalid retention.ms specified. The allowed range is [3600000..2592000000]", false),
		},
		{
			name:           "good-request-no-qualifier-with-validation",
			streamsRequest: validStreamsRequest,
			tenantId:       tenantId,
			streamId:       streamId1,
			createResults: []cfk.TopicResult{
				{Topic: inTopicName},
				{Topic: notificationTopicName},
				{Topic: outTopicName},
				{Topic: invalidTopicName},
			},
			validationEnabled:  true,
			expectedReturnCode: http.StatusCreated,
			expectedTopics:     []string{inTopicName, notificationTopicName, outTopicName, invalidTopicName},
		},
	}

	for _, tc := range testCases {
		controller := gomock.NewController(t)
		defer controller.Finish()
		mockService := test.NewMockService(controller)

		topicNames := []string{inTopicName, notificationTopicName}
		if tc.validationEnabled {
			topicNames = []string{inTopicName, notificationTopicName, outTopicName, invalidTopicName}
		}
		partitionCounts := []int{int(numPartitions), 1, int(numPartitions), 1}
		config := setUpTopicConfig(tc.streamsRequest)
		topicSpecifications := buildTopicSpecifications(topicNames, partitionCounts, config)

		mockService.
			EXPECT().
			CreateTopics(context.Background(), topicSpecifications, cfk.SetAdminRequestTimeout(kafka.AdminTimeout)).
			Return(tc.createResults, tc.createError).
			MaxTimes(1)

		t.Run(tc.name, func(t *testing.T) {
			topicsCreated, returnCode, err := Create(tc.streamsRequest, tc.tenantId, tc.streamId, tc.validationEnabled, requestId, mockService)

			if tc.expectedTopics == nil {
				tc.expectedTopics = make([]string, 0)
			}

			actualReturnedValues := map[string]interface{}{
				"topicsCreated": topicsCreated,
				"returnCode":    returnCode,
				"errorMessage":  err,
			}

			expectedReturnValues := map[string]interface{}{
				"topicsCreated": tc.expectedTopics,
				"returnCode":    tc.expectedReturnCode,
				"errorMessage":  tc.expectedError,
			}

			if !reflect.DeepEqual(expectedReturnValues, actualReturnedValues) {
				t.Error(fmt.Sprintf("Expected: [%v], actual: [%v]", expectedReturnValues, actualReturnedValues))
			}
		})
	}
}

func TestSetUpTopicConfigs(t *testing.T) {
	var configsValue = 10485760
	expectedConfig := map[string]string{
		"retention.ms":        strconv.Itoa(retentionMs),
		"retention.bytes":     strconv.Itoa(configsValue),
		"cleanup.policy":      "compact",
		"segment.ms":          strconv.Itoa(configsValue),
		"segment.bytes":       strconv.Itoa(configsValue),
		"segment.index.bytes": strconv.Itoa(configsValue),
	}

	validArgs := model.CreateStreamsRequest{
		RetentionMs:       getIntPointer(retentionMs),
		RetentionBytes:    getIntPointer(configsValue),
		CleanupPolicy:     getStringPointer("compact"),
		SegmentMs:         getIntPointer(configsValue),
		SegmentBytes:      getIntPointer(configsValue),
		SegmentIndexBytes: getIntPointer(configsValue),
	}
	config := setUpTopicConfig(validArgs)
	assert.Equal(t, expectedConfig, config)
}

func TestCreateRequestValidation(t *testing.T) {
	testCases := []struct {
		name                       string
		request                    model.CreateStreamsRequest
		expectedValidationFailures map[string]string
	}{
		{
			name:    "missing required fields fail validation",
			request: model.CreateStreamsRequest{
				// Required fields omitted
			},
			expectedValidationFailures: map[string]string{
				"TenantId":      "required",
				"StreamId":      "required",
				"NumPartitions": "required",
				"RetentionMs":   "required",
			},
		},
		{
			name: "values smaller than minimum fail validation",
			request: model.CreateStreamsRequest{
				TenantId:          test.ValidTenantId,
				StreamId:          test.ValidStreamId,
				NumPartitions:     getInt64Pointer(-1),
				RetentionMs:       getIntPointer(-1),
				RetentionBytes:    getIntPointer(-1),
				SegmentMs:         getIntPointer(-1),
				SegmentBytes:      getIntPointer(-1),
				SegmentIndexBytes: getIntPointer(-1),
			},
			expectedValidationFailures: map[string]string{
				"NumPartitions":     "min",
				"RetentionMs":       "min",
				"RetentionBytes":    "min",
				"SegmentMs":         "min",
				"SegmentBytes":      "min",
				"SegmentIndexBytes": "min",
			},
		},
		{
			name: "values bigger than maximum fail validation",
			request: model.CreateStreamsRequest{
				TenantId:          test.ValidTenantId,
				StreamId:          test.ValidStreamId,
				NumPartitions:     getInt64Pointer(100),
				RetentionMs:       getIntPointer(2592000001),
				RetentionBytes:    getIntPointer(1073741825),
				SegmentMs:         getIntPointer(2592000001),
				SegmentBytes:      getIntPointer(536870913),
				SegmentIndexBytes: getIntPointer(104857601),
			},
			expectedValidationFailures: map[string]string{
				"NumPartitions":     "max",
				"RetentionMs":       "max",
				"RetentionBytes":    "max",
				"SegmentMs":         "max",
				"SegmentBytes":      "max",
				"SegmentIndexBytes": "max",
			},
		},
		{
			name: "invalid cleanupPolicy fails validation",
			request: model.CreateStreamsRequest{
				TenantId:      test.ValidTenantId,
				StreamId:      test.ValidStreamId,
				NumPartitions: getInt64Pointer(1),
				RetentionMs:   getIntPointer(3600000),
				CleanupPolicy: getStringPointer("bogus"),
			},
			expectedValidationFailures: map[string]string{
				"CleanupPolicy": "oneof",
			},
		},
		{
			name: "delete is a valid cleanupPolicy",
			request: model.CreateStreamsRequest{
				TenantId:      test.ValidTenantId,
				StreamId:      test.ValidStreamId,
				NumPartitions: getInt64Pointer(1),
				RetentionMs:   getIntPointer(3600000),
				CleanupPolicy: getStringPointer("delete"),
			},
		},
		{
			name: "compact is a valid cleanupPolicy",
			request: model.CreateStreamsRequest{
				TenantId:      test.ValidTenantId,
				StreamId:      test.ValidStreamId,
				NumPartitions: getInt64Pointer(1),
				RetentionMs:   getIntPointer(3600000),
				CleanupPolicy: getStringPointer("compact"),
			},
		},
		{
			name: "valid request",
			request: model.CreateStreamsRequest{
				TenantId:          test.ValidTenantId,
				StreamId:          test.ValidStreamId,
				NumPartitions:     getInt64Pointer(1),
				RetentionMs:       getIntPointer(3600000),
				CleanupPolicy:     getStringPointer("compact"),
				RetentionBytes:    getIntPointer(10485760),
				SegmentMs:         getIntPointer(300000),
				SegmentBytes:      getIntPointer(10485760),
				SegmentIndexBytes: getIntPointer(102400),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			customValidator, _ := model.GetValidator()
			err := customValidator.Validator.Struct(tc.request)

			if validationErrors, isValidationErrors := err.(validator.ValidationErrors); isValidationErrors {
				expectedInvalidFields := make([]string, len(tc.expectedValidationFailures))
				for key := range tc.expectedValidationFailures {
					expectedInvalidFields = append(expectedInvalidFields, key)
				}
				actualInvalidFields := make([]string, len(tc.expectedValidationFailures))

				for _, fieldError := range validationErrors {
					assert.Equal(t, tc.expectedValidationFailures[fieldError.StructField()], fieldError.Tag(),
						fmt.Sprintf("Unexpected validation error for %s", fieldError.StructField()))
					actualInvalidFields = append(actualInvalidFields, fieldError.StructField())
				}

				sort.Strings(actualInvalidFields)
				sort.Strings(expectedInvalidFields)
				if !reflect.DeepEqual(actualInvalidFields, expectedInvalidFields) {
					t.Error(fmt.Sprintf("Expected invalid fields: [%v], actual invalid fields: [%v]",
						expectedInvalidFields, actualInvalidFields))
				}
			} else if err != nil {
				assert.Fail(t, "unexpectedly received InvalidValidationError", err)
			}
		})
	}
}
