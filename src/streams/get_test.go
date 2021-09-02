/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package streams

import (
	"errors"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/eventstreams"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/path"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	es "github.com/IBM/event-streams-go-sdk-generator/build/generated"
	"github.com/golang/mock/gomock"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
)

var tenantId1 = "tenant1"
var tenantId2 = "tenant2"
var tenantId3 = "tenant3"
var streamId = "dataIntegrator1.qualifier123"
var streamIdNoQualifier = "dataIntegrator1"
var tenant1WithQualifier = strings.Join([]string{tenantId1, streamId}, ".")
var tenant1NoQualifier = strings.Join([]string{tenantId1, streamIdNoQualifier}, ".")
var tenant2WithQualifier = strings.Join([]string{tenantId2, streamId}, ".")
var tenant2NoQualifier = strings.Join([]string{tenantId2, streamIdNoQualifier}, ".")

func TestGet(t *testing.T) {
	os.Setenv(response.EnvOwActivationId, "activation123")

	tenant1Args := map[string]interface{}{
		path.ParamOwPath: fmt.Sprintf("/hri/tenants/%s/streams", tenantId1),
	}
	tenant3Args := map[string]interface{}{
		path.ParamOwPath: fmt.Sprintf("/hri/tenants/%s/streams", tenantId3),
	}
	missingPathArgs := map[string]interface{}{}
	missingTenantArgs := map[string]interface{}{
		path.ParamOwPath: fmt.Sprintf("/hri/tenants"),
	}

	goodRequestStreams := []map[string]interface{}{
		{param.StreamId: streamId},
		{param.StreamId: streamIdNoQualifier},
	}

	testCases := []struct {
		name              string
		args              map[string]interface{}
		validatorResponse map[string]interface{}
		mockError         error
		mockResponse      *http.Response
		expected          map[string]interface{}
	}{
		{
			name: "missing-path",
			args: missingPathArgs,
			expected: response.Error(
				http.StatusBadRequest,
				"Required parameter '__ow_path' is missing"),
		},
		{
			name:     "missing-tenant-id",
			args:     missingTenantArgs,
			expected: response.Error(http.StatusBadRequest, "The path is shorter than the requested path parameter; path: [ hri tenants], requested index: 3"),
		},
		{
			name:         "conn-error",
			args:         tenant1Args,
			mockError:    errors.New(kafkaConnectionMessage),
			mockResponse: &StatusUnprocessableEntity,
			expected:     response.Error(http.StatusInternalServerError, kafkaConnectionMessage),
		},
		{
			name:         "not-authorized",
			args:         tenant1Args,
			mockError:    errors.New(forbiddenMessage),
			mockResponse: &StatusForbidden,
			expected:     response.Error(http.StatusUnauthorized, eventstreams.UnauthorizedMsg),
		},
		{
			name:         "missing-auth-header",
			args:         tenant1Args,
			mockError:    errors.New(unauthorizedMessage),
			mockResponse: &StatusUnauthorized,
			expected:     response.Error(http.StatusUnauthorized, eventstreams.MissingHeaderMsg),
		},
		{
			name:         "tenant1-good-request",
			args:         tenant1Args,
			mockResponse: &http.Response{StatusCode: 200},
			expected:     response.Success(http.StatusOK, map[string]interface{}{"results": goodRequestStreams}),
		},
		{
			name:         "tenant-not-found",
			args:         tenant3Args,
			mockResponse: &http.Response{StatusCode: 200},
			expected:     response.Success(http.StatusOK, map[string]interface{}{"results": []map[string]interface{}{}}),
		},
	}

	for _, tc := range testCases {

		controller := gomock.NewController(t)
		defer controller.Finish()
		mockService := test.NewMockService(controller)

		//Mock return topics for all tenants, GetById should return only the unique stream names for the specified tenant
		topicDetails := []es.TopicDetail{
			{Name: eventstreams.TopicPrefix + tenant1WithQualifier + eventstreams.InSuffix},
			{Name: eventstreams.TopicPrefix + tenant1WithQualifier + eventstreams.NotificationSuffix},
			{Name: eventstreams.TopicPrefix + tenant1NoQualifier + eventstreams.InSuffix},
			{Name: eventstreams.TopicPrefix + tenant1NoQualifier + eventstreams.NotificationSuffix},
			{Name: eventstreams.TopicPrefix + tenant2WithQualifier + eventstreams.InSuffix},
			{Name: eventstreams.TopicPrefix + tenant2WithQualifier + eventstreams.NotificationSuffix},
			{Name: eventstreams.TopicPrefix + tenant2NoQualifier + eventstreams.InSuffix},
			{Name: eventstreams.TopicPrefix + tenant2NoQualifier + eventstreams.NotificationSuffix},
		}
		mockService.
			EXPECT().
			ListTopics(gomock.Any(), &es.ListTopicsOpts{}).
			Return(topicDetails, tc.mockResponse, tc.mockError).
			MaxTimes(1)

		t.Run(tc.name, func(t *testing.T) {
			if actual := Get(tc.args, mockService); !reflect.DeepEqual(tc.expected, actual) {
				t.Error(fmt.Sprintf("Expected: [%v], actual: [%v]", tc.expected, actual))
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
		topics   []es.TopicDetail
		tenantId string
		expected []map[string]interface{}
	}{
		{
			name:     "no-partitions",
			topics:   []es.TopicDetail{},
			tenantId: tenantId1,
			expected: []map[string]interface{}{},
		},
		{
			name: "with-optional-qualifier",
			topics: []es.TopicDetail{
				{Name: eventstreams.TopicPrefix + tenant1WithQualifier + eventstreams.InSuffix},
				{Name: eventstreams.TopicPrefix + tenant1WithQualifier + eventstreams.NotificationSuffix},
				{Name: eventstreams.TopicPrefix + tenant1WithQualifier + eventstreams.OutSuffix},
				{Name: eventstreams.TopicPrefix + tenant1WithQualifier + eventstreams.InvalidSuffix},
				{Name: eventstreams.TopicPrefix + tenant1NoQualifier + eventstreams.InSuffix},
				{Name: eventstreams.TopicPrefix + tenant1NoQualifier + eventstreams.NotificationSuffix},
				{Name: eventstreams.TopicPrefix + tenant1NoQualifier + eventstreams.OutSuffix},
				{Name: eventstreams.TopicPrefix + tenant1NoQualifier + eventstreams.InvalidSuffix},
				{Name: eventstreams.TopicPrefix + tenant2WithQualifier + eventstreams.InSuffix},
				{Name: eventstreams.TopicPrefix + tenant2WithQualifier + eventstreams.NotificationSuffix},
				{Name: eventstreams.TopicPrefix + tenant2WithQualifier + eventstreams.OutSuffix},
				{Name: eventstreams.TopicPrefix + tenant2WithQualifier + eventstreams.InvalidSuffix},
				{Name: eventstreams.TopicPrefix + tenant2NoQualifier + eventstreams.InSuffix},
				{Name: eventstreams.TopicPrefix + tenant2NoQualifier + eventstreams.NotificationSuffix},
				{Name: eventstreams.TopicPrefix + tenant2NoQualifier + eventstreams.OutSuffix},
				{Name: eventstreams.TopicPrefix + tenant2NoQualifier + eventstreams.InvalidSuffix},
			},
			tenantId: tenantId1,
			expected: []map[string]interface{}{
				{param.StreamId: streamId},
				{param.StreamId: streamIdNoQualifier},
			},
		},
		{
			name: "with-optional-qualifier-in-only",
			topics: []es.TopicDetail{
				{Name: eventstreams.TopicPrefix + tenant1WithQualifier + eventstreams.InSuffix},
				{Name: eventstreams.TopicPrefix + tenant1NoQualifier + eventstreams.InSuffix},
				{Name: eventstreams.TopicPrefix + tenant2WithQualifier + eventstreams.InSuffix},
				{Name: eventstreams.TopicPrefix + tenant2NoQualifier + eventstreams.InSuffix},
			},
			tenantId: tenantId1,
			expected: []map[string]interface{}{
				{param.StreamId: streamId},
				{param.StreamId: streamIdNoQualifier},
			},
		},
		{
			name: "with-optional-qualifier-out-only",
			topics: []es.TopicDetail{
				{Name: eventstreams.TopicPrefix + tenant1WithQualifier + eventstreams.OutSuffix},
				{Name: eventstreams.TopicPrefix + tenant1NoQualifier + eventstreams.OutSuffix},
				{Name: eventstreams.TopicPrefix + tenant2WithQualifier + eventstreams.OutSuffix},
				{Name: eventstreams.TopicPrefix + tenant2NoQualifier + eventstreams.OutSuffix},
			},
			tenantId: tenantId1,
			expected: []map[string]interface{}{
				{param.StreamId: streamId},
				{param.StreamId: streamIdNoQualifier},
			},
		},
		{
			name: "with-optional-qualifier-invalid-only",
			topics: []es.TopicDetail{
				{Name: eventstreams.TopicPrefix + tenant1WithQualifier + eventstreams.InvalidSuffix},
				{Name: eventstreams.TopicPrefix + tenant1NoQualifier + eventstreams.InvalidSuffix},
				{Name: eventstreams.TopicPrefix + tenant2WithQualifier + eventstreams.InvalidSuffix},
				{Name: eventstreams.TopicPrefix + tenant2NoQualifier + eventstreams.InvalidSuffix},
			},
			tenantId: tenantId1,
			expected: []map[string]interface{}{
				{param.StreamId: streamId},
				{param.StreamId: streamIdNoQualifier},
			},
		},
		{
			name: "with-optional-qualifier-notification-only",
			topics: []es.TopicDetail{
				{Name: eventstreams.TopicPrefix + tenant1WithQualifier + eventstreams.NotificationSuffix},
				{Name: eventstreams.TopicPrefix + tenant1NoQualifier + eventstreams.NotificationSuffix},
				{Name: eventstreams.TopicPrefix + tenant2WithQualifier + eventstreams.NotificationSuffix},
				{Name: eventstreams.TopicPrefix + tenant2NoQualifier + eventstreams.NotificationSuffix},
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
			topics: []es.TopicDetail{
				{Name: eventstreams.TopicPrefix + tenant1WithQualifier + eventstreams.InSuffix},
				{Name: eventstreams.TopicPrefix + tenant1WithQualifier + eventstreams.NotificationSuffix},
				{Name: eventstreams.TopicPrefix + tenant1WithQualifier + eventstreams.OutSuffix},
				{Name: eventstreams.TopicPrefix + tenant1WithQualifier + eventstreams.InvalidSuffix},
				{Name: eventstreams.TopicPrefix + tenant1NoQualifier + eventstreams.InSuffix},
				{Name: eventstreams.TopicPrefix + tenant1NoQualifier + eventstreams.NotificationSuffix},
				{Name: eventstreams.TopicPrefix + tenant1NoQualifier + eventstreams.OutSuffix},
				{Name: eventstreams.TopicPrefix + tenant1NoQualifier + eventstreams.InvalidSuffix},
				{Name: eventstreams.TopicPrefix + tenant0WithQualifier + eventstreams.InSuffix},
				{Name: eventstreams.TopicPrefix + tenant0WithQualifier + eventstreams.NotificationSuffix},
				{Name: eventstreams.TopicPrefix + tenant0WithQualifier + eventstreams.OutSuffix},
				{Name: eventstreams.TopicPrefix + tenant0WithQualifier + eventstreams.InvalidSuffix},
			},
			tenantId: tenantId0,
			expected: []map[string]interface{}{{param.StreamId: streamId}},
		},
		{
			name: "qualifier-with-extra-period",
			topics: []es.TopicDetail{
				{Name: eventstreams.TopicPrefix + tenant1ExtraPeriod + eventstreams.InSuffix},
				{Name: eventstreams.TopicPrefix + tenant1ExtraPeriod + eventstreams.NotificationSuffix},
				{Name: eventstreams.TopicPrefix + tenant1ExtraPeriod + eventstreams.OutSuffix},
				{Name: eventstreams.TopicPrefix + tenant1ExtraPeriod + eventstreams.InvalidSuffix},
				{Name: eventstreams.TopicPrefix + tenant0WithQualifier + eventstreams.InSuffix},
				{Name: eventstreams.TopicPrefix + tenant0WithQualifier + eventstreams.NotificationSuffix},
				{Name: eventstreams.TopicPrefix + tenant0WithQualifier + eventstreams.OutSuffix},
				{Name: eventstreams.TopicPrefix + tenant0WithQualifier + eventstreams.InvalidSuffix},
			},
			tenantId: tenantId1,
			expected: []map[string]interface{}{{param.StreamId: streamIdExtraPeriod}},
		},
		{
			name: "tenant-not-found",
			topics: []es.TopicDetail{
				{Name: eventstreams.TopicPrefix + tenant1WithQualifier + eventstreams.InSuffix},
				{Name: eventstreams.TopicPrefix + tenant1WithQualifier + eventstreams.NotificationSuffix},
				{Name: eventstreams.TopicPrefix + tenant1WithQualifier + eventstreams.OutSuffix},
				{Name: eventstreams.TopicPrefix + tenant1WithQualifier + eventstreams.InvalidSuffix},
				{Name: eventstreams.TopicPrefix + tenant1NoQualifier + eventstreams.InSuffix},
				{Name: eventstreams.TopicPrefix + tenant1NoQualifier + eventstreams.NotificationSuffix},
				{Name: eventstreams.TopicPrefix + tenant1NoQualifier + eventstreams.OutSuffix},
				{Name: eventstreams.TopicPrefix + tenant1NoQualifier + eventstreams.InvalidSuffix},
				{Name: eventstreams.TopicPrefix + tenant2WithQualifier + eventstreams.InSuffix},
				{Name: eventstreams.TopicPrefix + tenant2WithQualifier + eventstreams.NotificationSuffix},
				{Name: eventstreams.TopicPrefix + tenant2WithQualifier + eventstreams.OutSuffix},
				{Name: eventstreams.TopicPrefix + tenant2WithQualifier + eventstreams.InvalidSuffix},
				{Name: eventstreams.TopicPrefix + tenant2NoQualifier + eventstreams.InSuffix},
				{Name: eventstreams.TopicPrefix + tenant2NoQualifier + eventstreams.NotificationSuffix},
				{Name: eventstreams.TopicPrefix + tenant2NoQualifier + eventstreams.OutSuffix},
				{Name: eventstreams.TopicPrefix + tenant2NoQualifier + eventstreams.InvalidSuffix},
			},
			tenantId: tenantId0,
			expected: []map[string]interface{}{},
		},
		{
			name: "ignore-invalid-prefix",
			topics: []es.TopicDetail{
				{Name: tenant1WithQualifier + eventstreams.InSuffix},
				{Name: tenant1WithQualifier + eventstreams.NotificationSuffix},
				{Name: tenant1WithQualifier + eventstreams.OutSuffix},
				{Name: tenant1WithQualifier + eventstreams.InvalidSuffix},
				{Name: "bad-prefix" + tenant1WithQualifier + eventstreams.InSuffix},
				{Name: "bad-prefix" + tenant1WithQualifier + eventstreams.NotificationSuffix},
				{Name: "bad-prefix" + tenant1WithQualifier + eventstreams.OutSuffix},
				{Name: "bad-prefix" + tenant1WithQualifier + eventstreams.InvalidSuffix},
				{Name: eventstreams.TopicPrefix + tenant1NoQualifier + eventstreams.InSuffix},
				{Name: eventstreams.TopicPrefix + tenant1NoQualifier + eventstreams.NotificationSuffix},
				{Name: eventstreams.TopicPrefix + tenant1NoQualifier + eventstreams.OutSuffix},
				{Name: eventstreams.TopicPrefix + tenant1NoQualifier + eventstreams.InvalidSuffix},
			},
			tenantId: tenantId1,
			expected: []map[string]interface{}{{param.StreamId: streamIdNoQualifier}},
		},
		{
			name: "ignore-invalid-suffix",
			topics: []es.TopicDetail{
				{Name: eventstreams.TopicPrefix + tenant1WithQualifier},
				{Name: eventstreams.TopicPrefix + tenant1WithQualifier},
				{Name: eventstreams.TopicPrefix + tenant1WithQualifier + "bad-suffix"},
				{Name: eventstreams.TopicPrefix + tenant1WithQualifier + "bad-suffix"},
				{Name: eventstreams.TopicPrefix + tenant1NoQualifier + eventstreams.InSuffix},
				{Name: eventstreams.TopicPrefix + tenant1NoQualifier + eventstreams.NotificationSuffix},
				{Name: eventstreams.TopicPrefix + tenant1NoQualifier + eventstreams.OutSuffix},
				{Name: eventstreams.TopicPrefix + tenant1NoQualifier + eventstreams.InvalidSuffix},
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
