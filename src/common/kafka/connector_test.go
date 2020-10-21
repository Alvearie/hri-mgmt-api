/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package kafka

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	"github.com/golang/mock/gomock"
	kg "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

const (
	testBroker1 string = "broker-0-porcypine.kafka.eventstreams.monkey.ibm.com:9093"
	testBroker2 string = "broker-1-porcypine.kafka.eventstreams.monkey.ibm.com:9093"
	testBroker3 string = "broker-2-porcypine.kafka.eventstreams.monkey.ibm.com:9093"
	testBroker4 string = "broker-3-porcypine.kafka.eventstreams.monkey.ibm.com:9093"
	testBroker5 string = "broker-4-porcypine.kafka.eventstreams.monkey.ibm.com:9093"
	badBroker   string = "bad-broker-address.monkey.ibm.com:9093"
	hostPrefix  string = "broker-"
	hostRoot    string = "-porcypine.kafka.eventstreams.monkey.ibm.com"
	host1       string = hostPrefix + "0" + hostRoot
	host2       string = hostPrefix + "1" + hostRoot
	host3       string = hostPrefix + "2" + hostRoot
	host4       string = hostPrefix + "3" + hostRoot
	host5       string = hostPrefix + "4" + hostRoot
	host6       string = hostPrefix + "5" + hostRoot
	password    string = "monkeePass"
)

func missingCreds() string {
	return `{
        "__bx_creds": {
        }
    }`
}

func emptyCreds() string {
	return `{
        "__bx_creds": {
             "messagehub": {}
        }
    }`
}

func validCreds() string {
	return fmt.Sprintf(`{
        "__bx_creds": {
            "messagehub": {
                "kafka_brokers_sasl": ["%s", "%s"],
                "password": "%s",
                "user": "token"
            }
        }
    }`, testBroker1, testBroker2, password)
}

func validFiveBrokers() string {
	return fmt.Sprintf(`{
        "__bx_creds": {
            "messagehub": {
			   "apikey": "FAKE_Api_Key",
			   "credentials": "dev-test",
				"instance": "HRI-Event Streams",
       			"instance_id": "12345-1310-4d53-9f5e-pork",
       			"kafka_admin_url": "https://pig.svc01.us-south.eventstreams.cloud.ibm.com",
                "kafka_brokers_sasl": ["%s", "%s", "%s", "%s", "%s"],
                "password": "%s",
                "user": "token"
            }
        }
    }`, testBroker1, testBroker2, testBroker3, testBroker4, testBroker5, password)
}

func invalidBroker() string {
	return fmt.Sprintf(`{
        "__bx_creds": {
            "messagehub": {
			   "apikey": "FAKE_Api_Key",
			   "credentials": "dev-test",
				"instance": "HRI-Event Streams",
       			"instance_id": "12345-1310-4d53-9f5e-pork",
       			"kafka_admin_url": "https://pig.svc01.us-south.eventstreams.cloud.ibm.com",
                "kafka_brokers_sasl": ["%s"],
                "password": "%s",
                "user": "token"
            }
        }
    }`, badBroker, password)
}

func expectedHosts() []string {
	return []string{host1, host2, host3, host4, host5, host6}
}

func TestExtractConnectorConfig(t *testing.T) {
	testCases := []struct {
		name               string
		creds              string
		expectedBrokers    []string
		expectedBrokersErr string
		expectedPass       string
		expectedPassErr    string
	}{
		{
			name:               "missing-creds",
			creds:              missingCreds(),
			expectedBrokersErr: fmt.Sprintf(param.MissingSectionMsg, param.KafkaResourceId),
			expectedPassErr:    fmt.Sprintf(param.MissingSectionMsg, param.KafkaResourceId),
		},
		{
			name:               "empty-creds",
			creds:              emptyCreds(),
			expectedBrokersErr: fmt.Sprintf(param.MissingKafkaFieldMsg, fieldBrokers),
			expectedPassErr:    fmt.Sprintf(param.MissingKafkaFieldMsg, fieldPassword),
		},
		{
			name:            "valid-creds",
			creds:           validCreds(),
			expectedBrokers: []string{testBroker1, testBroker2},
			expectedPass:    password,
		},
	}

	for _, tc := range testCases {
		var params map[string]interface{}
		if err := json.Unmarshal([]byte(tc.creds), &params); err != nil {
			t.Fatal(err)
		}

		t.Run(tc.name, func(t *testing.T) {
			// extract brokers
			brokers, err := extractBrokers(params)
			if !reflect.DeepEqual(brokers, tc.expectedBrokers) {
				t.Error(fmt.Sprintf("Expected: %v, Actual: %v", tc.expectedBrokers, brokers))
			}
			if err != nil && err.Error() != tc.expectedBrokersErr {
				t.Error(fmt.Sprintf("Expected: %v, Actual: %v", tc.expectedBrokersErr, err.Error()))
			}

			// extract password
			pass, err := param.ExtractString(params, fieldPassword)
			if pass != tc.expectedPass {
				t.Error(fmt.Sprintf("Expected: %v, Actual: %v", tc.expectedPass, pass))
			}
			if err != nil && err.Error() != tc.expectedPassErr {
				t.Error(fmt.Sprintf("Expected: %v, Actual: %v", tc.expectedPassErr, err.Error()))
			}
		})
	}
}

