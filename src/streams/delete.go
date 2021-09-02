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
	"reflect"
)

func Delete(
	args map[string]interface{},
	validator param.Validator,
	service eventstreams.Service) map[string]interface{} {

	logger := log.New(os.Stdout, "streams/delete: ", log.Llongfile)

	// validate that required input params are present
	errResp := validator.Validate(
		args,
		param.Info{param.Validation, reflect.Bool},
	)
	if errResp != nil {
		logger.Printf("Bad input params: %s", errResp)
		return errResp
	}

	// extract tenantId and streamId path params from URL
	tenantId, err := path.ExtractParam(args, param.TenantIndex)
	if err != nil {
		logger.Println(err.Error())
		return response.Error(http.StatusBadRequest, err.Error())
	}
	streamId, err := path.ExtractParam(args, param.StreamIndex)
	if err != nil {
		logger.Println(err.Error())
		return response.Error(http.StatusBadRequest, err.Error())
	}

	inTopicName, notificationTopicName, outTopicName, invalidTopicName := eventstreams.CreateTopicNames(tenantId, streamId)

	// delete the in and notification topics for the given tenant and data integrator pairing
	_, inResp, inErr := service.DeleteTopic(context.Background(), inTopicName)
	if inErr != nil {
		logger.Printf("Unable to delete topic [%s]. %s", inTopicName, inErr.Error())
		return getDeleteResponseError(inResp, service.HandleModelError(inErr))
	}

	_, notificationResp, notificationErr := service.DeleteTopic(context.Background(), notificationTopicName)
	if notificationErr != nil {
		logger.Printf("Unable to delete topic [%s]. %s", notificationTopicName, notificationErr.Error())
		return getDeleteResponseError(notificationResp, service.HandleModelError(notificationErr))
	}

	validation := args[param.Validation].(bool)

	//if validation is enabled, delete the out and invalid topics that were created
	if validation {
		_, outResp, outErr := service.DeleteTopic(context.Background(), outTopicName)
		if outErr != nil {
			logger.Printf("Unable to delete topic [%s]. %s", outTopicName, outErr.Error())
			return getDeleteResponseError(outResp, service.HandleModelError(outErr))
		}

		_, invalidResp, invalidErr := service.DeleteTopic(context.Background(), invalidTopicName)
		if invalidErr != nil {
			logger.Printf("Unable to delete topic [%s]. %s", invalidTopicName, invalidErr.Error())
			return getDeleteResponseError(invalidResp, service.HandleModelError(invalidErr))
		}
	}

	return response.Success(http.StatusOK, map[string]interface{}{})
}

func getDeleteResponseError(resp *http.Response, err *es.ModelError) map[string]interface{} {

	//EventStreams Admin API gives us status 403 when provided bearer token is unauthorized
	//and status 401 when Authorization isn't provided or is nil
	if resp.StatusCode == http.StatusForbidden {
		return response.Error(http.StatusUnauthorized, eventstreams.UnauthorizedMsg)
	} else if resp.StatusCode == http.StatusUnauthorized {
		return response.Error(http.StatusUnauthorized, eventstreams.MissingHeaderMsg)
	} else if resp.StatusCode == http.StatusNotFound {
		return response.Error(http.StatusNotFound, err.Message)
	}
	return response.Error(http.StatusInternalServerError, err.Message)
}
