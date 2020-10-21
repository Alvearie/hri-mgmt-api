/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package kafka

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

const FakeKafkaParams string = `{
		    "__bx_creds": {
		        "messagehub": {
					   "api_key": "FAKE_Api_Key",
					   "apikey": "FAKE_Api_Key",
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

const FakeParamsNoBrokers string = `{
		    "__bx_creds": {
		        "messagehub": {
					   "api_key": "FAKE_Api_Key",
					   "apikey": "FAKE_Api_Key",
					   "credentials": "dev-test",
					   "iam_apikey_name": "dev-test",
					   "kafka_http_url": "https://porcypine.kafka.eventstreams.cloud.ibm.com",
					   "password": "FAKE_Api_Key",
					   "user": "token"
		        }
		    }
	}`

const FakeParamsNoCreds string = `{
		    "__bx_creds": {
		        "messagehub": {
					   "api_key": "FAKE_Api_Key",
					   "apikey": "FAKE_Api_Key",
					   "credentials": "dev-test",
					   "iam_apikey_name": "dev-test",
					   "kafka_brokers_sasl": [
                       		"broker-5-porcypine.kafka.eventstreams.monkey.ibm.com:9093",
	                   		"broker-4-porcypine.kafka.eventstreams.monkey.ibm.com:9093",
				       		"broker-1-porcypine.kafka.eventstreams.monkey.ibm.com:9093",
	  			       		"broker-0-porcypine.kafka.eventstreams.monkey.ibm.com:9093",
				       		"broker-3-porcypine.kafka.eventstreams.monkey.ibm.com:9093",
				       		"broker-2-porcypine.kafka.eventstreams.monkey.ibm.com:9093"
					   ],
					   "kafka_http_url": "https://porcypine.kafka.eventstreams.cloud.ibm.com"
		        }
		    }
	}`

const (
	batchId   string = "batch987"
	topicBase string = "batchTopic"
)

var inputTopic = topicBase + ".in"
var batchMetadata = map[string]interface{}{"operation": "update"}

func TestNewWriterFromParam(t *testing.T) {

	var params map[string]interface{}
	if err := json.Unmarshal([]byte(FakeKafkaParams), &params); err != nil {
		t.Fatal(err)
	}

	_, err := NewWriterFromParams(params)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewWriterExtractBrokerFailure(t *testing.T) {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(FakeParamsNoBrokers), &params); err != nil {
		t.Fatal(err)
	}

	kafkaWriter, err := NewWriterFromParams(params)
	assert.Nil(t, kafkaWriter)
	assert.NotNil(t, err)
	assert.Equal(t, "error extracting kafka_brokers_sasl from Kafka credentials", err.Error())
}

func TestNewWriterCreateDialerError(t *testing.T) {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(FakeParamsNoCreds), &params); err != nil {
		t.Fatal(err)
	}

	kafkaWriter, err := NewWriterFromParams(params)
	assert.Nil(t, kafkaWriter)
	assert.NotNil(t, err)
	assert.Equal(t, "error extracting user from Kafka credentials", err.Error())
}

func TestWriterJsonErrors(t *testing.T) {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(FakeKafkaParams), &params); err != nil {
		t.Fatal(err)
	}

	kafkaWriter, err := NewWriterFromParams(params)
	if err != nil {
		t.Fatal(err)
	}

	notificationTopic := topicBase + ".notification"
	value := make(chan int)
	invalidBatchJson := map[string]interface{}{"some str": value}
	err = kafkaWriter.Write(notificationTopic, batchId, invalidBatchJson)
	assert.NotNil(t, err)
	assert.Equal(t, "json: unsupported type: chan int", err.Error())

	badValue := math.Inf(1)
	invalidValueJson := map[string]interface{}{"some str": badValue}
	err2 := kafkaWriter.Write(notificationTopic, batchId, invalidValueJson)
	assert.NotNil(t, err2)
	assert.Equal(t, "json: unsupported value: +Inf", err2.Error())
}
