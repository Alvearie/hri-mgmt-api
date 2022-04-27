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
	"github.com/Alvearie/hri-mgmt-api/common/kafka"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	cfk "github.com/confluentinc/confluent-kafka-go/kafka"
	"net/http"
	"time"
)

const deleteErrMessageTemplate = "Unable to delete topic \"%s\": %s"

func Delete(requestId string, topics []string, adminClient kafka.KafkaAdmin) (int, error) {
	prefix := "streams/Delete"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Streams Delete")

	ctx := context.Background()

	results, err := adminClient.DeleteTopics(ctx, topics, cfk.SetAdminRequestTimeout(time.Second*10))
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("unexpected error deleting Kafka topics: %w", err)
	}

	errs := make([]cfk.Error, 0, 4)
	for _, res := range results {
		if res.Error.Code() != cfk.ErrNoError {
			logger.Errorf(res.Error.Error())
			errs = append(errs, res.Error)
		}
	}

	if len(errs) > 0 {
		code, msg := getDeleteResponseError(errs[0])
		return code, errors.New(msg)
	}

	return http.StatusOK, nil
}

func getDeleteResponseError(err cfk.Error) (int, string) {
	code := err.Code()
	// In confluent-kafka-go library, Auth errors seem to get logged
	// as auth errors, but returned with very generic error messages so
	// we can't distinguish them
	if code == cfk.ErrUnknownTopic || code == cfk.ErrUnknownTopicOrPart {
		return http.StatusNotFound, err.Error()
	}
	return http.StatusInternalServerError, err.Error()
}
