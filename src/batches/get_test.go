/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package batches

import (
	"errors"
	"github.com/Alvearie/hri-mgmt-api/common/auth"
	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/Alvearie/hri-mgmt-api/common/path"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	"net/http"
	"os"
	"reflect"
	"testing"
)

func TestGet(t *testing.T) {
	activationId := "activationId"
	subject := "clientId"
	_ = os.Setenv(response.EnvOwActivationId, activationId)

	tests := []struct {
		name     string
		params   map[string]interface{}
		claims   auth.HriClaims
		ft       *test.FakeTransport
		expected map[string]interface{}
	}{
		{"simple",
			map[string]interface{}{path.ParamOwPath: "/hri/tenants/1234/batches"},
			auth.HriClaims{Scope: auth.HriConsumer},
			test.NewFakeTransport(t).AddCall(
				"/1234-batches/_search",
				test.ElasticCall{
					RequestQuery: "from=0&size=10&track_total_hits=true",
					ResponseBody: `
						{
							"hits":{
								"total" : {
								  "value" : 1,
								  "relation" : "eq"
 								},
								"hits":[
									{
										"_id":"uuid",
										"_source":{
											"dataType" : "rspec-batch",
											"invalidThreshold" : -1,
											"metadata" : {
												"rspec1" : "test1"
											},
											"name" : "mybatch",
											"startDate" : "2021-02-24T18:08:36Z",
											"status" : "started",
											"topic" : "ingest.test.claims.in",
											"integratorId" : "modified-integrator-id"
										}
									}
								]
							}
						}`,
				},
			),
			response.Success(http.StatusOK, map[string]interface{}{"total": float64(1), "results": []interface{}{map[string]interface{}{"id": "uuid", "dataType": "rspec-batch", "invalidThreshold": float64(-1), "name": "mybatch", "startDate": "2021-02-24T18:08:36Z", "status": "started", "topic": "ingest.test.claims.in", "integratorId": "modified-integrator-id", "metadata": map[string]interface{}{"rspec1": "test1"}}}})},
		{"allparams",
			map[string]interface{}{path.ParamOwPath: "/hri/tenants/1234/batches", "size": "20", "from": "10", "name": "mybatch", "status": "started", "gteDate": "01/01/2019", "lteDate": "01/01/2020"},
			auth.HriClaims{Scope: auth.HriConsumer},
			test.NewFakeTransport(t).AddCall(
				"/1234-batches/_search",
				test.ElasticCall{
					RequestQuery: "from=10&size=20&track_total_hits=true",
					// Note that ] and [ must be escaped because RequestBody is used as a regex pattern
					RequestBody: `{"query":{"bool":{"must":\[{"term":{"name":"mybatch"}},{"term":{"status":"started"}},{"range":{"startDate":{"gte":"01/01/2019","lte":"01/01/2020"}}}\]}}}` + "\n",
					ResponseBody: `
						{
							"hits":{
								"total" : {
								  "value" : 1,
								  "relation" : "eq"
								},
								"hits":[
									{
										"_id":"uuid",
										"_source":{
 											"dataType" : "rspec-batch",
											"invalidThreshold" : -1,
											"metadata" : {
												"rspec1" : "test1"
											},
											"name" : "mybatch",
											"startDate" : "01/02/2019",
											"status" : "started",
											"topic" : "ingest.test.claims.in",
											"integratorId" : "modified-integrator-id"
										}
									}
								]
							}
						}`,
				},
			),
			response.Success(http.StatusOK, map[string]interface{}{"total": float64(1), "results": []interface{}{map[string]interface{}{"id": "uuid", "dataType": "rspec-batch", "invalidThreshold": float64(-1), "name": "mybatch", "startDate": "01/02/2019", "status": "started", "topic": "ingest.test.claims.in", "integratorId": "modified-integrator-id", "metadata": map[string]interface{}{"rspec1": "test1"}}}})},
		{"missing open whisk path param",
			map[string]interface{}{},
			auth.HriClaims{Scope: auth.HriConsumer},
			test.NewFakeTransport(t),
			response.Error(http.StatusBadRequest, "Required parameter '__ow_path' is missing"),
		},
		{"bad open whisk path param",
			map[string]interface{}{path.ParamOwPath: "/hri/tenants"},
			auth.HriClaims{Scope: auth.HriConsumer},
			test.NewFakeTransport(t),
			response.Error(http.StatusBadRequest, "The path is shorter than the requested path parameter; path: [ hri tenants], requested index: 3"),
		},
		{"bad size param",
			map[string]interface{}{path.ParamOwPath: "/hri/tenants/1234/batches", "size": "a1"},
			auth.HriClaims{Scope: auth.HriConsumer},
			test.NewFakeTransport(t),
			response.Error(http.StatusBadRequest, "Error parsing 'size' parameter: strconv.Atoi: parsing \"a1\": invalid syntax"),
		},
		{"bad from param",
			map[string]interface{}{path.ParamOwPath: "/hri/tenants/1234/batches", "from": "b2"},
			auth.HriClaims{Scope: auth.HriConsumer},
			test.NewFakeTransport(t),
			response.Error(http.StatusBadRequest, "Error parsing 'from' parameter: strconv.Atoi: parsing \"b2\": invalid syntax"),
		},
		{"bad name param_prohibited character",
			map[string]interface{}{path.ParamOwPath: "/hri/tenants/1234/batches", "name": "{[]//zzx[]}"},
			auth.HriClaims{Scope: auth.HriConsumer},
			test.NewFakeTransport(t),
			response.Error(http.StatusBadRequest, "query parameters may not contain these characters: \"[]{}"),
		},
		{"bad status param_prohibited character",
			map[string]interface{}{path.ParamOwPath: "/hri/tenants/1234/batches", "status": "z{[z]//j[]}"},
			auth.HriClaims{Scope: auth.HriConsumer},
			test.NewFakeTransport(t),
			response.Error(http.StatusBadRequest, "query parameters may not contain these characters: \"[]{}"),
		},
		{"bad startDate param_prohibited character",
			map[string]interface{}{path.ParamOwPath: "/hri/tenants/1234/batches", "gteDate": "01/01/2019", "lteDate": "{}[][xxx]\"}"},
			auth.HriClaims{Scope: auth.HriConsumer},
			test.NewFakeTransport(t),
			response.Error(http.StatusBadRequest, "query parameters may not contain these characters: \"[]{}"),
		},
		{"client error",
			map[string]interface{}{path.ParamOwPath: "/hri/tenants/1234/batches"},
			auth.HriClaims{Scope: auth.HriConsumer},
			test.NewFakeTransport(t).AddCall(
				"/1234-batches/_search",
				test.ElasticCall{
					RequestQuery: "from=0&size=10&track_total_hits=true",
					ResponseErr:  errors.New("client error"),
				},
			),
			response.Error(http.StatusInternalServerError,
				"Could not retrieve batches: elasticsearch client error: client error"),
		},
		{"Missing scopes",
			map[string]interface{}{path.ParamOwPath: "/hri/tenants/1234/batches"},
			auth.HriClaims{},
			test.NewFakeTransport(t),
			response.Error(http.StatusUnauthorized, auth.MsgAccessTokenMissingScopes),
		},
		{"Integrator filter",
			map[string]interface{}{path.ParamOwPath: "/hri/tenants/1234/batches"},
			auth.HriClaims{Scope: auth.HriIntegrator, Subject: subject},
			test.NewFakeTransport(t).AddCall(
				"/1234-batches/_search",
				test.ElasticCall{
					RequestQuery: "from=0&size=10&track_total_hits=true",
					RequestBody:  `{"query":{"bool":{"must":\[{"term":{"integratorId":"clientId"}}\]}}}` + "\n",
					ResponseBody: `
						{
							"hits":{
								"total" : {
								  "value" : 1,
								  "relation" : "eq"
								},
								"hits":[
									{
										"_id":"uuid",
										"_source":{
											"dataType" : "rspec-batch",
											"invalidThreshold" : -1,
											"metadata" : {
												"rspec1" : "test1"
											},
											"name" : "mybatch",
											"startDate" : "2021-02-24T18:08:36Z",
											"status" : "started",
											"topic" : "ingest.test.claims.in",
											"integratorId" : "modified-integrator-id"
										}
									}
								]
							}
						}`,
				},
			),
			response.Success(http.StatusOK, map[string]interface{}{"total": float64(1), "results": []interface{}{map[string]interface{}{"id": "uuid", "dataType": "rspec-batch", "invalidThreshold": float64(-1), "name": "mybatch", "startDate": "2021-02-24T18:08:36Z", "status": "started", "topic": "ingest.test.claims.in", "integratorId": "modified-integrator-id", "metadata": map[string]interface{}{"rspec1": "test1"}}}})},
		{"Consumer & Integrator no filter",
			map[string]interface{}{path.ParamOwPath: "/hri/tenants/1234/batches"},
			auth.HriClaims{Scope: auth.HriIntegrator + " " + auth.HriConsumer, Subject: subject},
			test.NewFakeTransport(t).AddCall(
				"/1234-batches/_search",
				test.ElasticCall{
					RequestQuery: "from=0&size=10&track_total_hits=true",
					ResponseBody: `
						{
							"hits":{
								"total" : {
								  "value" : 1,
								  "relation" : "eq"
								},
								"hits":[
									{
										"_id":"uuid",
										"_source":{
											"dataType" : "rspec-batch",
											"invalidThreshold" : -1,
											"metadata" : {
												"rspec1" : "test1"
											},
											"name" : "mybatch",
											"startDate" : "2021-02-24T18:08:36Z",
											"status" : "started",
											"topic" : "ingest.test.claims.in",
											"integratorId" : "modified-integrator-id"
										}
									}
								]
							}
						}`,
				},
			),
			response.Success(http.StatusOK, map[string]interface{}{"total": float64(1), "results": []interface{}{map[string]interface{}{"id": "uuid", "dataType": "rspec-batch", "invalidThreshold": float64(-1), "name": "mybatch", "startDate": "2021-02-24T18:08:36Z", "status": "started", "topic": "ingest.test.claims.in", "integratorId": "modified-integrator-id", "metadata": map[string]interface{}{"rspec1": "test1"}}}})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			esClient, err := elastic.ClientFromTransport(tt.ft)
			if err != nil {
				t.Error(err)
			}

			if got := Get(tt.params, tt.claims, esClient); !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("Get() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func Test_appendRange(t *testing.T) {
	type args struct {
		params    map[string]interface{}
		paramName string
		gteParam  string
		lteParam  string
		clauses   *[]map[string]interface{}
	}
	tests := []struct {
		name     string
		args     args
		expected map[string]interface{}
		err      error
	}{
		{"gte & lte",
			args{params: map[string]interface{}{"gteDate": "01/01/2019", "lteDate": "01/01/2020"}, paramName: "startDate", gteParam: "gteDate", lteParam: "lteDate", clauses: &([]map[string]interface{}{})},
			map[string]interface{}{"range": map[string]interface{}{"startDate": map[string]interface{}{"gte": "01/01/2019", "lte": "01/01/2020"}}},
			nil,
		},
		{"gte",
			args{params: map[string]interface{}{"gteDate": "01/01/2019"}, paramName: "startDate", gteParam: "gteDate", lteParam: "lteDate", clauses: &([]map[string]interface{}{})},
			map[string]interface{}{"range": map[string]interface{}{"startDate": map[string]interface{}{"gte": "01/01/2019"}}},
			nil,
		},
		{"lte",
			args{params: map[string]interface{}{"lteDate": "01/01/2020"}, paramName: "startDate", gteParam: "gteDate", lteParam: "lteDate", clauses: &([]map[string]interface{}{})},
			map[string]interface{}{"range": map[string]interface{}{"startDate": map[string]interface{}{"lte": "01/01/2020"}}},
			nil,
		},
		{"none",
			args{params: map[string]interface{}{}, paramName: "startDate", gteParam: "gteDate", lteParam: "lteDate", clauses: &([]map[string]interface{}{})},
			nil,
			nil,
		},
		{"prohibited characters",
			args{params: map[string]interface{}{"gteDate": "{}"}, paramName: "startDate", gteParam: "gteDate", lteParam: "lteDate", clauses: &([]map[string]interface{}{})},
			nil,
			errors.New("query parameters may not contain these characters: \"[]{}"),
		},
		{"prohibited characters-lte",
			args{params: map[string]interface{}{"lteDate": "{[][]}}"}, paramName: "startDate", gteParam: "gteDate", lteParam: "lteDate", clauses: &([]map[string]interface{}{})},
			nil,
			errors.New("query parameters may not contain these characters: \"[]{}"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := appendRange(tt.args.params, tt.args.paramName, tt.args.gteParam, tt.args.lteParam, tt.args.clauses)
			if tt.err != err && tt.err.Error() != err.Error() {
				t.Errorf("Expected error doesn't match: expected: %s, got: %s", tt.err, err)
			} else if len(*tt.args.clauses) != 1 || !reflect.DeepEqual((*tt.args.clauses)[0], tt.expected) {
				if tt.expected == nil && len(*tt.args.clauses) != 0 {
					t.Errorf("excepted: nil, got: %v", *tt.args.clauses)
				} else if tt.expected != nil {
					t.Errorf("excepted: %v, got: %v", tt.expected, (*tt.args.clauses)[0])
				}
			}
		})
	}
}

func Test_appendTermPresent(t *testing.T) {
	params := map[string]interface{}{"name": "mybatch"}
	param := "name"
	clauses := make([]map[string]interface{}, 0, 1)

	err := appendTerm(params, param, &clauses)
	if err != nil {
		t.Error(err)
	}
	expected := map[string]interface{}{"term": map[string]interface{}{"name": "mybatch"}}

	if len(clauses) != 1 || !reflect.DeepEqual(clauses[0], expected) {
		t.Errorf("excepted: %v, got: %v", expected, clauses[0])
	}
}

func Test_appendTermMissing(t *testing.T) {
	params := map[string]interface{}{"name": "mybatch"}
	param := "status"
	clauses := make([]map[string]interface{}, 0, 1)

	err := appendTerm(params, param, &clauses)
	if err != nil {
		t.Error(err)
	}

	if len(clauses) != 0 {
		t.Errorf("excepted no terms to be added, got: %v", clauses)
	}
}

func Test_appendTermProhibitedCharacter(t *testing.T) {
	params := map[string]interface{}{"name": "mybatch\",\"key2\":\"value2\""}
	param := "name"
	clauses := make([]map[string]interface{}, 0, 1)

	err := appendTerm(params, param, &clauses)
	if err == nil || err.Error() != "query parameters may not contain these characters: \"[]{}" {
		t.Errorf("Expected error: query parameters may not contain these characters: \"[]{}, got: %v", err)
	}
}
