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
	"net/http"
	"strings"

	"github.com/Alvearie/hri-mgmt-api/common/kafka"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	cfk "github.com/confluentinc/confluent-kafka-go/kafka"
)

const deleteErrMessageTemplate = "Unable to delete topic \"%s\": %s"

func DeleteStream(requestId string, topics []string, adminClient kafka.KafkaAdmin) (int, error) {
	prefix := "streams/Delete"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Streams Delete")

	ctx := context.Background()

	results, err := adminClient.DeleteTopics(ctx, topics, cfk.SetAdminRequestTimeout(kafka.AdminTimeout))
	if err != nil {
		var kafkaErr = &cfk.Error{}
		if errors.As(err, kafkaErr) {
			return getDeleteResponseErrorCode(*kafkaErr), fmt.Errorf("unexpected error deleting Kafka topics: %w", *kafkaErr)
		}
		// Should never be reached, err should be of type cfk.Error.
		return http.StatusInternalServerError, fmt.Errorf("unexpected error deleting Kafka topics: %w", err)
	}
	returnCode := http.StatusOK
	var errorMessageBuilder strings.Builder
	for _, res := range results {
		if res.Error.Code() != cfk.ErrNoError {
			logger.Errorln(res.Error.Error())
			if returnCode == http.StatusOK {
				returnCode = getDeleteResponseErrorCode(res.Error)
				fmt.Fprintf(&errorMessageBuilder, deleteErrMessageTemplate, res.Topic, res.Error.Error())
			} else {
				fmt.Fprintf(&errorMessageBuilder, "\n"+deleteErrMessageTemplate, res.Topic, res.Error.Error())
			}
		}
	}

	if returnCode != http.StatusOK {
		var err = fmt.Errorf(errorMessageBuilder.String())
		return returnCode, err
	}

	return http.StatusOK, nil
}

func getDeleteResponseErrorCode(err cfk.Error) int {
	code := err.Code()
	if code == cfk.ErrUnknownTopic || code == cfk.ErrUnknownTopicOrPart {
		return http.StatusNotFound
	} else if code == cfk.ErrTopicAuthorizationFailed || code == cfk.ErrGroupAuthorizationFailed || code == cfk.ErrClusterAuthorizationFailed {
		return http.StatusUnauthorized
	}
	return http.StatusInternalServerError
}
