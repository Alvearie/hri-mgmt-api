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
	es "github.com/IBM/event-streams-go-sdk-generator/build/generated"
	"github.com/golang/mock/gomock"
	"ibm.com/watson/health/foundation/hri/common/eventstreams"
	"ibm.com/watson/health/foundation/hri/common/logwrapper"
	"ibm.com/watson/health/foundation/hri/common/test"
	"net/http"
	"os"
	"reflect"
	"testing"
)

const topicNotFoundMessage string = "Topic not found"

func TestDelete(t *testing.T) {
	logwrapper.Initialize("error", os.Stdout)
	var requestId = "reqIdZz05"
	StatusNotFound := http.Response{StatusCode: 404}
	var NotFoundError = es.ModelError{
		Message: topicNotFoundMessage,
	}

	testCases := []struct {
		name                string
		topics              []string
		mockDeleteResponses map[string]*http.Response
		deleteErrors        map[string]*es.ModelError
		expectedReturnCode  int
		expectedError       error
	}{
		{
			name:               "happy path",
			topics:             []string{"in", "out", "notification", "invalid"},
			expectedReturnCode: http.StatusOK,
		},
		{
			name:                "not-authorized",
			topics:              []string{"in"},
			deleteErrors:        map[string]*es.ModelError{"in": {}},
			mockDeleteResponses: map[string]*http.Response{"in": &StatusForbidden},
			expectedReturnCode:  http.StatusUnauthorized,
			expectedError:       fmt.Errorf(`Unable to delete topic "in": ` + eventstreams.UnauthorizedMsg),
		},
		{
			name:                "not-authorized",
			topics:              []string{"in"},
			deleteErrors:        map[string]*es.ModelError{"in": {}},
			mockDeleteResponses: map[string]*http.Response{"in": &StatusUnauthorized},
			expectedReturnCode:  http.StatusUnauthorized,
			expectedError:       fmt.Errorf(`Unable to delete topic "in": ` + eventstreams.MissingHeaderMsg),
		},
		{
			name:                "topic-not-found",
			topics:              []string{"in"},
			deleteErrors:        map[string]*es.ModelError{"in": &NotFoundError},
			mockDeleteResponses: map[string]*http.Response{"in": &StatusNotFound},
			expectedReturnCode:  http.StatusNotFound,
			expectedError:       fmt.Errorf(`Unable to delete topic "in": ` + topicNotFoundMessage),
		},
		{
			name:                "conn-error",
			topics:              []string{"in"},
			deleteErrors:        map[string]*es.ModelError{"in": &OtherError},
			mockDeleteResponses: map[string]*http.Response{"in": &StatusUnprocessableEntity},
			expectedReturnCode:  http.StatusInternalServerError,
			expectedError:       fmt.Errorf(`Unable to delete topic "in": ` + kafkaConnectionMessage),
		},
		{
			name:   "multiple-delete-errors",
			topics: []string{"in", "out", "notification", "invalid"},
			deleteErrors: map[string]*es.ModelError{
				"out":          &OtherError,
				"notification": &NotFoundError,
				"invalid":      {},
			},
			mockDeleteResponses: map[string]*http.Response{
				"out":          &StatusUnprocessableEntity,
				"notification": &StatusNotFound,
				"invalid":      &StatusUnauthorized,
			},
			expectedReturnCode: http.StatusInternalServerError,
			expectedError: fmt.Errorf(`Unable to delete topic "out": ` + kafkaConnectionMessage + "\n" +
				`Unable to delete topic "notification": ` + topicNotFoundMessage + "\n" +
				`Unable to delete topic "invalid": ` + eventstreams.MissingHeaderMsg),
		},
	}

	for _, tc := range testCases {
		controller := gomock.NewController(t)
		defer controller.Finish()
		mockService := test.NewMockService(controller)

		for _, topic := range tc.topics {
			if tc.deleteErrors[topic] == nil {
				mockService.
					EXPECT().
					DeleteTopic(context.Background(), topic).
					Return(nil, nil, nil).
					MaxTimes(1)
			} else {
				mockErr := errors.New(tc.deleteErrors[topic].Message)

				mockService.
					EXPECT().
					DeleteTopic(context.Background(), topic).
					Return(nil, tc.mockDeleteResponses[topic], mockErr).
					MaxTimes(1)

				mockService.
					EXPECT().
					HandleModelError(mockErr).
					Return(tc.deleteErrors[topic]).
					MaxTimes(1)
			}
		}

		actualCode, actualErrMsg := Delete(requestId, tc.topics, mockService)

		t.Run(tc.name, func(t *testing.T) {
			if actualCode != tc.expectedReturnCode || !reflect.DeepEqual(tc.expectedError, actualErrMsg) {
				t.Errorf("Stream-Delete()\n   actual: %v,%v\n expected: %v,%v",
					actualCode, actualErrMsg, tc.expectedReturnCode, tc.expectedError)
			}
		})
	}
}
