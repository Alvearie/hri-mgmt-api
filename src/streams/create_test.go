/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package streams

import (
	"context"
	"errors"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/eventstreams"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	es "github.com/IBM/event-streams-go-sdk-generator/build/generated"
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
	numPartitions               int64  = 1
	retentionMs                 int    = 86400000
	topicAlreadyExistsMessage   string = "topic already exists"
	invalidCleanupPolicyMessage string = "invalid cleanup policy"
	forbiddenMessage            string = "forbidden"
	unauthorizedMessage         string = "unauthorized"
	kafkaConnectionMessage      string = "Unable to connect to Kafka"
)

var StatusForbidden = http.Response{StatusCode: 403}
var StatusUnauthorized = http.Response{StatusCode: 401}

var StatusUnprocessableEntity = http.Response{StatusCode: 422}
var TopicAlreadyExistsError = es.ModelError{
	ErrorCode: topicAlreadyExists,
	Message:   topicAlreadyExistsMessage,
}
var invalidCleanupPolicyError = es.ModelError{
	ErrorCode: invalidCleanupPolicy,
	Message:   invalidCleanupPolicyMessage,
}
var OtherError = es.ModelError{
	ErrorCode: 1,
	Message:   kafkaConnectionMessage,
}

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

	validStreamsRequest := model.CreateStreamsRequest{
		NumPartitions: getInt64Pointer(numPartitions),
		RetentionMs:   getIntPointer(retentionMs),
	}

	testCases := []struct {
		name                   string
		streamsRequest         model.CreateStreamsRequest
		tenantId               string
		streamId               string
		validationEnabled      bool
		modelInError           *es.ModelError
		modelNotificationError *es.ModelError
		modelOutError          *es.ModelError
		modelInvalidError      *es.ModelError
		mockResponse           *http.Response
		expectedTopic          string
		expectedTopics         []string
		expectedReturnCode     int
		expectedError          error
	}{
		{
			name:               "not-authorized",
			streamsRequest:     validStreamsRequest,
			tenantId:           tenantId,
			streamId:           streamId1,
			modelInError:       &es.ModelError{},
			mockResponse:       &StatusForbidden,
			expectedError:      fmt.Errorf(eventstreams.UnauthorizedMsg),
			expectedReturnCode: http.StatusUnauthorized,
			expectedTopic:      baseTopicName,
		},
		{
			name:               "missing-auth-header",
			streamsRequest:     validStreamsRequest,
			tenantId:           tenantId,
			streamId:           streamId1,
			modelInError:       &es.ModelError{},
			mockResponse:       &StatusUnauthorized,
			expectedError:      fmt.Errorf(eventstreams.MissingHeaderMsg),
			expectedReturnCode: http.StatusUnauthorized,
			expectedTopic:      baseTopicName,
		},
		{
			name:               "in-topic-already-exists",
			streamsRequest:     validStreamsRequest,
			tenantId:           tenantId,
			streamId:           streamId1,
			modelInError:       &TopicAlreadyExistsError,
			mockResponse:       &StatusUnprocessableEntity,
			expectedError:      fmt.Errorf(topicAlreadyExistsMessage),
			expectedReturnCode: http.StatusConflict,
			expectedTopic:      baseTopicName,
		},
		{
			name:                   "notification-topic-already-exists",
			streamsRequest:         validStreamsRequest,
			tenantId:               tenantId,
			streamId:               streamId1,
			modelNotificationError: &TopicAlreadyExistsError,
			mockResponse:           &StatusUnprocessableEntity,
			expectedError:          fmt.Errorf(topicAlreadyExistsMessage),
			expectedReturnCode:     http.StatusConflict,
			expectedTopic:          baseTopicName,
			expectedTopics:         []string{eventstreams.TopicPrefix + tenantId + "." + streamId1 + eventstreams.InSuffix},
		},
		{
			name:               "out-topic-already-exists",
			streamsRequest:     validStreamsRequest,
			tenantId:           tenantId,
			streamId:           streamId1,
			validationEnabled:  true,
			modelOutError:      &TopicAlreadyExistsError,
			mockResponse:       &StatusUnprocessableEntity,
			expectedError:      fmt.Errorf(topicAlreadyExistsMessage),
			expectedReturnCode: http.StatusConflict,
			expectedTopic:      baseTopicName,
			expectedTopics: []string{
				eventstreams.TopicPrefix + tenantId + "." + streamId1 + eventstreams.InSuffix,
				eventstreams.TopicPrefix + tenantId + "." + streamId1 + eventstreams.NotificationSuffix,
			},
		},
		{
			name:               "invalid-topic-already-exists",
			streamsRequest:     validStreamsRequest,
			tenantId:           tenantId,
			streamId:           streamId1,
			validationEnabled:  true,
			modelInvalidError:  &TopicAlreadyExistsError,
			mockResponse:       &StatusUnprocessableEntity,
			expectedError:      fmt.Errorf(topicAlreadyExistsMessage),
			expectedReturnCode: http.StatusConflict,
			expectedTopic:      baseTopicName,
			expectedTopics: []string{
				eventstreams.TopicPrefix + tenantId + "." + streamId1 + eventstreams.InSuffix,
				eventstreams.TopicPrefix + tenantId + "." + streamId1 + eventstreams.NotificationSuffix,
				eventstreams.TopicPrefix + tenantId + "." + streamId1 + eventstreams.OutSuffix,
			},
		},
		{
			name:               "invalid-cleanup-policy",
			streamsRequest:     validStreamsRequest,
			tenantId:           tenantId,
			streamId:           streamId1,
			modelInError:       &invalidCleanupPolicyError,
			mockResponse:       &StatusUnprocessableEntity,
			expectedReturnCode: http.StatusBadRequest,
			expectedError:      fmt.Errorf(invalidCleanupPolicyMessage),
			expectedTopic:      baseTopicName,
		},
		{
			name:               "in-conn-error",
			streamsRequest:     validStreamsRequest,
			tenantId:           tenantId,
			streamId:           streamId1,
			modelInError:       &OtherError,
			mockResponse:       &StatusUnprocessableEntity,
			expectedReturnCode: http.StatusInternalServerError,
			expectedError:      fmt.Errorf(kafkaConnectionMessage),
			expectedTopic:      baseTopicName,
		},
		{
			name:               "good-request-no-qualifier-with-validation",
			streamsRequest:     validStreamsRequest,
			tenantId:           tenantId,
			streamId:           streamId1,
			validationEnabled:  true,
			expectedReturnCode: http.StatusCreated,
			expectedTopic:      baseTopicName,
			expectedTopics: []string{
				eventstreams.TopicPrefix + tenantId + "." + streamId1 + eventstreams.InSuffix,
				eventstreams.TopicPrefix + tenantId + "." + streamId1 + eventstreams.NotificationSuffix,
				eventstreams.TopicPrefix + tenantId + "." + streamId1 + eventstreams.OutSuffix,
				eventstreams.TopicPrefix + tenantId + "." + streamId1 + eventstreams.InvalidSuffix,
			},
		},
	}

	for _, tc := range testCases {
		controller := gomock.NewController(t)
		defer controller.Finish()
		mockService := test.NewMockService(controller)

		var mockInErr error
		var mockNotificationErr error
		var mockOutErr error
		var mockInvalidErr error
		if tc.modelInError != nil {
			mockInErr = errors.New(tc.modelInError.Message)
		}
		if tc.modelNotificationError != nil {
			mockNotificationErr = errors.New(tc.modelNotificationError.Message)
		}
		if tc.modelOutError != nil {
			mockOutErr = errors.New(tc.modelOutError.Message)
		}
		if tc.modelInvalidError != nil {
			mockInvalidErr = errors.New(tc.modelInvalidError.Message)
		}

		mockService.
			EXPECT().
			CreateTopic(context.Background(), getTestTopicRequest(tc.expectedTopic, eventstreams.InSuffix)).
			Return(nil, tc.mockResponse, mockInErr).
			MaxTimes(1)

		mockService.
			EXPECT().
			CreateTopic(context.Background(), getTestTopicRequest(tc.expectedTopic, eventstreams.NotificationSuffix)).
			Return(nil, tc.mockResponse, mockNotificationErr).
			MaxTimes(1)

		mockService.
			EXPECT().
			CreateTopic(context.Background(), getTestTopicRequest(tc.expectedTopic, eventstreams.OutSuffix)).
			Return(nil, tc.mockResponse, mockOutErr).
			MaxTimes(1)

		mockService.
			EXPECT().
			CreateTopic(context.Background(), getTestTopicRequest(tc.expectedTopic, eventstreams.InvalidSuffix)).
			Return(nil, tc.mockResponse, mockInvalidErr).
			MaxTimes(1)

		mockService.
			EXPECT().
			HandleModelError(mockInErr).
			Return(tc.modelInError).
			MaxTimes(1)

		mockService.
			EXPECT().
			HandleModelError(mockNotificationErr).
			Return(tc.modelNotificationError).
			MaxTimes(1)

		mockService.
			EXPECT().
			HandleModelError(mockOutErr).
			Return(tc.modelOutError).
			MaxTimes(1)

		mockService.
			EXPECT().
			HandleModelError(mockInvalidErr).
			Return(tc.modelInvalidError).
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
	expectedConfigs := []es.ConfigCreate{
		{
			Name:  "retention.ms",
			Value: strconv.Itoa(retentionMs),
		},
		{
			Name:  "retention.bytes",
			Value: strconv.Itoa(configsValue),
		},
		{
			Name:  "cleanup.policy",
			Value: "compact",
		},
		{
			Name:  "segment.ms",
			Value: strconv.Itoa(configsValue),
		},
		{
			Name:  "segment.bytes",
			Value: strconv.Itoa(configsValue),
		},
		{
			Name:  "segment.index.bytes",
			Value: strconv.Itoa(configsValue),
		},
	}
	validArgs := model.CreateStreamsRequest{
		RetentionMs:       getIntPointer(retentionMs),
		RetentionBytes:    getIntPointer(configsValue),
		CleanupPolicy:     getStringPointer("compact"),
		SegmentMs:         getIntPointer(configsValue),
		SegmentBytes:      getIntPointer(configsValue),
		SegmentIndexBytes: getIntPointer(configsValue),
	}
	configs := setUpTopicConfigs(validArgs)
	assert.Equal(t, expectedConfigs, configs)
}

func getTestTopicRequest(streamName string, topicSuffix string) es.TopicCreateRequest {
	return es.TopicCreateRequest{
		Name:           eventstreams.TopicPrefix + streamName + topicSuffix,
		PartitionCount: numPartitions,
		Configs: []es.ConfigCreate{{
			Name:  "retention.ms",
			Value: strconv.Itoa(retentionMs),
		}},
	}
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
