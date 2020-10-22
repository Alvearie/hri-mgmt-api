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
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/path"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	es "github.com/IBM/event-streams-go-sdk-generator/build/generated"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

const (
	numPartitions               float64 = 1
	retentionMs                 float64 = 86400000
	topicAlreadyExistsMessage   string  = "topic already exists"
	invalidCleanupPolicyMessage string  = "invalid cleanup policy"
	forbiddenMessage            string  = "forbidden"
	unauthorizedMessage         string  = "unauthorized"
	kafkaConnectionMessage              = "Unable to connect to Kafka"
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

func TestCreate(t *testing.T) {
	os.Setenv(response.EnvOwActivationId, "activation123")

	tenantId := "tenant123"
	streamId1 := "data-integrator123.qualifier_123"
	streamId2 := "data-integrator123"
	baseTopicName := strings.Join([]string{tenantId, streamId1}, ".")
	baseTopicNameNoQualifier := strings.Join([]string{tenantId, streamId2}, ".")

	validArgs := map[string]interface{}{
		path.ParamOwPath:    fmt.Sprintf("/hri/tenants/%s/streams/%s", tenantId, streamId1),
		param.NumPartitions: numPartitions,
		param.RetentionMs:   retentionMs,
	}
	validArgsNoQualifier := map[string]interface{}{
		path.ParamOwPath:    fmt.Sprintf("/hri/tenants/%s/streams/%s", tenantId, streamId2),
		param.NumPartitions: numPartitions,
		param.RetentionMs:   retentionMs,
	}
	missingPathArgs := map[string]interface{}{
		param.NumPartitions: numPartitions,
		param.RetentionMs:   retentionMs,
	}
	missingTenantArgs := map[string]interface{}{
		path.ParamOwPath:    fmt.Sprintf("/hri/tenants"),
		param.NumPartitions: numPartitions,
		param.RetentionMs:   retentionMs,
	}
	missingStreamArgs := map[string]interface{}{
		path.ParamOwPath:    fmt.Sprintf("/hri/tenants/%s/streams", tenantId),
		param.NumPartitions: numPartitions,
		param.RetentionMs:   retentionMs,
	}

	badParamResponse := map[string]interface{}{"bad": "param"}

	testCases := []struct {
		name                   string
		args                   map[string]interface{}
		validatorResponse      map[string]interface{}
		modelInError           *es.ModelError
		modelNotificationError *es.ModelError
		mockResponse           *http.Response
		deleteError            error
		expectedTopic          string
		expected               map[string]interface{}
	}{
		{
			name:              "bad-param",
			args:              validArgs,
			validatorResponse: badParamResponse,
			expected:          badParamResponse,
		},
		{
			name:     "missing-path",
			args:     missingPathArgs,
			expected: response.Error(http.StatusBadRequest, "Required parameter '__ow_path' is missing"),
		},
		{
			name:     "missing-tenant-id",
			args:     missingTenantArgs,
			expected: response.Error(http.StatusBadRequest, "The path is shorter than the requested path parameter; path: [ hri tenants], requested index: 3"),
		},
		{
			name:     "missing-stream-id",
			args:     missingStreamArgs,
			expected: response.Error(http.StatusBadRequest, "The path is shorter than the requested path parameter; path: [ hri tenants tenant123 streams], requested index: 5"),
		},
		{
			name: "invalid-tenant-id capital",
			args: map[string]interface{}{
				path.ParamOwPath:    fmt.Sprintf("/hri/tenants/tenantId/streams/stream-id"),
				param.NumPartitions: numPartitions,
				param.RetentionMs:   retentionMs,
			},
			expected: response.Error(http.StatusBadRequest, "TenantId: tenantId must be lower-case alpha-numeric, '-', or '_'. 'I' is not allowed."),
		},
		{
			name: "invalid-stream-id capital",
			args: map[string]interface{}{
				path.ParamOwPath:    fmt.Sprintf("/hri/tenants/%s/streams/streamId", tenantId),
				param.NumPartitions: numPartitions,
				param.RetentionMs:   retentionMs,
			},
			expected: response.Error(http.StatusBadRequest, "StreamId: streamId must be lower-case alpha-numeric, '-', or '_', and no more than one '.'. 'I' is not allowed."),
		},
		{
			name:          "not-authorized",
			args:          validArgs,
			modelInError:  &es.ModelError{},
			mockResponse:  &StatusForbidden,
			expected:      response.Error(http.StatusUnauthorized, eventstreams.UnauthorizedMsg),
			expectedTopic: baseTopicName,
		},
		{
			name:          "missing-auth-header",
			args:          validArgs,
			modelInError:  &es.ModelError{},
			mockResponse:  &StatusUnauthorized,
			expected:      response.Error(http.StatusUnauthorized, eventstreams.MissingHeaderMsg),
			expectedTopic: baseTopicName,
		},
		{
			name:          "in-topic-already-exists",
			args:          validArgs,
			modelInError:  &TopicAlreadyExistsError,
			mockResponse:  &StatusUnprocessableEntity,
			expected:      response.Error(http.StatusConflict, topicAlreadyExistsMessage),
			expectedTopic: baseTopicName,
		},
		{
			name:                   "notification-topic-already-exists",
			args:                   validArgs,
			modelNotificationError: &TopicAlreadyExistsError,
			mockResponse:           &StatusUnprocessableEntity,
			expected:               response.Error(http.StatusConflict, topicAlreadyExistsMessage),
			expectedTopic:          baseTopicName,
		},
		{
			name:                   "notification-fail-and-delete-fail",
			args:                   validArgs,
			modelNotificationError: &TopicAlreadyExistsError,
			mockResponse:           &StatusUnprocessableEntity,
			deleteError:            errors.New("failed to delete in topic"),
			expected:               response.Error(http.StatusConflict, topicAlreadyExistsMessage+fmt.Sprintf(cleanupFailureMsg, eventstreams.TopicPrefix+baseTopicName+eventstreams.InSuffix, "failed to delete in topic")),
			expectedTopic:          baseTopicName,
		},
		{
			name:          "invalid-cleanup-policy",
			args:          validArgs,
			modelInError:  &invalidCleanupPolicyError,
			mockResponse:  &StatusUnprocessableEntity,
			expected:      response.Error(http.StatusBadRequest, invalidCleanupPolicyMessage),
			expectedTopic: baseTopicName,
		},
		{
			name:          "in-conn-error",
			args:          validArgs,
			modelInError:  &OtherError,
			mockResponse:  &StatusUnprocessableEntity,
			expected:      response.Error(http.StatusInternalServerError, kafkaConnectionMessage),
			expectedTopic: baseTopicName,
		},
		{
			name:                   "notification-conn-error",
			args:                   validArgs,
			modelNotificationError: &OtherError,
			mockResponse:           &StatusUnprocessableEntity,
			expected:               response.Error(http.StatusInternalServerError, kafkaConnectionMessage),
			expectedTopic:          baseTopicName,
		},
		{
			name:          "good-request-qualifier",
			args:          validArgs,
			expected:      response.Success(http.StatusCreated, map[string]interface{}{param.StreamId: streamId1}),
			expectedTopic: baseTopicName,
		},
		{
			name:          "good-request-no-qualifier",
			args:          validArgsNoQualifier,
			expected:      response.Success(http.StatusCreated, map[string]interface{}{param.StreamId: streamId2}),
			expectedTopic: baseTopicNameNoQualifier,
		},
	}

	for _, tc := range testCases {
		validator := test.FakeValidator{
			T: t,
			Required: []param.Info{
				{param.NumPartitions, reflect.Float64},
				{param.RetentionMs, reflect.Float64},
			},
			Optional: []param.Info{
				{param.CleanupPolicy, reflect.String},
				{param.RetentionBytes, reflect.Float64},
				{param.SegmentMs, reflect.Float64},
				{param.SegmentBytes, reflect.Float64},
				{param.SegmentIndexBytes, reflect.Float64},
			},
			Response: tc.validatorResponse,
		}

		controller := gomock.NewController(t)
		defer controller.Finish()
		mockService := test.NewMockService(controller)

		var mockInErr error
		var mockNotificationErr error
		if tc.modelInError != nil {
			mockInErr = errors.New(tc.modelInError.Message)
		}
		if tc.modelNotificationError != nil {
			mockNotificationErr = errors.New(tc.modelNotificationError.Message)
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
			DeleteTopic(context.Background(), eventstreams.TopicPrefix+tc.expectedTopic+eventstreams.InSuffix).
			Return(nil, nil, tc.deleteError).
			MaxTimes(1)

		t.Run(tc.name, func(t *testing.T) {
			if actual := Create(tc.args, validator, mockService); !reflect.DeepEqual(tc.expected, actual) {
				t.Error(fmt.Sprintf("Expected: [%v], actual: [%v]", tc.expected, actual))
			}
		})
	}
}

func TestSetUpTopicConfigs(t *testing.T) {
	var configsValue float64 = 10485760
	expectedConfigs := []es.ConfigCreate{
		{
			Name:  "retention.ms",
			Value: strconv.Itoa(int(retentionMs)),
		},
		{
			Name:  "retention.bytes",
			Value: strconv.Itoa(int(configsValue)),
		},
		{
			Name:  "cleanup.policy",
			Value: "compact",
		},
		{
			Name:  "segment.ms",
			Value: strconv.Itoa(int(configsValue)),
		},
		{
			Name:  "segment.bytes",
			Value: strconv.Itoa(int(configsValue)),
		},
		{
			Name:  "segment.index.bytes",
			Value: strconv.Itoa(int(configsValue)),
		},
	}
	validArgs := map[string]interface{}{
		param.RetentionMs:       retentionMs,
		param.RetentionBytes:    configsValue,
		param.CleanupPolicy:     "compact",
		param.SegmentMs:         configsValue,
		param.SegmentBytes:      configsValue,
		param.SegmentIndexBytes: configsValue,
	}
	configs := setUpTopicConfigs(validArgs)
	assert.Equal(t, expectedConfigs, configs)
}

func TestSetUpTopicConfigsNoExtras(t *testing.T) {
	expectedConfigs := []es.ConfigCreate{
		{
			Name:  "retention.ms",
			Value: strconv.Itoa(int(retentionMs)),
		},
	}
	validArgs := map[string]interface{}{param.RetentionMs: retentionMs}
	configs := setUpTopicConfigs(validArgs)
	assert.Equal(t, expectedConfigs, configs)
}

func getTestTopicRequest(streamName string, topicSuffix string) es.TopicCreateRequest {
	return es.TopicCreateRequest{
		Name:           eventstreams.TopicPrefix + streamName + topicSuffix,
		PartitionCount: int64(numPartitions),
		Configs: []es.ConfigCreate{{
			Name:  "retention.ms",
			Value: strconv.Itoa(int(retentionMs)),
		}},
	}
}
