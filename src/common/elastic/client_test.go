/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package elastic

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	"github.com/IBM/resource-controller-go-sdk-generator/build/generated"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

const MOCK_RESPONSE string = "mocked response"

type mockTransport struct{}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{Body: ioutil.NopCloser(strings.NewReader(MOCK_RESPONSE))}, nil
}

func TestClientFromTransport(t *testing.T) {
	client, err := ClientFromTransport(&mockTransport{})
	if err != nil {
		t.Fatal(err)
	}

	response, err := client.Search()
	if err != nil {
		t.Fatal(err)
	}

	// don't forget to close response body once we're done with it
	defer response.Body.Close()

	// read response body into a buffer
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	if buf.String() != MOCK_RESPONSE {
		t.Errorf("Unexpected response body. Expected: [%v], Actual: [%v]", MOCK_RESPONSE, buf.String())
	}
}

func TestClientFromParamsSuccess(t *testing.T) {
	jsonParams := `{
		  "__bx_creds": {
			"databases-for-elasticsearch": {
			  "connection": {
				"https": {
				  "authentication": {
					"method": "direct",
					"password": "elasticPassword",
					"username": "elasticUser"
				  },
				  "certificate": {
					"certificate_base64": "elasticBase64EncodedCert",
					"name": "0aa7713d-b76a-11e9-8963-7a53aa8a4b28"
				  },
				  "composed": [
					"https://elasticUser:elasticPassword@8165307e-6130-4581-942d-20fcfc4e795d.bkvfvtld0lmh0umkfi70.databases.appdomain.cloud:30600"
				  ],
				  "hosts": [
					{
					  "hostname": "8165307e-6130-4581-942d-20fcfc4e795d.bkvfvtld0lmh0umkfi70.databases.appdomain.cloud",
					  "port": 30600
					}
				  ],
				  "path": "",
				  "query_options": {},
				  "scheme": "https",
				  "type": "uri"
				}
			  }
			}
		  }
		}`

	var params map[string]interface{}
	if err := json.Unmarshal([]byte(jsonParams), &params); err != nil {
		t.Fatal(err)
	}

	_, err := ClientFromParams(params)
	if err != nil {
		t.Fatal(err)
	}

	// Used for manual tests using real credentials
	//res,_ := client.Cat.Indices(client.Cat.Indices.WithPretty())
	//fmt.Print(res)
}

func TestClientFromParamsFailures(t *testing.T) {

	tests := []struct {
		name       string
		jsonParams string
		expected   string
	}{
		{"missing credential parameter",
			`{"name": "value"}`,
			"error extracting the __bx_creds section of the JSON",
		},
		{"missing https section",
			`{"__bx_creds": {
				"databases-for-elasticsearch": {
					"connection":{
					}
				}
			}}`,
			"error extracting the https section of the JSON",
		},
		{"missing certificate param",
			`{"__bx_creds": {
				"databases-for-elasticsearch": {
					"connection":{
					  "https": {
						"composed": [
						  "https://elasticUser:elasticPassword@8165307e-6130-4581-942d-20fcfc4e795d.bkvfvtld0lmh0umkfi70.databases.appdomain.cloud:30600"
						]
					  }
					}
				}
			}}`,
			"error extracting the certificate section of the JSON",
		},
		{"missing certificate",
			`{"__bx_creds": {
				"databases-for-elasticsearch": {
					"connection":{
					  "https": {
						"certificate": {
						  "name": "elasticBase64EncodedCert"
						},
						"composed": [
						  "https://elasticUser:elasticPassword@8165307e-6130-4581-942d-20fcfc4e795d.bkvfvtld0lmh0umkfi70.databases.appdomain.cloud:30600"
						]
					  }
					}
				}
			}}`,
			"error extracting the certificate from Elastic credentials",
		},
		{"missing composed urls",
			`{"__bx_creds": {
				"databases-for-elasticsearch": {
					"connection":{
					  "https": {
						"certificate": {
						  "certificate_base64": "elasticBase64EncodedCert"
						}
					  }
					}
				}
			}}`,
			"error extracting the host from Elastic credentials",
		},
		{"error_non_base64_cert",
			`{
				  "__bx_creds": {
					"databases-for-elasticsearch": {
					  "connection": {
						"https": {
						  "authentication": {
							"method": "direct",
							"password": "elasticPassword",
							"username": "elasticUser"
						  },
						  "certificate": {
							"certificate_base64": "AAA=\n!!"
						  },
						  "composed": [
							"https://elasticUser:elasticPassword@8165307e-6130-4581-942d-20fcfc4e795d.bkvfvtld0lmh0umkfi70.databases.appdomain.cloud:30600"
						  ],
						  "hosts": [
							{
							  "hostname": "8165307e-6130-4581-942d-20fcfc4e795d.bkvfvtld0lmh0umkfi70.databases.appdomain.cloud",
							  "port": 30600
							}
						  ]
						}
					  }
					}
				  }
			}`,
			"illegal base64 data at input byte 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var params map[string]interface{}
			if err := json.Unmarshal([]byte(tt.jsonParams), &params); err != nil {
				t.Fatal(err)
			}

			_, err := ClientFromParams(params)
			if err == nil || err.Error() != tt.expected {
				t.Errorf("ClientFromParams() = %v, expected: %s", err, tt.expected)
			}
		})
	}
}

func TestCreateResourceControllerService(t *testing.T) {
	conf := generated.NewConfiguration()
	client := generated.NewAPIClient(conf)
	servicetest := client.ResourceInstancesApi
	service := CreateResourceControllerService()
	assert.Equal(t, servicetest, service)
}

