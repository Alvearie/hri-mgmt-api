/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package streams

import (
	"context"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	cfk "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/golang/mock/gomock"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"
)

const (
	topicNotFoundMessage string = "Broker: Unknown topic or partition"
	unauthorizedMessage  string = "Authorization failed."
)

func TestDelete(t *testing.T) {
	logwrapper.Initialize("error", os.Stdout)
	var requestId = "reqIdZz05"

	testCases := []struct {
		name               string
		topics             []string
		deleteResults      []cfk.TopicResult
		deleteError        error
		expectedReturnCode int
		expectedError      error
	}{
		{
			name:   "happy path",
			topics: []string{"in", "out", "notification", "invalid"},
			deleteResults: []cfk.TopicResult{
				{Topic: "in", Error: cfk.NewError(cfk.ErrNoError, "", false)},
				{Topic: "out", Error: cfk.NewError(cfk.ErrNoError, "", false)},
				{Topic: "notification", Error: cfk.NewError(cfk.ErrNoError, "", false)},
				{Topic: "invalid", Error: cfk.NewError(cfk.ErrNoError, "", false)},
			},
			expectedReturnCode: http.StatusOK,
		},
		{
			name:   "not-authorized",
			topics: []string{"in"},
			deleteResults: []cfk.TopicResult{
				{Topic: "in", Error: cfk.NewError(cfk.ErrTopicAuthorizationFailed, unauthorizedMessage, false)},
			},
			expectedReturnCode: http.StatusUnauthorized,
			expectedError:      fmt.Errorf("Unable to delete topic \"in\": " + unauthorizedMessage),
		},
		{
			name:               "timed-out",
			topics:             []string{"in"},
			deleteResults:      []cfk.TopicResult{},
			deleteError:        cfk.NewError(cfk.ErrTimedOut, "Failed while waiting for controller: Local: Timed out", false),
			expectedReturnCode: http.StatusInternalServerError,
			expectedError:      fmt.Errorf("unexpected error deleting Kafka topics: %w", cfk.NewError(cfk.ErrTimedOut, "Failed while waiting for controller: Local: Timed out", false)),
		},
		{
			name:   "topic-not-found",
			topics: []string{"in"},
			deleteResults: []cfk.TopicResult{
				{Topic: "in", Error: cfk.NewError(cfk.ErrUnknownTopicOrPart, topicNotFoundMessage, false)},
			},
			expectedReturnCode: http.StatusNotFound,
			expectedError:      fmt.Errorf("Unable to delete topic \"in\": " + topicNotFoundMessage),
		},
		{
			name:   "multiple-delete-errors",
			topics: []string{"in", "out", "notification", "invalid"},
			deleteResults: []cfk.TopicResult{
				{Topic: "in"},
				{Topic: "out", Error: cfk.NewError(cfk.ErrTimedOut, "Failed while waiting for controller: Local: Timed out", false)},
				{Topic: "notification", Error: cfk.NewError(cfk.ErrUnknownTopicOrPart, topicNotFoundMessage, false)},
				{Topic: "invalid", Error: cfk.NewError(cfk.ErrTimedOut, "Failed while waiting for controller: Local: Timed out", false)},
			},
			expectedReturnCode: http.StatusInternalServerError,
			expectedError: fmt.Errorf(`Unable to delete topic "out": Failed while waiting for controller: Local: Timed out` + "\n" +
				`Unable to delete topic "notification": ` + topicNotFoundMessage + "\n" +
				`Unable to delete topic "invalid": Failed while waiting for controller: Local: Timed out`),
		},
	}

	for _, tc := range testCases {
		controller := gomock.NewController(t)
		defer controller.Finish()
		mockService := test.NewMockService(controller)

		mockService.
			EXPECT().
			DeleteTopics(context.Background(), tc.topics, cfk.SetAdminRequestTimeout(time.Second*10)).
			Return(tc.deleteResults, tc.deleteError).
			MaxTimes(1)

		actualCode, actualErrMsg := Delete(requestId, tc.topics, mockService)

		t.Run(tc.name, func(t *testing.T) {
			if actualCode != tc.expectedReturnCode || !reflect.DeepEqual(tc.expectedError, actualErrMsg) {
				t.Errorf("Stream-Delete()\n   actual: %v,%v\n expected: %v,%v",
					actualCode, actualErrMsg, tc.expectedReturnCode, tc.expectedError)
			}
		})
	}
}
