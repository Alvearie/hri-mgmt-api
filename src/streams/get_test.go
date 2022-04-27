/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package streams

import (
	"errors"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/kafka"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	cfk "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/golang/mock/gomock"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
)

var tenantId1 = "tenant1"
var tenantId2 = "tenant2"

var streamId = "dataIntegrator1.qualifier123"
var streamIdNoQualifier = "dataIntegrator1"
var tenant1WithQualifier = strings.Join([]string{tenantId1, streamId}, ".")
var tenant1NoQualifier = strings.Join([]string{tenantId1, streamIdNoQualifier}, ".")
var tenant2WithQualifier = strings.Join([]string{tenantId2, streamId}, ".")
var tenant2NoQualifier = strings.Join([]string{tenantId2, streamIdNoQualifier}, ".")

func TestGet(t *testing.T) {
	logwrapper.Initialize("error", os.Stdout)

	goodRequestStreams := []map[string]interface{}{
		{param.StreamId: streamId},
		{param.StreamId: streamIdNoQualifier},
	}
	emptyStreamsResults := []map[string]interface{}{}
	var requestId = "req78vQ22dp9"
	var validTenant1 = "tenantSnoopDogg"
	var validTenant1WithQualifier = strings.Join([]string{validTenant1, streamId}, ".")
	var validTenant1NoQualifier = strings.Join([]string{validTenant1, streamIdNoQualifier}, ".")

	testCases := []struct {
		name         string
		tenantId     string
		mockError    error
		mockResponse *http.Response
		expectedCode int
		expectedBody interface{}
	}{
		{
			name:         "conn-error",
			tenantId:     validTenant1,
			mockError:    errors.New(kafkaConnectionMessage),
			mockResponse: &StatusUnprocessableEntity,
			expectedCode: http.StatusInternalServerError,
			expectedBody: response.NewErrorDetail(requestId, kafkaConnectionMessage),
		},
		{
			name:         "not-authorized",
			tenantId:     validTenant1,
			mockError:    errors.New(forbiddenMessage),
			mockResponse: &StatusForbidden,
			expectedCode: http.StatusUnauthorized,
			expectedBody: response.NewErrorDetail(requestId, kafka.UnauthorizedMsg),
		},
		{
			name:         "missing-auth-header",
			tenantId:     validTenant1,
			mockError:    errors.New(unauthorizedMessage),
			mockResponse: &StatusUnauthorized,
			expectedCode: http.StatusUnauthorized,
			expectedBody: response.NewErrorDetail(requestId, kafka.MissingHeaderMsg),
		},
		{
			name:         "happy-path-good-request",
			tenantId:     validTenant1,
			mockResponse: &http.Response{StatusCode: 200},
			expectedCode: http.StatusOK,
			expectedBody: map[string]interface{}{"results": goodRequestStreams},
		},
		{
			name:         "happy-path-tenant-not-found",
			tenantId:     "tenant_NO_STREAMS_FOUND",
			mockResponse: &http.Response{StatusCode: 200},
			expectedCode: http.StatusOK,
			expectedBody: map[string]interface{}{"results": emptyStreamsResults},
		},
	}

	for _, tc := range testCases {
		controller := gomock.NewController(t)
		defer controller.Finish()
		mockService := test.NewMockService(controller)

		//Mock return topics for all tenants, GetById should return only the unique stream names for the specified tenant
		topicDetails := &cfk.Metadata{}
		topicDetails.Topics[kafka.TopicPrefix+validTenant1WithQualifier+kafka.InSuffix] = cfk.TopicMetadata{}
		topicDetails.Topics[kafka.TopicPrefix+validTenant1WithQualifier+kafka.NotificationSuffix] = cfk.TopicMetadata{}
		topicDetails.Topics[kafka.TopicPrefix+validTenant1NoQualifier+kafka.InSuffix] = cfk.TopicMetadata{}
		topicDetails.Topics[kafka.TopicPrefix+validTenant1NoQualifier+kafka.NotificationSuffix] = cfk.TopicMetadata{}

		mockService.
			EXPECT().
			GetMetadata(nil, true, 10000).
			Return(topicDetails, tc.mockResponse, tc.mockError).
			MaxTimes(1)

		t.Run(tc.name, func(t *testing.T) {
			actualCode, actualBody := Get(requestId, tc.tenantId, mockService)
			if actualCode != tc.expectedCode || !reflect.DeepEqual(tc.expectedBody, actualBody) {
				t.Errorf("Streams-Get() \n actual: %v,%v\n expected: %v,%v",
					actualCode, actualBody, tc.expectedCode, tc.expectedBody)
			}
		})
	}
}

