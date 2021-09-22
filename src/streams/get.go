
package streams

import (
	"context"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/eventstreams"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	es "github.com/IBM/event-streams-go-sdk-generator/build/generated"
	"net/http"
	"strings"
)

const msgStreamsNotFound = "Unable to get stream names for tenant [%s]. %s"

func Get(
	requestId string, tenantId string,
	service eventstreams.Service) (int, interface{}) {
	prefix := "streams/Get"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("List streams for: " + tenantId)

	// get all topics for the kafka connection, then take only the streams for the given tenantId
	topicDetails, resp, err := service.ListTopics(context.Background(), &es.ListTopicsOpts{})
	if err != nil {
		msg := fmt.Sprintf(msgStreamsNotFound, tenantId, err.Error())
		logger.Errorln(msg)
		return getResponseError(requestId, resp, err)
	}

	streamNames := GetStreamNames(topicDetails, tenantId)
	return http.StatusOK, map[string]interface{}{"results": streamNames}
}

func getResponseError(requestId string, resp *http.Response, err error) (int, *response.ErrorDetail) {
	var returnCode = http.StatusInternalServerError
	var returnError = response.NewErrorDetail(requestId, err.Error())

	//EventStreams Admin API gives us status 403 when provided bearer token is unauthorized
	//and status 401 when Authorization isn't provided or is nil
	if resp.StatusCode == http.StatusForbidden {
		returnCode = http.StatusUnauthorized
		returnError = response.NewErrorDetail(requestId, eventstreams.UnauthorizedMsg)
	} else if resp.StatusCode == http.StatusUnauthorized {
		returnCode = http.StatusUnauthorized
		returnError = response.NewErrorDetail(requestId, eventstreams.MissingHeaderMsg)
	}
	return returnCode, returnError
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
