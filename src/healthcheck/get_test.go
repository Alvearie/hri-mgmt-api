/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package healthcheck

import (
	"bytes"
	"errors"
	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/Alvearie/hri-mgmt-api/common/kafka"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"testing"
)

const activationId string = "testActivationId"

func TestHealthcheck(t *testing.T) {
	_ = os.Setenv(response.EnvOwActivationId, activationId)

	//Success Case Kafka Partition Reader
	defaultKafkaReader := test.FakePartitionReader{
		T:          t,
		Partitions: test.GetFakeTwoPartitionSlice(),
		Err:        nil,
	}

	testCases := []struct {
		name        string
		args        map[string]interface{} //Will always be empty for these tests/healthcheck.get()
		transport   *test.FakeTransport
		kafkaReader kafka.PartitionReader
		expected    map[string]interface{}
	}{
		{
			name: "Success-case",
			transport: test.NewFakeTransport(t).AddCall(
				"/_cat/health",
				test.ElasticCall{
					ResponseBody: test.ReaderToString(ioutil.NopCloser(bytes.NewReader([]byte(`
					[{
						"epoch": "1578512886",
						"timestamp": "19:48:06",
						"cluster": "8165307e-6130-4581-942d-20fcfc4e795d",
						"status": "green",
						"node.total": "3",
						"node.data": "3",
						"shards": "19",
						"pri": "9",
						"relo": "0",
						"init": "0",
						"unassign": "0",
						"pending_tasks": "0",
						"max_task_wait_time": "-",
						"active_shards_percent": "100.0%"
					}]`)))),
				},
			),
			kafkaReader: defaultKafkaReader,
			expected:    response.Success(http.StatusOK, map[string]interface{}{}),
		},
		{
			name: "elastic-search-bad-status",
			transport: test.NewFakeTransport(t).AddCall(
				"/_cat/health",
				test.ElasticCall{
					ResponseBody: test.ReaderToString(ioutil.NopCloser(bytes.NewReader([]byte(`
					[{
						 "epoch": "1578512886",
						 "timestamp": "19:48:06",
						 "cluster": "8165307e-6130-4581-942d-20fcfc4e795d",
						 "status": "red",
						 "node.total": "3",
						 "node.data": "3",
						 "shards": "5",
						 "pri": "1",
						 "relo": "0",
						 "init": "0",
						 "unassign": "2",
						 "pending_tasks": "4",
						 "max_task_wait_time": "-",
						 "active_shards_percent": "50.0%"
					}]`)))),
				},
			),
			kafkaReader: defaultKafkaReader,
			expected:    response.Error(http.StatusServiceUnavailable, "HRI Service Temporarily Unavailable | error Detail: ElasticSearch status: red, clusterId: 8165307e-6130-4581-942d-20fcfc4e795d, unixTimestamp: 1578512886"),
		},
		{
			name: "invalid-ES-response-missing-status-field",
			transport: test.NewFakeTransport(t).AddCall(
				"/_cat/health",
				test.ElasticCall{
					ResponseBody: test.ReaderToString(ioutil.NopCloser(bytes.NewReader([]byte(`
					[{
						 "epoch": "1578512886",
						 "timestamp": "19:48:06",
						 "cluster": "8165307e-6130-4581-942d-20fcfc4e795d"
					}]`)))),
				},
			),
			kafkaReader: defaultKafkaReader,
			expected:    response.Error(http.StatusServiceUnavailable, "HRI Service Temporarily Unavailable | error Detail: ElasticSearch status: NONE/NotReported, clusterId: 8165307e-6130-4581-942d-20fcfc4e795d, unixTimestamp: 1578512886"),
		},
		{
			name: "invalid-ES-response-missing-cluster-or-epoch-field",
			transport: test.NewFakeTransport(t).AddCall(
				"/_cat/health",
				test.ElasticCall{
					ResponseBody: test.ReaderToString(ioutil.NopCloser(bytes.NewReader([]byte(`
					[{
						 "status": "red",
						 "node.total": "3",
						 "node.data": "3",
						 "shards": "5",
						 "pri": "1",
						 "relo": "0",
						 "init": "0",
						 "unassign": "2",
						 "pending_tasks": "4",
						 "max_task_wait_time": "-",
						 "active_shards_percent": "50.0%"
					}]`)))),
				},
			),
			kafkaReader: defaultKafkaReader,
			expected:    response.Error(http.StatusServiceUnavailable, "HRI Service Temporarily Unavailable | error Detail: ElasticSearch status: red, clusterId: NotReported, unixTimestamp: NotReported"),
		},
		{
			name: "bad-ES-response-body-EOF",
			transport: test.NewFakeTransport(t).AddCall(
				"/_cat/health",
				test.ElasticCall{
					ResponseErr: errors.New("client error"),
				},
			),
			kafkaReader: defaultKafkaReader,
			expected:    response.Error(http.StatusInternalServerError, "Elastic client error: client error"),
		},
		{
			name: "ES-client-error",
			transport: test.NewFakeTransport(t).AddCall(
				"/_cat/health",
				test.ElasticCall{
					ResponseBody: test.ReaderToString(ioutil.NopCloser(bytes.NewReader([]byte(``)))),
				},
			),
			kafkaReader: defaultKafkaReader,
			expected: response.Error(
				http.StatusInternalServerError,
				"Error parsing the Elastic search response body: EOF"),
		},
		{
			name: "Kafka-connection-returns-err",
			transport: test.NewFakeTransport(t).AddCall(
				"/_cat/health",
				test.ElasticCall{
					ResponseBody: test.ReaderToString(ioutil.NopCloser(bytes.NewReader([]byte(`
					[{
						"epoch": "1578512886",
						"timestamp": "19:48:06",
						"cluster": "8165307e-6130-4581-942d-20fcfc4e795d",
						"status": "green",
						"node.total": "3",
						"node.data": "3",
						"shards": "19",
						"pri": "9",
						"relo": "0",
						"init": "0",
						"unassign": "0",
						"pending_tasks": "0",
						"max_task_wait_time": "-",
						"active_shards_percent": "100.0%"
					}]`)))),
				},
			),
			kafkaReader: test.FakePartitionReader{
				T:          t,
				Partitions: nil,
				Err:        errors.New("Error contacting Kafka cluster: could not read partitions"),
			},
			expected: response.Error(http.StatusServiceUnavailable, "HRI Service Temporarily Unavailable | error Detail: Kafka status: Kafka Connection/Read Partition failed"),
		},
		{
			name: "Kafka-returns-Err-AND-ES-return-bad-status",
			transport: test.NewFakeTransport(t).AddCall(
				"/_cat/health",
				test.ElasticCall{
					ResponseBody: test.ReaderToString(ioutil.NopCloser(bytes.NewReader([]byte(`
					[{
						 "epoch": "1578512886",
						 "timestamp": "19:48:06",
						 "cluster": "8165307e-6130-4581-942d-20fcfc4e795d",
						 "status": "red",
						 "node.total": "3",
						 "node.data": "3",
						 "shards": "5",
						 "pri": "1",
						 "relo": "0",
						 "init": "0",
						 "unassign": "2",
						 "pending_tasks": "4",
						 "max_task_wait_time": "-",
						 "active_shards_percent": "50.0%"
					}]`)))),
				},
			),
			kafkaReader: test.FakePartitionReader{
				T:          t,
				Partitions: nil,
				Err:        errors.New("Error contacting Kafka cluster: could not read partitions"),
			},
			expected: response.Error(http.StatusServiceUnavailable, "HRI Service Temporarily Unavailable | error Detail: ElasticSearch status: red, clusterId: 8165307e-6130-4581-942d-20fcfc4e795d, unixTimestamp: 1578512886| Kafka status: Kafka Connection/Read Partition failed"),
		},
	}

	for _, tc := range testCases {

		client, err := elastic.ClientFromTransport(tc.transport)
		if err != nil {
			t.Error(err)
		}

		t.Run(tc.name, func(t *testing.T) {
			if actual := Get(tc.args, client, tc.kafkaReader); !reflect.DeepEqual(tc.expected, actual) {
				//notify/print error event as test result
				t.Errorf("HealthCheck-Get(): %v, expected: %v", actual, tc.expected)
			}
		})
	}

}
