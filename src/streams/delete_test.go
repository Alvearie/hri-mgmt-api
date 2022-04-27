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
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	cfk "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/golang/mock/gomock"
	"net/http"
	"os"
	"reflect"
	"testing"
)

const topicNotFoundMessage string = "Topic not found"

func TestDelete(t *testing.T) {
	logwrapper.Initialize("error", os.Stdout)
	var requestId = "reqIdZz05"

	var kafkaConnectionError = errors.New(kafkaConnectionMessage)
	var topicNotFoundError = errors.New(topicNotFoundMessage)

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
				cfk.TopicResult{Topic: "in"},
				cfk.TopicResult{Topic: "out"},
				cfk.TopicResult{Topic: "notification"},
				cfk.TopicResult{Topic: "invalid"},
			},
			expectedReturnCode: http.StatusOK,
		},
		{
			name:               "not-authorized",
			topics:             []string{"in"},
			deleteResults:      []cfk.TopicResult{},
			deleteError:        errors.New("Failed while waiting for controller: Local: Timed out"),
			expectedReturnCode: http.StatusInternalServerError,
			expectedError:      fmt.Errorf(`Unable to delete topic "in": ` + unauthorizedMessage),
		},
		{
			name:   "not-authorized",
			topics: []string{"in"},
			deleteResults: []cfk.TopicResult{
				cfk.TopicResult{Topic: "in"},
			},
			deleteError:        errors.New("REPLACE_ME"),
			expectedReturnCode: http.StatusUnauthorized,
			expectedError:      fmt.Errorf(`Unable to delete topic "in": ` + unauthorizedMessage),
		},
		{
			name:   "topic-not-found",
			topics: []string{"in"},
			deleteResults: []cfk.TopicResult{
				cfk.TopicResult{Topic: "in"},
			},
			deleteError:        errors.New("REPLACE_ME"),
			expectedReturnCode: http.StatusNotFound,
			expectedError:      fmt.Errorf(`Unable to delete topic "in": ` + topicNotFoundMessage),
		},
		{
			name:   "conn-error",
			topics: []string{"in"},
			deleteResults: []cfk.TopicResult{
				cfk.TopicResult{Topic: "in"},
			},
			deleteError:        errors.New("REPLACE_ME"),
			expectedReturnCode: http.StatusInternalServerError,
			expectedError:      fmt.Errorf(`Unable to delete topic "in": ` + kafkaConnectionMessage),
		},
		{
			name:   "multiple-delete-errors",
			topics: []string{"in", "out", "notification", "invalid"},
			deleteResults: []cfk.TopicResult{
				cfk.TopicResult{Topic: "in"},
			},
			deleteError: errors.New("REPLACE_ME"),
			/*deleteErrors: map[string]*error{
				"out":          &kafkaConnectionError,
				"notification": &topicNotFoundError,
				"invalid":      nil,
			},
			mockDeleteResponses: map[string]*http.Response{
				"out":          &StatusUnprocessableEntity,
				"notification": &StatusNotFound,
				"invalid":      &StatusUnauthorized,
			},*/
			expectedReturnCode: http.StatusInternalServerError,
			expectedError: fmt.Errorf(`Unable to delete topic "out": ` + kafkaConnectionMessage + "\n" +
				`Unable to delete topic "notification": missing header 'Authorization'` + "\n" +
				`Unable to delete topic "invalid": missing header 'Authorization`),
		},
	}

	for _, tc := range testCases {
		controller := gomock.NewController(t)
		defer controller.Finish()
		mockService := test.NewMockService(controller)

		mockService.
			EXPECT().
			DeleteTopics(context.Background(), tc.topics).
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