func TestCheckElasticBearerTokenSuccess(t *testing.T) {
	jsonParams := `{
		  "__bx_creds": {
			"databases-for-elasticsearch": {
			  "instance_administration_api": {
				  "deployment_id": "fake_deployment_id"
			  }
			}
		  },
          "__ow_headers": {
			"authorization": "Bearer 1234"
	      }
		}`

	var params map[string]interface{}
	if err := json.Unmarshal([]byte(jsonParams), &params); err != nil {
		t.Fatal(err)
	}
	controller := gomock.NewController(t)
	defer controller.Finish()
	mockResourceInstanceService := test.NewMockResourceControllerService(controller)
	mockResourceInstanceService.
		EXPECT().
		GetResourceInstance(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(generated.ResourceInstance{}, &http.Response{StatusCode: http.StatusOK}, nil).AnyTimes()

	_, err := CheckElasticIAM(params, mockResourceInstanceService)
	if err != nil {
		t.Fatal(err)
	}

}

func TestCheckElasticBearerTokenNon200Response(t *testing.T) {
	jsonParams := `{
		  "__bx_creds": {
			"databases-for-elasticsearch": {
			  "instance_administration_api": {
				  "deployment_id": "fake_deployment_id"
			  }
			}
		  },
          "__ow_headers": {
			"authorization": "Bearer 1234"
	      }
		}`

	var params map[string]interface{}
	if err := json.Unmarshal([]byte(jsonParams), &params); err != nil {
		t.Fatal(err)
	}
	controller := gomock.NewController(t)
	defer controller.Finish()
	mockResourceInstanceService := test.NewMockResourceControllerService(controller)
	mockResourceInstanceService.
		EXPECT().
		GetResourceInstance(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(generated.ResourceInstance{}, &http.Response{StatusCode: http.StatusUnauthorized}, nil).AnyTimes()

	_, err := CheckElasticIAM(params, mockResourceInstanceService)
	if err == nil || err.Error() != "Resource controller returned status of: " {
		t.Errorf("ClientFromParams() = %v, expected: Resource controller returned status of:", err)
	}

}

func TestCheckElasticBearerTokenFailures(t *testing.T) {
	tests := []struct {
		name       string
		jsonParams string
		expected   string
	}{
		{"bc creds error",
			`{
		  "__bx_cred": {
			"databases-for-elasticsearch": {
			  "instance_administration_api": {
			  }
			}
		  }
		}`,
			"error extracting the __bx_creds section of the JSON",
		},
		{"error extracting auth",
			`{
		  "__bx_creds": {
			"databases-for-elasticsearch": {
			  "instance_administration_api": {
			  }
			}
		  }
		}`,
			"error extracting the deployment Id from administration api",
		},
		{"missing authorization",
			`{
		  "__bx_creds": {
			"databases-for-elasticsearch": {
			  "instance_administration_api": {
				"deployment_id": "fake_deployment_id"
			  }
			}
		  },
          "__ow_headers": {
	      }
		}`,
			"Required parameter authorization is missing",
		},
		{"missing Bearer token",
			`{
		  "__bx_creds": {
			"databases-for-elasticsearch": {
			  "instance_administration_api": {
				  "deployment_id": "fake_deployment_id"
			  }
			}
		  }
		}`,
			"error extracting the __ow_headers section of the JSON",
		},
		{"404 error returned",
			`{
		  "__bx_creds": {
			"databases-for-elasticsearch": {
			  "instance_administration_api": {
				  "deployment_id": "fake_deployment_id"
			  }
			}
		  },
          "__ow_headers": {
			"authorization": "Bearer 1234"
	      }
		}`,
			"404",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var params map[string]interface{}
			if err := json.Unmarshal([]byte(tt.jsonParams), &params); err != nil {
				t.Fatal(err)
			}
			controller := gomock.NewController(t)
			defer controller.Finish()
			mockResourceInstanceService := test.NewMockResourceControllerService(controller)
			mockResourceInstanceService.
				EXPECT().
				GetResourceInstance(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(generated.ResourceInstance{}, &http.Response{StatusCode: http.StatusNotFound}, errors.New("404")).AnyTimes()

			_, err := CheckElasticIAM(params, mockResourceInstanceService)
			if err == nil || err.Error() != tt.expected {
				t.Errorf("ClientFromParams() = %v, expected: %s", err, tt.expected)
			}
		})
	}

}

func TestEncodeQueryBody(t *testing.T) {
	tests := []struct {
		name          string
		queryBody     map[string]interface{}
		expectedError string
	}{
		{"successful encoding",
			map[string]interface{}{
				"script": map[string]interface{}{
					"source": "elastic update script",
				},
			},
			"",
		},
		{"error on encoding failure",
			map[string]interface{}{
				// channel is an unsupported type for JSON marshalling
				"string": make(chan int),
			},
			"Unable to encode query body as byte buffer: json: unsupported type: chan int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := EncodeQueryBody(tt.queryBody)
			if err != nil && err.Error() != tt.expectedError {
				t.Errorf("EncodeQueryBody() = %v, expected: %s", err, tt.expectedError)
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

func TestIndexFromTenantId(t *testing.T) {
	tenant1 := "fakeTenant"
	rtnIndex := IndexFromTenantId(tenant1)
	assert.Equal(t, "fakeTenant-batches", rtnIndex)
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
	resultssMap := TenantsFromIndices(tenants)
	id := make(map[string]interface{})
	id2 := make(map[string]interface{})
	var expIndices []interface{}
	id["id"] = "fakeTenant"
	expIndices = append(expIndices, id)
	id2["id"] = "fakeTenant2"
	expIndices = append(expIndices, id2)
	expResultssMap := make(map[string]interface{})
	expResultssMap["results"] = expIndices
	assert.Equal(t, expResultssMap, resultssMap)
}
