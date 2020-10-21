/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package eventstreams

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	es "github.com/IBM/event-streams-go-sdk-generator/build/generated"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

const (
	missingFieldMsg string = "error extracting %s from Kafka credentials"
)

var statusOkResponse = http.Response{StatusCode: http.StatusOK}

func validCreds() string {
	return fmt.Sprintf(`{
        "__bx_creds": {
            "messagehub": {
                "kafka_admin_url": "valid url"
            }
        },
		"__ow_headers": {
			"authorization": "valid bearer token"
		}
    }`)
}

func missingKafkaUrlCreds() string {
	return fmt.Sprintf(`{
        "__bx_creds": {
            "messagehub": {
            }
        },
		"__ow_headers": {
			"authorization": "valid bearer token"
		}
    }`)
}

func missingHeadersCreds() string {
	return fmt.Sprintf(`{
        "__bx_creds": {
            "messagehub": {
                "kafka_admin_url": "valid url"
            }
        }
    }`)
}

func missingBearerCreds() string {
	return fmt.Sprintf(`{
        "__bx_creds": {
            "messagehub": {
                "kafka_admin_url": "valid url"
            }
        },
		"__ow_headers": {
		}
    }`)
}

func TestCreateEventStreamsServiceSuccess(t *testing.T) {

	var params map[string]interface{}
	if err := json.Unmarshal([]byte(validCreds()), &params); err != nil {
		t.Fatal(err)
	}

	service, err := CreateService(params)
	assert.NotNil(t, service)
	assert.Nil(t, err)
}

func TestCreateEventStreamsServiceMissingKafkaUrl(t *testing.T) {

	var params map[string]interface{}
	if err := json.Unmarshal([]byte(missingKafkaUrlCreds()), &params); err != nil {
		t.Fatal(err)
	}

	service, err := CreateService(params)
	assert.Nil(t, service)
	assert.Equal(t, response.Error(http.StatusInternalServerError, fmt.Sprintf(missingFieldMsg, kafkaAdminUrl)), err)
}

func TestCreateEventStreamsServiceMissingHeaders(t *testing.T) {

	var params map[string]interface{}
	if err := json.Unmarshal([]byte(missingHeadersCreds()), &params); err != nil {
		t.Fatal(err)
	}

	service, err := CreateService(params)
	assert.Nil(t, service)
	assert.Equal(t, response.Error(http.StatusInternalServerError, fmt.Sprintf(param.MissingSectionMsg, "__ow_headers")), err)
}

func TestCreateEventStreamsServiceMissingBearer(t *testing.T) {

	var params map[string]interface{}
	if err := json.Unmarshal([]byte(missingBearerCreds()), &params); err != nil {
		t.Fatal(err)
	}

	service, err := CreateService(params)
	assert.Nil(t, service)
	assert.Equal(t, response.Error(http.StatusUnauthorized, fmt.Sprintf(MissingHeaderMsg)), err)
}

func TestHandleModelErrorNil(t *testing.T) {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(validCreds()), &params); err != nil {
		t.Fatal(err)
	}
	service, _ := CreateService(params)
	modelErr := service.HandleModelError(nil)
	assert.Nil(t, modelErr)
}

func TestHandleModelErrorNonGeneric(t *testing.T) {

	testErr := errors.New("test error")
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(validCreds()), &params); err != nil {
		t.Fatal(err)
	}

	service, _ := CreateService(params)
	modelErr := service.HandleModelError(testErr)
	assert.Equal(t, &es.ModelError{Message: testErr.Error()}, modelErr)
}
