/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package elastic

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/IBM/resource-controller-go-sdk-generator/build/generated"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"ibm.com/watson/health/foundation/hri/common/config"
	"ibm.com/watson/health/foundation/hri/common/logwrapper"
	"ibm.com/watson/health/foundation/hri/common/test"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
)

const MOCK_RESPONSE string = "mocked response"

type mockTransport struct{}

func (t *mockTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	return &http.Response{Body: ioutil.NopCloser(strings.NewReader(MOCK_RESPONSE))}, nil
}

func TestResponseError(t *testing.T) {
	logwrapper.Initialize("error", os.Stdout)
	var prefix = "client_test/testResponseError"
	var logger = logwrapper.GetMyLogger("", prefix)
	responseMessage := "response message"

	testCases := []struct {
		name  string
		code  int
		error error
	}{
		{
			name:  "Elastic Error Code and Error in Response",
			code:  http.StatusNotFound,
			error: fmt.Errorf("error message"),
		},
		{
			name: "Elastic Error Code and No Error in Response",
			code: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := &ResponseError{ErrorObj: tc.error, Code: tc.code}

			errorDetail := actual.LogAndBuildErrorDetail("requestId", logger, responseMessage)
			assert.Equal(t, "requestId", errorDetail.ErrorEventId)
			if tc.error == nil {
				assert.Equal(t, fmt.Sprintf("%s: [%d] unexpected Elasticsearch %d error", responseMessage, tc.code, tc.code), errorDetail.ErrorDescription)
			} else {
				assert.Equal(t, fmt.Sprintf("%s: [%d] %v", responseMessage, tc.code, tc.error), errorDetail.ErrorDescription)
			}
		})
	}
}

func TestClientFromTransport(t *testing.T) {
	client, err := ClientFromTransport(&mockTransport{})
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Search()
	if err != nil {
		t.Fatal(err)
	}

	// don't forget to close response body once we're done with it
	defer resp.Body.Close()

	// read response body into a buffer
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if buf.String() != MOCK_RESPONSE {
		t.Errorf("Unexpected response body. Expected: [%v], Actual: [%v]", MOCK_RESPONSE, buf.String())
	}
}

func TestClientFromConfig(t *testing.T) {
	conf, err := config.GetConfig(test.FindConfigPath(t), nil)
	if err != nil {
		t.Error(err)
	}
	client, err := ClientFromConfig(conf)
	if client == nil {
		t.Fatal("Returned Elastic client was nil")
	}
	if err != nil {
		t.Fatalf("ClientFromConfig returned an error: %s", err.Error())
	}
}

func TestCreateResourceControllerService(t *testing.T) {
	conf := generated.NewConfiguration()
	client := generated.NewAPIClient(conf)
	serviceTest := client.ResourceInstancesApi
	service := CreateResourceControllerService()
	assert.Equal(t, serviceTest, service)
}