func TestConnFromParamsFailures(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	tests := []struct {
		name           string
		jsonParams     string
		dialer         ContextDialer
		expectedErrMsg string
	}{
		{
			name: "missing-kafka_brokers_sasl-param",
			jsonParams: `{
			    "__bx_creds": {
			        "messagehub": {
						   "apikey": "FAKE_Api_Key",
						   "credentials": "dev-test",
						   "instance": "HRI-Event Streams",
						   "kafka_admin_url": "https://porcypine.kafka.eventstreams.cloud.ibm.com",
						   "kafka_http_url": "https://porcypine.kafka.eventstreams.cloud.ibm.com",
						   "password": "FAKE_Api_Key",
						   "user": "token"
			        }
			    }
			}`,
			dialer:         test.NewMockContextDialer(controller),
			expectedErrMsg: "error extracting kafka_brokers_sasl from Kafka credentials",
		}, {
			name: "empty-kafka_brokers_sasl-list-param",
			jsonParams: `{
			    "__bx_creds": {
			        "messagehub": {
						   "apikey": "FAKE_Api_Key",
						   "credentials": "dev-test",
						   "instance": "HRI-Event Streams",
						   "kafka_admin_url": "https://porcypine.kafka.eventstreams.cloud.ibm.com",
						   "kafka_http_url": "https://porcypine.kafka.eventstreams.cloud.ibm.com",
						   "kafka_brokers_sasl": [ ],						   
						   "password": "FAKE_Api_Key",
						   "user": "token"
			        }
			    }
			}`,
			dialer:         test.NewMockContextDialer(controller),
			expectedErrMsg: "No Kafka network address provided in field kafka_brokers_sasl from Kafka credentials",
		}, {
			name: "kafka-brokers-param-unexpected-json-only-1-broker",
			jsonParams: `{
			    "__bx_creds": {
			        "messagehub": {
						   "apikey": "FAKE_Api_Key",
						   "credentials": "dev-test",
						   "instance": "HRI-Event Streams",
						   "kafka_admin_url": "https://porcypine.kafka.eventstreams.cloud.ibm.com",
						   "kafka_http_url": "https://porcypine.kafka.eventstreams.cloud.ibm.com",
						   "kafka_brokers_sasl": "broker-orcypine.kafka.eventstreams.monkey.ibm.com:9093",						   
						   "password": "FAKE_Api_Key",
						   "user": "token"
			        }
			    }
			}`,
			dialer:         test.NewMockContextDialer(controller),
			expectedErrMsg: "error extracting kafka_brokers_sasl from Kafka credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var params map[string]interface{}
			if err := json.Unmarshal([]byte(tt.jsonParams), &params); err != nil {
				t.Fatal(err)
			}

			_, err := ConnFromParams(params, tt.dialer)
			if err == nil || err.Error() != tt.expectedErrMsg {
				t.Errorf("Kafka ConnFromParams() = %v, expected: %s", err, tt.expectedErrMsg)
			}
		})
	}

}

