/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package eventstreams

import (
	"context"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	es "github.com/IBM/event-streams-go-sdk-generator/build/generated"
	"net/http"
)

const (
	kafkaAdminUrl     string = "kafka_admin_url"
	bearerTokenHeader string = "authorization"
	MissingHeaderMsg  string = "Missing header 'Authorization'"
	UnauthorizedMsg   string = "Unauthorized to manage resource"
)

type Service interface {
	CreateTopic(ctx context.Context, topicCreate es.TopicCreateRequest) (map[string]interface{}, *http.Response, error)
	DeleteTopic(ctx context.Context, topicName string) (map[string]interface{}, *http.Response, error)
	ListTopics(ctx context.Context, localVarOptionals *es.ListTopicsOpts) ([]es.TopicDetail, *http.Response, error)
	HandleModelError(err error) *es.ModelError
}

type EventStreamsConnect struct {
	*es.DefaultApiService
}

// This method is a part of the Service interface solely for the purpose of mocking it in unit tests to cover all error cases
func (connect EventStreamsConnect) HandleModelError(err error) *es.ModelError {
	if err != nil {
		genericError, ok := err.(es.GenericOpenAPIError)
		if !ok {
			return &es.ModelError{Message: err.Error()}
		}
		modelError := genericError.Model().(es.ModelError)
		return &modelError
	}
	return nil
}

func CreateService(params map[string]interface{}) (Service, map[string]interface{}) {
	conf := es.NewConfiguration()
	adminUrl, err := param.ExtractString(params, kafkaAdminUrl)
	if err != nil {
		return nil, response.Error(http.StatusInternalServerError, err.Error())
	}
	conf.BasePath = adminUrl

	headerMap, err := param.ExtractValues(params, param.OpenWhiskHeaders)
	if err != nil {
		return nil, response.Error(http.StatusInternalServerError, err.Error())
	}

	bearerToken := headerMap[bearerTokenHeader]
	if bearerToken == nil {
		return nil, response.Error(http.StatusUnauthorized, MissingHeaderMsg)
	}
	conf.AddDefaultHeader(bearerTokenHeader, bearerToken.(string))

	client := es.NewAPIClient(conf)
	return EventStreamsConnect{client.DefaultApi}, nil
}
