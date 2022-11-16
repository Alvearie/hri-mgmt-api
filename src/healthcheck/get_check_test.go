/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package healthcheck

const requestId string = "testRequestId"

/*
func TestHealthcheck(t *testing.T) {
	logwrapper.Initialize("error", os.Stdout)

	testCases := []struct {
		name               string
		transport          *test.FakeTransport
		kafkaHealthChecker kafka.HealthChecker
		expectedCode       int
		expectedBody       *response.ErrorDetail
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
			kafkaHealthChecker: fakeKafkaHealthChecker{},
			expectedCode:       http.StatusOK,
			expectedBody:       nil,
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
			kafkaHealthChecker: fakeKafkaHealthChecker{},
			expectedCode:       http.StatusServiceUnavailable,
			expectedBody:       response.NewErrorDetail(requestId, "HRI Service Temporarily Unavailable | error Detail: ElasticSearch status: red, clusterId: 8165307e-6130-4581-942d-20fcfc4e795d, unixTimestamp: 1578512886"),
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
			kafkaHealthChecker: fakeKafkaHealthChecker{},
			expectedCode:       http.StatusServiceUnavailable,
			expectedBody:       response.NewErrorDetail(requestId, "HRI Service Temporarily Unavailable | error Detail: ElasticSearch status: NONE/NotReported, clusterId: 8165307e-6130-4581-942d-20fcfc4e795d, unixTimestamp: 1578512886"),
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
			kafkaHealthChecker: fakeKafkaHealthChecker{},
			expectedCode:       http.StatusServiceUnavailable,
			expectedBody:       response.NewErrorDetail(requestId, "HRI Service Temporarily Unavailable | error Detail: ElasticSearch status: red, clusterId: NotReported, unixTimestamp: NotReported"),
		},
		{
			name: "ES-client-error",
			transport: test.NewFakeTransport(t).AddCall(
				"/_cat/health",
				test.ElasticCall{
					ResponseErr: errors.New("client error"),
				},
			),
			kafkaHealthChecker: fakeKafkaHealthChecker{},
			expectedCode:       http.StatusServiceUnavailable,
			expectedBody: response.NewErrorDetail(requestId,
				"Could not perform elasticsearch health check: [500] elasticsearch client error: client error"),
		},
		{
			name: "Kafka-health-check-returns-err",
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
			kafkaHealthChecker: fakeKafkaHealthChecker{
				err: errors.New("ResponseError contacting Kafka cluster: could not read partitions"),
			},
			expectedCode: http.StatusServiceUnavailable,
			expectedBody: response.NewErrorDetail(requestId, "HRI Service Temporarily Unavailable | error Detail: ResponseError contacting Kafka cluster: could not read partitions"),
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
			kafkaHealthChecker: fakeKafkaHealthChecker{
				err: errors.New("ResponseError contacting Kafka cluster: could not read partitions"),
			},
			expectedCode: http.StatusServiceUnavailable,
			expectedBody: response.NewErrorDetail(requestId, "HRI Service Temporarily Unavailable | error Detail: ElasticSearch status: red, clusterId: 8165307e-6130-4581-942d-20fcfc4e795d, unixTimestamp: 1578512886 | ResponseError contacting Kafka cluster: could not read partitions"),
		},
	}

	for _, tc := range testCases {
		client, err := elastic.ClientFromTransport(tc.transport)
		if err != nil {
			t.Error(err)
		}

		t.Run(tc.name, func(t *testing.T) {
			actualCode, actualBody := Get(requestId, client, tc.kafkaHealthChecker)
			if actualCode != tc.expectedCode || !reflect.DeepEqual(tc.expectedBody, actualBody) {
				//notify/print error event as test result
				t.Errorf("HealthCheck-Get()\n   actual: %v,%v\n expected: %v,%v", actualCode, actualBody, tc.expectedCode, tc.expectedBody)
			}
		})
	}
}

type fakeKafkaHealthChecker struct {
	err error
}

func (fhc fakeKafkaHealthChecker) Check() error {
	return fhc.err
}

func (fhc fakeKafkaHealthChecker) Close() {
	return
}*/