/*
	Four things to note on the expected returned stream names:
	1. We should only get stream names for the specified tenant
	2. We ignore stream names corresponding to topics without valid prefixes and suffixes (ingest and in/notification)
	3. We should only get unique stream names for a tenant/dataIntegrator pairing, even though there are multiple topics
	4. "dataIntegrator1.qualifier1" is a unique stream from "dataIntegrator1" and both should be returned
*/
func TestGetStreamNames(t *testing.T) {
	tenantId0 := "tenant"
	tenant0WithQualifier := strings.Join([]string{tenantId0, streamId}, ".")
	streamIdExtraPeriod := "dataIntegrator1..qualifier"
	tenant1ExtraPeriod := strings.Join([]string{tenantId1, streamIdExtraPeriod}, ".")

	testCases := []struct {
		name     string
		topics   []string
		tenantId string
		expected []map[string]interface{}
	}{
		{
			name:     "no-partitions",
			topics:   []string{},
			tenantId: tenantId1,
			expected: []map[string]interface{}{},
		},
		{
			name: "with-optional-qualifier",
			topics: []string{
				kafka.TopicPrefix + tenant1WithQualifier + kafka.InSuffix,
				kafka.TopicPrefix + tenant1WithQualifier + kafka.NotificationSuffix,
				kafka.TopicPrefix + tenant1WithQualifier + kafka.OutSuffix,
				kafka.TopicPrefix + tenant1WithQualifier + kafka.InvalidSuffix,
				kafka.TopicPrefix + tenant1NoQualifier + kafka.InSuffix,
				kafka.TopicPrefix + tenant1NoQualifier + kafka.NotificationSuffix,
				kafka.TopicPrefix + tenant1NoQualifier + kafka.OutSuffix,
				kafka.TopicPrefix + tenant1NoQualifier + kafka.InvalidSuffix,
				kafka.TopicPrefix + tenant2WithQualifier + kafka.InSuffix,
				kafka.TopicPrefix + tenant2WithQualifier + kafka.NotificationSuffix,
				kafka.TopicPrefix + tenant2WithQualifier + kafka.OutSuffix,
				kafka.TopicPrefix + tenant2WithQualifier + kafka.InvalidSuffix,
				kafka.TopicPrefix + tenant2NoQualifier + kafka.InSuffix,
				kafka.TopicPrefix + tenant2NoQualifier + kafka.NotificationSuffix,
				kafka.TopicPrefix + tenant2NoQualifier + kafka.OutSuffix,
				kafka.TopicPrefix + tenant2NoQualifier + kafka.InvalidSuffix,
			},
			tenantId: tenantId1,
			expected: []map[string]interface{}{
				{param.StreamId: streamId},
				{param.StreamId: streamIdNoQualifier},
			},
		},
		{
			name: "with-optional-qualifier-in-only",
			topics: []string{
				kafka.TopicPrefix + tenant1WithQualifier + kafka.InSuffix,
				kafka.TopicPrefix + tenant1NoQualifier + kafka.InSuffix,
				kafka.TopicPrefix + tenant2WithQualifier + kafka.InSuffix,
				kafka.TopicPrefix + tenant2NoQualifier + kafka.InSuffix,
			},
			tenantId: tenantId1,
			expected: []map[string]interface{}{
				{param.StreamId: streamId},
				{param.StreamId: streamIdNoQualifier},
			},
		},
		{
			name: "with-optional-qualifier-out-only",
			topics: []string{
				kafka.TopicPrefix + tenant1WithQualifier + kafka.OutSuffix,
				kafka.TopicPrefix + tenant1NoQualifier + kafka.OutSuffix,
				kafka.TopicPrefix + tenant2WithQualifier + kafka.OutSuffix,
				kafka.TopicPrefix + tenant2NoQualifier + kafka.OutSuffix,
			},
			tenantId: tenantId1,
			expected: []map[string]interface{}{
				{param.StreamId: streamId},
				{param.StreamId: streamIdNoQualifier},
			},
		},
		{
			name: "with-optional-qualifier-invalid-only",
			topics: []string{
				kafka.TopicPrefix + tenant1WithQualifier + kafka.InvalidSuffix,
				kafka.TopicPrefix + tenant1NoQualifier + kafka.InvalidSuffix,
				kafka.TopicPrefix + tenant2WithQualifier + kafka.InvalidSuffix,
				kafka.TopicPrefix + tenant2NoQualifier + kafka.InvalidSuffix,
			},
			tenantId: tenantId1,
			expected: []map[string]interface{}{
				{param.StreamId: streamId},
				{param.StreamId: streamIdNoQualifier},
			},
		},
		{
			name: "with-optional-qualifier-notification-only",
			topics: []string{
				kafka.TopicPrefix + tenant1WithQualifier + kafka.NotificationSuffix,
				kafka.TopicPrefix + tenant1NoQualifier + kafka.NotificationSuffix,
				kafka.TopicPrefix + tenant2WithQualifier + kafka.NotificationSuffix,
				kafka.TopicPrefix + tenant2NoQualifier + kafka.NotificationSuffix,
			},
			tenantId: tenantId1,
			expected: []map[string]interface{}{
				{param.StreamId: streamId},
				{param.StreamId: streamIdNoQualifier},
			},
		},
		{
			//"tenant" and "tenant1" should be treated uniquely
			name: "similar-tenant-names",
			topics: []string{
				kafka.TopicPrefix + tenant1WithQualifier + kafka.InSuffix,
				kafka.TopicPrefix + tenant1WithQualifier + kafka.NotificationSuffix,
				kafka.TopicPrefix + tenant1WithQualifier + kafka.OutSuffix,
				kafka.TopicPrefix + tenant1WithQualifier + kafka.InvalidSuffix,
				kafka.TopicPrefix + tenant1NoQualifier + kafka.InSuffix,
				kafka.TopicPrefix + tenant1NoQualifier + kafka.NotificationSuffix,
				kafka.TopicPrefix + tenant1NoQualifier + kafka.OutSuffix,
				kafka.TopicPrefix + tenant1NoQualifier + kafka.InvalidSuffix,
				kafka.TopicPrefix + tenant0WithQualifier + kafka.InSuffix,
				kafka.TopicPrefix + tenant0WithQualifier + kafka.NotificationSuffix,
				kafka.TopicPrefix + tenant0WithQualifier + kafka.OutSuffix,
				kafka.TopicPrefix + tenant0WithQualifier + kafka.InvalidSuffix,
			},
			tenantId: tenantId0,
			expected: []map[string]interface{}{{param.StreamId: streamId}},
		},
		{
			name: "qualifier-with-extra-period",
			topics: []string{
				kafka.TopicPrefix + tenant1ExtraPeriod + kafka.InSuffix,
				kafka.TopicPrefix + tenant1ExtraPeriod + kafka.NotificationSuffix,
				kafka.TopicPrefix + tenant1ExtraPeriod + kafka.OutSuffix,
				kafka.TopicPrefix + tenant1ExtraPeriod + kafka.InvalidSuffix,
				kafka.TopicPrefix + tenant0WithQualifier + kafka.InSuffix,
				kafka.TopicPrefix + tenant0WithQualifier + kafka.NotificationSuffix,
				kafka.TopicPrefix + tenant0WithQualifier + kafka.OutSuffix,
				kafka.TopicPrefix + tenant0WithQualifier + kafka.InvalidSuffix,
			},
			tenantId: tenantId1,
			expected: []map[string]interface{}{{param.StreamId: streamIdExtraPeriod}},
		},
		{
			name: "tenant-not-found",
			topics: []string{
				kafka.TopicPrefix + tenant1WithQualifier + kafka.InSuffix,
				kafka.TopicPrefix + tenant1WithQualifier + kafka.NotificationSuffix,
				kafka.TopicPrefix + tenant1WithQualifier + kafka.OutSuffix,
				kafka.TopicPrefix + tenant1WithQualifier + kafka.InvalidSuffix,
				kafka.TopicPrefix + tenant1NoQualifier + kafka.InSuffix,
				kafka.TopicPrefix + tenant1NoQualifier + kafka.NotificationSuffix,
				kafka.TopicPrefix + tenant1NoQualifier + kafka.OutSuffix,
				kafka.TopicPrefix + tenant1NoQualifier + kafka.InvalidSuffix,
				kafka.TopicPrefix + tenant2WithQualifier + kafka.InSuffix,
				kafka.TopicPrefix + tenant2WithQualifier + kafka.NotificationSuffix,
				kafka.TopicPrefix + tenant2WithQualifier + kafka.OutSuffix,
				kafka.TopicPrefix + tenant2WithQualifier + kafka.InvalidSuffix,
				kafka.TopicPrefix + tenant2NoQualifier + kafka.InSuffix,
				kafka.TopicPrefix + tenant2NoQualifier + kafka.NotificationSuffix,
				kafka.TopicPrefix + tenant2NoQualifier + kafka.OutSuffix,
				kafka.TopicPrefix + tenant2NoQualifier + kafka.InvalidSuffix,
			},
			tenantId: tenantId0,
			expected: []map[string]interface{}{},
		},
		{
			name: "ignore-invalid-prefix",
			topics: []string{
				tenant1WithQualifier + kafka.InSuffix,
				tenant1WithQualifier + kafka.NotificationSuffix,
				tenant1WithQualifier + kafka.OutSuffix,
				tenant1WithQualifier + kafka.InvalidSuffix,
				"bad-prefix" + tenant1WithQualifier + kafka.InSuffix,
				"bad-prefix" + tenant1WithQualifier + kafka.NotificationSuffix,
				"bad-prefix" + tenant1WithQualifier + kafka.OutSuffix,
				"bad-prefix" + tenant1WithQualifier + kafka.InvalidSuffix,
				kafka.TopicPrefix + tenant1NoQualifier + kafka.InSuffix,
				kafka.TopicPrefix + tenant1NoQualifier + kafka.NotificationSuffix,
				kafka.TopicPrefix + tenant1NoQualifier + kafka.OutSuffix,
				kafka.TopicPrefix + tenant1NoQualifier + kafka.InvalidSuffix,
			},
			tenantId: tenantId1,
			expected: []map[string]interface{}{{param.StreamId: streamIdNoQualifier}},
		},
		{
			name: "ignore-invalid-suffix",
			topics: []string{
				kafka.TopicPrefix + tenant1WithQualifier,
				kafka.TopicPrefix + tenant1WithQualifier,
				kafka.TopicPrefix + tenant1WithQualifier + "bad-suffix",
				kafka.TopicPrefix + tenant1WithQualifier + "bad-suffix",
				kafka.TopicPrefix + tenant1NoQualifier + kafka.InSuffix,
				kafka.TopicPrefix + tenant1NoQualifier + kafka.NotificationSuffix,
				kafka.TopicPrefix + tenant1NoQualifier + kafka.OutSuffix,
				kafka.TopicPrefix + tenant1NoQualifier + kafka.InvalidSuffix,
			},
			tenantId: tenantId1,
			expected: []map[string]interface{}{{param.StreamId: streamIdNoQualifier}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if actual := GetStreamNames(tc.topics, tc.tenantId); !reflect.DeepEqual(tc.expected, actual) {
				t.Error(fmt.Sprintf("Expected: [%v], actual: [%v]", tc.expected, actual))
			}
		})
	}
}
