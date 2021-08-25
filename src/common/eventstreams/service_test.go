/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package eventstreams

import (
	"errors"
	es "github.com/IBM/event-streams-go-sdk-generator/build/generated"
	"github.com/stretchr/testify/assert"
	"ibm.com/watson/health/foundation/hri/common/config"
	"testing"
)

func TestCreateEventstreamsService(t *testing.T) {
	config := config.Config{
		ConfigPath:      "",
		OidcIssuer:      "",
		JwtAudienceId:   "",
		Validation:      false,
		ElasticUrl:      "",
		ElasticUsername: "",
		ElasticPassword: "",
		ElasticCert:     "",
	}
	service := CreateServiceFromConfig(config, "token")
	assert.NotNil(t, service)
}

func TestHandleModelErrorNil(t *testing.T) {
	config := config.Config{
		ConfigPath:      "",
		OidcIssuer:      "",
		JwtAudienceId:   "",
		Validation:      false,
		ElasticUrl:      "",
		ElasticUsername: "",
		ElasticPassword: "",
		ElasticCert:     "",
	}
	service := CreateServiceFromConfig(config, "token")
	modelErr := service.HandleModelError(nil)
	assert.Nil(t, modelErr)
}

func TestHandleModelErrorNonGeneric(t *testing.T) {
	testErr := errors.New("test error")
	config := config.Config{
		ConfigPath:      "",
		OidcIssuer:      "",
		JwtAudienceId:   "",
		Validation:      false,
		ElasticUrl:      "",
		ElasticUsername: "",
		ElasticPassword: "",
		ElasticCert:     "",
	}
	service := CreateServiceFromConfig(config, "token")
	modelErr := service.HandleModelError(testErr)
	assert.Equal(t, &es.ModelError{Message: testErr.Error()}, modelErr)
}
