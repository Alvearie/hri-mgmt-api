/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package param

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

const FakeParams string = `{
		    "__bx_creds": {
		        "messagehub": {
					   "api_key": "FAKE_Api_Key",
					   "apiskeys": {
							"apikey":"FAKE_Api_Key"
						},
					  "authentication": {
						"method": "direct",
						"password": "elasticPassword",
						"username": "elasticUser"
					  },
					   "credentials": "dev-test",
					   "iam_apikey_description": "Auto-generated for key FAKE_00-12345",
					   "iam_apikey_name": "dev-test",
					   "iam_role_crn": "crn:v1:bluemix:public:iam::::serviceRole:Manager",
					   "iam_serviceid_crn": "crn:v1:bluemix:public:iam-identity::a/49f48a067ac4433a911740653049e83d::serviceid:ServiceId-blah-820a-blah-b372-291e60b71b67",
					   "instance": "HRI-Event Streams",
					   "instance_id": "my-unique-kafka-instance",
					   "kafka_admin_url": "https://porcypine.kafka.eventstreams.cloud.ibm.com",
					   "kafka_brokers_sasl": [
                       "broker-5-porcypine.kafka.eventstreams.monkey.ibm.com:9093",
	                   "broker-4-porcypine.kafka.eventstreams.monkey.ibm.com:9093",
				       "broker-1-porcypine.kafka.eventstreams.monkey.ibm.com:9093",
	  			       "broker-0-porcypine.kafka.eventstreams.monkey.ibm.com:9093",
				       "broker-3-porcypine.kafka.eventstreams.monkey.ibm.com:9093",
				       "broker-2-porcypine.kafka.eventstreams.monkey.ibm.com:9093"
					   ],
					   "kafka_http_url": "https://porcypine.kafka.eventstreams.cloud.ibm.com",
					   "password": "FAKE_Api_Key",
					   "user": "token"
		        }
		    }
		}`

const FakeMissingCredsParams string = `{
		    "__bx_creds": {
		    }
		}`

const FakeMissingParams string = `{
		    "__bx_creds": {
		        "messagehub": {
		        }
		    }
		}`

func TestExtractValuesSuccess(t *testing.T) {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(FakeParams), &params); err != nil {
		t.Fatal(err)
	}

	apisKeys, err := ExtractValues(params, BoundCreds, "messagehub", "apiskeys")
	assert.NotNil(t, apisKeys)
	assert.Nil(t, err)

	assert.Equal(t, apisKeys["apikey"], "FAKE_Api_Key")
}

func TestExtractValuesMissingUser(t *testing.T) {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(FakeMissingParams), &params); err != nil {
		t.Fatal(err)
	}

	user, err := ExtractValues(params, BoundCreds, "messagehub", "user")
	assert.Nil(t, user)
	assert.NotNil(t, err)
	assert.Equal(t, "error extracting the user section of the JSON", err.Error())
}

func TestExtractConnectorConfig(t *testing.T) {
	testCases := []struct {
		name            string
		creds           string
		expectedPass    string
		expectedPassErr string
	}{
		{
			name:            "missing-creds",
			creds:           FakeMissingCredsParams,
			expectedPassErr: fmt.Sprintf(MissingSectionMsg, KafkaResourceId),
		},
		{
			name:            "empty-creds",
			creds:           FakeMissingParams,
			expectedPassErr: fmt.Sprintf(MissingKafkaFieldMsg, "password"),
		},
		{
			name:         "valid-creds",
			creds:        FakeParams,
			expectedPass: "FAKE_Api_Key",
		},
	}

	for _, tc := range testCases {
		var params map[string]interface{}
		if err := json.Unmarshal([]byte(tc.creds), &params); err != nil {
			t.Fatal(err)
		}

		t.Run(tc.name, func(t *testing.T) {
			// extract password
			pass, err := ExtractString(params, "password")
			if pass != tc.expectedPass {
				t.Error(fmt.Sprintf("Expected: %v, Actual: %v", tc.expectedPass, pass))
			}
			if err != nil && err.Error() != tc.expectedPassErr {
				t.Error(fmt.Sprintf("Expected: %v, Actual: %v", tc.expectedPassErr, err.Error()))
			}
		})
	}
}