func TestCheckElasticBearerTokenSuccess(t *testing.T) {
	crn := "myElasticCrn"
	bearerToken := "myBearerToken"
	controller := gomock.NewController(t)
	defer controller.Finish()
	mockResourceInstanceService := test.NewMockResourceControllerService(controller)
	mockResourceInstanceService.
		EXPECT().
		GetResourceInstance(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(generated.ResourceInstance{}, &http.Response{StatusCode: http.StatusOK}, nil).AnyTimes()

	code, err := CheckElasticIAM(crn, bearerToken, mockResourceInstanceService)
	if code != http.StatusOK || err != nil {
		t.Fatal(err)
	}
}

func TestCheckElasticBearerTokenNon200Response(t *testing.T) {
	crn := "myElasticCrn"
	bearerToken := "myBearerToken"
	errMsg := "getResourceInstanceErrMsg"
	controller := gomock.NewController(t)
	defer controller.Finish()
	mockResourceInstanceService := test.NewMockResourceControllerService(controller)
	mockResourceInstanceService.
		EXPECT().
		GetResourceInstance(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(generated.ResourceInstance{}, &http.Response{StatusCode: http.StatusUnauthorized}, errors.New(errMsg)).AnyTimes()

	code, err := CheckElasticIAM(crn, bearerToken, mockResourceInstanceService)
	if code != http.StatusUnauthorized || err == nil || err.Error() != "elastic IAM authentication returned 401 : "+errMsg {
		t.Errorf("CheckElasticIAM() = %v, expected: Resource controller returned status of:", err)
	}
}

func TestCheckElasticBearerTokenFailures(t *testing.T) {
	crn := "myElasticCrn"
	bearerToken := "myBearerToken"
	errMsg := "getResourceInstanceErrMsg"

	controller := gomock.NewController(t)
	defer controller.Finish()
	mockResourceInstanceService := test.NewMockResourceControllerService(controller)
	mockResourceInstanceService.
		EXPECT().
		GetResourceInstance(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(generated.ResourceInstance{}, &http.Response{StatusCode: http.StatusNotFound}, errors.New(errMsg)).AnyTimes()

	code, err := CheckElasticIAM(crn, bearerToken, mockResourceInstanceService)
	if code != http.StatusInternalServerError || err == nil || err.Error() != "elastic IAM authentication returned 404 : "+errMsg {
		t.Errorf("CheckElasticIAM() = %v, expected: 404", err)
	}
}

func TestEncodeQueryBody(t *testing.T) {
	tests := []struct {
		name           string
		queryBody      map[string]interface{}
		expectedErrMsg string
	}{
		{
			name: "successful encoding",
			queryBody: map[string]interface{}{
				"script": map[string]interface{}{
					"source": "elastic update script",
				},
			},
			expectedErrMsg: "",
		},
		{
			name: "error on encoding failure",
			queryBody: map[string]interface{}{
				// channel is an unsupported type for JSON marshalling
				"string": make(chan int),
			},
			expectedErrMsg: "Unable to encode query body as byte buffer: json: unsupported type: chan int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := EncodeQueryBody(tt.queryBody)
			if err != nil && err.Error() != tt.expectedErrMsg {
				t.Errorf("EncodeQueryBody() = %v, expected: %s", err, tt.expectedErrMsg)
			}
		})
	}
}

func TestFromConfigError(t *testing.T) {
	hosts := make([]string, 2)
	hosts[0] = "test-1234-blah.bblah.databases.appdomain.cloud:30900"
	hosts[1] = "test-9876-blah.bblah.databases.appdomain.cloud:30900"

	config := elasticsearch.Config{
		Addresses: hosts,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: nil,
			},
		},
	}

	client, _ := fromConfig(config)
	assert.NotNil(t, client)
}

func TestTenantIdFromIndex(t *testing.T) {
	tenant1 := "fakeTenant-batches"
	rtnIndex := TenantIdFromIndex(tenant1)
	assert.Equal(t, "fakeTenant", rtnIndex)
}

func TestTenantsFromIndices(t *testing.T) {
	tenant1 := make(map[string]interface{})
	tenant2 := make(map[string]interface{})
	tenant3 := make(map[string]interface{})
	tenant1["index"] = "fakeTenant-batches"
	tenant2["index"] = "fakeTenant-notbatches"
	tenant3["index"] = "fakeTenant2-batches"
	tenants := []map[string]interface{}{tenant1, tenant2, tenant3}
	resultsMap := TenantsFromIndices(tenants)
	id := make(map[string]interface{})
	id2 := make(map[string]interface{})
	var expIndices []interface{}
	id["id"] = "fakeTenant"
	expIndices = append(expIndices, id)
	id2["id"] = "fakeTenant2"
	expIndices = append(expIndices, id2)
	expResultsMap := make(map[string]interface{})
	expResultsMap["results"] = expIndices
	assert.Equal(t, expResultsMap, resultsMap)
}
