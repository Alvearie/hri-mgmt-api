/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package streams

import (
	"context"
	"github.com/Alvearie/hri-mgmt-api/common/eventstreams"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/path"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	es "github.com/IBM/event-streams-go-sdk-generator/build/generated"
	"log"
	"net/http"
	"os"
	"strings"
)

func Get(
	args map[string]interface{},
	service eventstreams.Service) map[string]interface{} {

	logger := log.New(os.Stdout, "streams/get: ", log.Llongfile)

	// extract tenantId path params from URL
	tenantId, err := path.ExtractParam(args, param.TenantIndex)
	if err != nil {
		logger.Println(err.Error())
		return response.Error(http.StatusBadRequest, err.Error())
	}

	// get all topics for the kafka connection, then take only the streams for the given tenantId
	topicDetails, resp, err := service.ListTopics(context.Background(), &es.ListTopicsOpts{})
	if err != nil {
		logger.Printf("Unable to get stream names for tenant [%s]. %s", tenantId, err.Error())
		return getResponseError(resp, err)
	}

	streamNames := GetStreamNames(topicDetails, tenantId)
	return response.Success(http.StatusOK, map[string]interface{}{"results": streamNames})
}

func getResponseError(resp *http.Response, err error) map[string]interface{} {

	//EventStreams Admin API gives us status 403 when provided bearer token is unauthorized
	//and status 401 when Authorization isn't provided or is nil
	if resp.StatusCode == http.StatusForbidden {
		return response.Error(http.StatusUnauthorized, eventstreams.UnauthorizedMsg)
	} else if resp.StatusCode == http.StatusUnauthorized {
		return response.Error(http.StatusUnauthorized, eventstreams.MissingHeaderMsg)
	}
	return response.Error(http.StatusInternalServerError, err.Error())
}

func GetStreamNames(topics []es.TopicDetail, tenantId string) []map[string]interface{} {
	streamNames := []map[string]interface{}{}
	seenStreamIds := make(map[string]bool)
	for _, topic := range topics {
		topicName := topic.Name
		splits := strings.Split(topicName, ".")
		if len(splits) >= 4 && validTopicName(topicName) && splits[1] == tenantId {

			//streamId is between tenantId and suffix, and it includes the dataIntegratorId and optional qualifier (delimited by '.')
			streamId := strings.TrimPrefix(topicName, eventstreams.TopicPrefix+tenantId+".")
			streamId = strings.TrimSuffix(streamId, eventstreams.InSuffix)
			streamId = strings.TrimSuffix(streamId, eventstreams.NotificationSuffix)
			streamId = strings.TrimSuffix(streamId, eventstreams.OutSuffix)
			streamId = strings.TrimSuffix(streamId, eventstreams.InvalidSuffix)

			//take unique stream names, we don't want duplicates due to a stream's multiple topics (in/notification)
			if _, seen := seenStreamIds[streamId]; !seen {
				streamNames = append(streamNames, map[string]interface{}{param.StreamId: streamId})
				seenStreamIds[streamId] = true
			}
		}
	}
	return streamNames
}

func validTopicName(topicName string) bool {
	return strings.HasPrefix(topicName, eventstreams.TopicPrefix) &&
		(strings.HasSuffix(topicName, eventstreams.InSuffix) || strings.HasSuffix(topicName, eventstreams.NotificationSuffix) ||
			strings.HasSuffix(topicName, eventstreams.OutSuffix) || strings.HasSuffix(topicName, eventstreams.InvalidSuffix))
}