func TestCreateDialerCredentialFailures(t *testing.T) {
	tests := []struct {
		name       string
		jsonParams string
		expected   string
	}{
		{name: "missing user credential param",
			jsonParams: `{
			    "__bx_creds": {
			        "messagehub": {
						   "apikey": "FAKE_Api_Key",
						   "instance": "HRI-Event Streams",
						   "kafka_admin_url": "https://porcypine.kafka.eventstreams.cloud.ibm.com",
						   "kafka_brokers_sasl": [
					       "broker-1-porcypine.kafka.eventstreams.monkey.ibm.com:9093",
		  			       "broker-0-porcypine.kafka.eventstreams.monkey.ibm.com:9093",
					       "broker-2-porcypine.kafka.eventstreams.monkey.ibm.com:9093"
						   ],
						   "kafka_http_url": "https://porcypine.kafka.eventstreams.cloud.ibm.com",
					   	   "password": "FAKE_Api_Key"
			        }
			    }
			}`,
			expected: "error extracting user from Kafka credentials",
		},
		{name: "missing password credential param",
			jsonParams: `{
			    "__bx_creds": {
			        "messagehub": {
						   "apikey": "FAKE_Api_Key",
						   "instance": "HRI-Event Streams",
						   "kafka_admin_url": "https://porcypine.kafka.eventstreams.cloud.ibm.com",
						   "kafka_brokers_sasl": [
					       		"broker-1-porcypine.kafka.eventstreams.monkey.ibm.com:9093",
		  			       		"broker-0-porcypine.kafka.eventstreams.monkey.ibm.com:9093"
						   ],
						   "kafka_http_url": "https://porcypine.kafka.eventstreams.cloud.ibm.com",
					   	   "user": "token"
			        }
			    }
			}`,
			expected: "error extracting password from Kafka credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var params map[string]interface{}
			if err := json.Unmarshal([]byte(tt.jsonParams), &params); err != nil {
				t.Fatal(err)
			}

			_, err := CreateDialer(params)
			if err == nil || err.Error() != tt.expected {
				t.Errorf("Kafka CreateDialer() = %v, expected: %s", err, tt.expected)
			}
		})
	}
}

func TestConnectorSuccess(t *testing.T) {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(validCreds()), &params); err != nil {
		t.Fatal(err)
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	emptyConn := &kg.Conn{} //Upon Connection Success - we return a valid Conn obj.fakeConn := &kg.Conn{}
	mockDialer := test.NewMockContextDialer(controller)
	mockDialer.
		EXPECT().
		DialContext(gomock.Any(), defaultNetwork, testBroker1).
		Return(emptyConn, nil)

	rtnConn, err := ConnFromParams(params, mockDialer)
	if assert.NoError(t, err) {
		assert.Equal(t, emptyConn, rtnConn)
	}
}

func TestConnector_ShouldIterateOverBrokerAddresses2Failures(t *testing.T) {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(validFiveBrokers()), &params); err != nil {
		t.Fatal(err)
	}

	controller := gomock.NewController(t)
	defer controller.Finish()
	connError := errors.New("dial tcp 123.45.321.34:9093: i/o timeout ")

	emptyConn := &kg.Conn{} //Upon Connection Success - we return a valid Conn obj.
	mockDialer := test.NewMockContextDialer(controller)

	firstCall := mockDialer.
		EXPECT().
		DialContext(gomock.Any(), defaultNetwork, testBroker1).
		Return(nil, connError)

	secondCall := mockDialer.
		EXPECT().
		DialContext(gomock.Any(), defaultNetwork, testBroker2).
		Return(nil, connError).
		After(firstCall)

	mockDialer.
		EXPECT().
		DialContext(gomock.Any(), defaultNetwork, testBroker3).
		Return(emptyConn, nil).
		After(secondCall)

	rtnConn, err := ConnFromParams(params, mockDialer)
	if assert.NoError(t, err) {
		assert.Equal(t, emptyConn, rtnConn)
	}
}

func TestConnectorConnectionFailure(t *testing.T) {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(invalidBroker()), &params); err != nil {
		t.Fatal(err)
	}

	controller := gomock.NewController(t)
	defer controller.Finish()
	connError := errors.New("dial tcp 123.45.321.34:9093: i/o timeout ")
	mockDialer := test.NewMockContextDialer(controller)

	mockDialer.
		EXPECT().
		DialContext(gomock.Any(), defaultNetwork, badBroker).
		Return(nil, connError)

	_, err := ConnFromParams(params, mockDialer)
	expectedErr := fmt.Sprintf(kafkaConnFailMsg, connError)
	assert.Equal(t, expectedErr, err.Error())
}
