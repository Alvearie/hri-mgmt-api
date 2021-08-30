/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package eventstreams

import (
	"context"
	configPkg "github.com/Alvearie/hri-mgmt-api/common/config"
	es "github.com/IBM/event-streams-go-sdk-generator/build/generated"
	"net/http"
)

const (
	bearerTokenHeader string = "authorization"
	MissingHeaderMsg  string = "missing header 'Authorization'"
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

// HandleModelError is a part of the Service interface solely for the purpose of mocking it in unit tests to cover all error cases
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

func CreateServiceFromConfig(config configPkg.Config, bearerToken string) Service {
	esConfig := es.NewConfiguration()
	esConfig.BasePath = config.KafkaAdminUrl
	esConfig.AddDefaultHeader(bearerTokenHeader, bearerToken)
	client := es.NewAPIClient(esConfig)
	return EventStreamsConnect{client.DefaultApi}
}
