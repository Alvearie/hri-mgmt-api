/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package streams

import (
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/kafka"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/param"
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

var inTopic1WithQualifier = kafka.TopicPrefix + tenant1WithQualifier + kafka.InSuffix
var inTopic1NoQualifier = kafka.TopicPrefix + tenant1NoQualifier + kafka.InSuffix
var notificationTopic1WithQualifier = kafka.TopicPrefix + tenant1WithQualifier + kafka.NotificationSuffix
var notificationTopic1NoQualifier = kafka.TopicPrefix + tenant1NoQualifier + kafka.NotificationSuffix
var inTopic2WithQualifier = kafka.TopicPrefix + tenant2WithQualifier + kafka.InSuffix
var inTopic2NoQualifier = kafka.TopicPrefix + tenant2NoQualifier + kafka.InSuffix
var notificationTopic2WithQualifier = kafka.TopicPrefix + tenant2WithQualifier + kafka.NotificationSuffix
var notificationTopic2NoQualifier = kafka.TopicPrefix + tenant2NoQualifier + kafka.NotificationSuffix

var goodMetadata = cfk.Metadata{
	Topics: map[string]cfk.TopicMetadata{
		inTopic1WithQualifier:           cfk.TopicMetadata{},
		inTopic1NoQualifier:             cfk.TopicMetadata{},
		notificationTopic1WithQualifier: cfk.TopicMetadata{},
		notificationTopic1NoQualifier:   cfk.TopicMetadata{},
		inTopic2WithQualifier:           cfk.TopicMetadata{},
		inTopic2NoQualifier:             cfk.TopicMetadata{},
		notificationTopic2WithQualifier: cfk.TopicMetadata{},
		notificationTopic2NoQualifier:   cfk.TopicMetadata{},
	},
}

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

	//Mock return topics for all tenants, GetById should return only the unique stream names for the specified tenant
	validMetadata := cfk.Metadata{Topics: map[string]cfk.TopicMetadata{}}
	validMetadata.Topics[kafka.TopicPrefix+validTenant1WithQualifier+kafka.InSuffix] = cfk.TopicMetadata{}
	validMetadata.Topics[kafka.TopicPrefix+validTenant1WithQualifier+kafka.NotificationSuffix] = cfk.TopicMetadata{}
	validMetadata.Topics[kafka.TopicPrefix+validTenant1NoQualifier+kafka.InSuffix] = cfk.TopicMetadata{}
	validMetadata.Topics[kafka.TopicPrefix+validTenant1NoQualifier+kafka.NotificationSuffix] = cfk.TopicMetadata{}

	testCases := []struct {
		name         string
		tenantId     string
		mockReturn   *cfk.Metadata
		mockError    error
		expectedCode int
		expectedBody interface{}
	}{
		{
			name:         "not-authorized",
			tenantId:     validTenant1,
			mockError:    cfk.NewError(cfk.ErrTopicAuthorizationFailed, unauthorizedMessage, false),
			expectedCode: http.StatusUnauthorized,
			expectedBody: fmt.Errorf("error listing Kafka topics: %w", cfk.NewError(cfk.ErrTopicAuthorizationFailed, unauthorizedMessage, false)),
		},
		{
			name:         "timed-out",
			tenantId:     validTenant1,
			mockError:    cfk.NewError(cfk.ErrTimedOut, "Local: Broker transport failure", false),
			expectedCode: http.StatusInternalServerError,
			expectedBody: fmt.Errorf("error listing Kafka topics: %w", cfk.NewError(cfk.ErrTimedOut, "Local: Broker transport failure", false)),
		},
		{
			name:         "happy-path-good-request",
			tenantId:     validTenant1,
			mockReturn:   &validMetadata,
			expectedCode: http.StatusOK,
			expectedBody: map[string]interface{}{"results": goodRequestStreams},
		},
		{
			name:         "happy-path-tenant-not-found",
			mockReturn:   &cfk.Metadata{Topics: map[string]cfk.TopicMetadata{}},
			tenantId:     "tenant_NO_STREAMS_FOUND",
			expectedCode: http.StatusOK,
			expectedBody: map[string]interface{}{"results": emptyStreamsResults},
		},
	}

	for _, tc := range testCases {
		controller := gomock.NewController(t)
		defer controller.Finish()
		mockService := test.NewMockService(controller)

		mockService.
			EXPECT().
			GetMetadata(nil, true, 10000).
			Return(tc.mockReturn, tc.mockError).
			MaxTimes(1)

		t.Run(tc.name, func(t *testing.T) {
			actualCode, actualBody := Get(requestId, tc.tenantId, mockService)
			if actualCode != tc.expectedCode || !bodiesMatch(tc.expectedBody, actualBody) {
				t.Errorf("Streams-Get() \n actual: %v,%v\n expected: %v,%v",
					actualCode, actualBody, tc.expectedCode, tc.expectedBody)
			}
		})
	}
}

func bodiesMatch(b1 interface{}, b2 interface{}) bool {

	_, b1IsErr := b1.(error)
	_, b2IsErr := b2.(error)
	if b1IsErr && b2IsErr {
		return reflect.DeepEqual(b1, b2)
	} else if b1IsErr {
		return false
	}

	// Else both should be of format { "results" : [ { "id" : streamId1 } , { "id" : streamId2 } ... ] }
	// Order may not match, make sure streamIds are the same.
	b1results := b1.(map[string]interface{})["results"].([]map[string]interface{})
	b2results := b2.(map[string]interface{})["results"].([]map[string]interface{})
	if len(b1results) != len(b2results) {
		return false
	}

	// For each key/value in b1, if it's not in b2, return false.
	for _, ele1 := range b1results {
		var match = false
		for _, ele2 := range b2results {
			if reflect.DeepEqual(ele1, ele2) {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}

	return true
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
