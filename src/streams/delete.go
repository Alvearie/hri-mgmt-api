/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package streams

import (
	"context"
	"fmt"
	es "github.com/IBM/event-streams-go-sdk-generator/build/generated"
	"ibm.com/watson/health/foundation/hri/common/eventstreams"
	"ibm.com/watson/health/foundation/hri/common/logwrapper"
	"net/http"
	"strings"
)

const deleteErrMessageTemplate = "Unable to delete topic \"%s\": %s"

func Delete(requestId string, topics []string, service eventstreams.Service) (int, error) {
	prefix := "streams/Delete"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Streams Delete")

	returnCode := http.StatusOK
	var errorMessageBuilder strings.Builder

	for _, topic := range topics {
		logger.Debugln("Delete Stream: " + topic)
		_, deleteResp, err := service.DeleteTopic(context.Background(), topic)
		if err != nil {
			deleteReturnCode, deleteErrMessage := getDeleteResponseError(deleteResp, service.HandleModelError(err))
			if returnCode == http.StatusOK {
				// Save the first error's code to return later. Initialize the error message
				returnCode = deleteReturnCode
				fmt.Fprintf(&errorMessageBuilder, deleteErrMessageTemplate, topic, deleteErrMessage)
			} else {
				fmt.Fprintf(&errorMessageBuilder, "\n"+deleteErrMessageTemplate, topic, deleteErrMessage)
			}
		}
	}

	if returnCode != http.StatusOK {
		var err = fmt.Errorf(errorMessageBuilder.String())
		logger.Errorln(err.Error())
		return returnCode, err
	}

	return returnCode, nil
}

func getDeleteResponseError(resp *http.Response, err *es.ModelError) (int, string) {
	//EventStreams Admin API gives us status 403 when provided bearer token is unauthorized
	//and status 401 when Authorization isn't provided or is nil
	if resp.StatusCode == http.StatusForbidden {
		return http.StatusUnauthorized, eventstreams.UnauthorizedMsg
	} else if resp.StatusCode == http.StatusUnauthorized {
		return http.StatusUnauthorized, eventstreams.MissingHeaderMsg
	} else if resp.StatusCode == http.StatusNotFound {
		return http.StatusNotFound, err.Message
	}

	return http.StatusInternalServerError, err.Message
}
