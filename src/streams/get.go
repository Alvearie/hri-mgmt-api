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
	cfk "github.com/confluentinc/confluent-kafka-go/kafka"
	"net/http"
	"strings"
)

const msgStreamsNotFound = "Unable to get stream names for tenant [%s]. %s"

func Get(
	requestId string, tenantId string,
	adminClient kafka.KafkaAdmin) (int, interface{}) {
	prefix := "streams/Get"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("List streams for: " + tenantId)

	// get all topics for the kafka connection, then take only the streams for the given tenantId
	topics, err := listTopics(adminClient)
	if err != nil {
		msg := fmt.Sprintf(msgStreamsNotFound, tenantId, err.Error())
		logger.Errorln(msg)

		returnCode := http.StatusInternalServerError
		var kafkaErr = &cfk.Error{}
		if errors.As(err, kafkaErr) {
			code := kafkaErr.Code()
			if code == cfk.ErrTopicAuthorizationFailed || code == cfk.ErrGroupAuthorizationFailed || code == cfk.ErrClusterAuthorizationFailed {
				returnCode = http.StatusUnauthorized
			}
		}
		return returnCode, err
	}

	streamNames := GetStreamNames(topics, tenantId)
	return http.StatusOK, map[string]interface{}{"results": streamNames}
}

func GetStreamNames(topics []string, tenantId string) []map[string]interface{} {
	streamNames := []map[string]interface{}{}
	seenStreamIds := make(map[string]bool)
	for _, topicName := range topics {
		splits := strings.Split(topicName, ".")
		if len(splits) >= 4 && validTopicName(topicName) && splits[1] == tenantId {
			//streamId is between tenantId and suffix, and it includes the dataIntegratorId and optional qualifier (delimited by '.')
			streamId := strings.TrimPrefix(topicName, kafka.TopicPrefix+tenantId+".")
			streamId = strings.TrimSuffix(streamId, kafka.InSuffix)
			streamId = strings.TrimSuffix(streamId, kafka.NotificationSuffix)
			streamId = strings.TrimSuffix(streamId, kafka.OutSuffix)
			streamId = strings.TrimSuffix(streamId, kafka.InvalidSuffix)

			//take unique stream names, we don't want duplicates due to a stream's multiple topics (in/notification)
			if _, seen := seenStreamIds[streamId]; !seen {
				streamNames = append(streamNames, map[string]interface{}{param.StreamId: streamId})
				seenStreamIds[streamId] = true
			}
		}
	}
	return streamNames
}

func listTopics(adminClient kafka.KafkaAdmin) ([]string, error) {
	// Passing nil and true as first two params returns info on all topics.
	metadata, err := adminClient.GetMetadata(nil, true, 10000)
	if err != nil {
		return nil, fmt.Errorf("error listing Kafka topics: %w", err)
	}

	// metadata.Topics is a map from topic names to structs with topic info
	i := 0
	topics := make([]string, len(metadata.Topics))
	for key, _ := range metadata.Topics {
		topics[i] = key
		i++
	}

	return topics, nil
}

func validTopicName(topicName string) bool {
	return strings.HasPrefix(topicName, kafka.TopicPrefix) &&
		(strings.HasSuffix(topicName, kafka.InSuffix) || strings.HasSuffix(topicName, kafka.NotificationSuffix) ||
			strings.HasSuffix(topicName, kafka.OutSuffix) || strings.HasSuffix(topicName, kafka.InvalidSuffix))
}
