/**
 * (C) Copyright IBM Corp. 2021
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package main

// import (
// 	"errors"
// 	"github.com/Alvearie/hri-mgmt-api/common/model"
// 	"github.com/Alvearie/hri-mgmt-api/common/param"
// 	"github.com/Alvearie/hri-mgmt-api/common/test"
// 	"github.com/labstack/echo/v4"
// 	"github.com/stretchr/testify/assert"
// 	"net/http"
// 	"net/http/httptest"
// 	"runtime/debug"
// 	"strings"
// 	"testing"
// )

// func TestConfigureMgmtServerErrors(t *testing.T) {
// 	configPath := test.FindConfigPath(t)
// 	e := echo.New()

// 	tests := []struct {
// 		name               string
// 		args               []string
// 		expectedError      error
// 		expectedReturnCode int
// 	}{
// 		{
// 			name:               "Bad Config",
// 			expectedReturnCode: 1,
// 			args:               []string{"--validation=notABool"},
// 			expectedError:      errors.New("error parsing commandline args: invalid boolean value \"notABool\" for -validation: parse error"),
// 		},
// 		{
// 			name:               "Bad Log Level",
// 			expectedReturnCode: 3,
// 			args:               []string{"--log-level=notALevel"},
// 			expectedError:      errors.New("ERROR: Could NOT initialize Logger: error parsing log Level - not a valid logrus Level: \"notALevel\""),
// 		},
// 		{
// 			name:               "Bad New Relic License",
// 			expectedReturnCode: 1,
// 			args:               []string{"--new-relic-license-key=notLongEnough"},
// 			expectedError:      errors.New("ERROR CONFIGURING NEW RELIC: license length is not 40"),
// 		},
// 	}

// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			rc, startFunc, err := configureMgmtServer(e, append([]string{"--config-path=" + configPath}, tc.args...))
// 			assert.Nil(t, startFunc)
// 			if tc.expectedError == nil {
// 				assert.NoError(t, err)
// 			} else {
// 				assert.Error(t, err, tc.expectedError)
// 			}

// 			assert.Equal(t, tc.expectedReturnCode, rc)
// 		})
// 	}
// }

// func TestConfigureMgmtServer(t *testing.T) {
// 	configPath := test.FindConfigPath(t)
// 	e := echo.New()

// 	rc, startFunc, err := configureMgmtServer(e, []string{"--config-path=" + configPath})
// 	assert.Equal(t, 0, rc)
// 	assert.NotNil(t, startFunc)
// 	assert.Nil(t, err)

// 	// TODO: some refactor of logwrapper could allow better testing of its usage in serve.go

// 	// Echo doesn't have a method to review/test which middleware has been configured

// 	// Test that the validator has been set to the CustomValidator type
// 	assert.NotNil(t, e.Validator)
// 	assert.IsType(t, e.Validator, &model.CustomValidator{})

// 	// Endpoint for ready/liveness probes
// 	rec := httptest.NewRecorder()
// 	context := e.NewContext(nil, rec)
// 	e.Router().Find(http.MethodGet, "/alive", context)
// 	context.Handler()(context)
// 	assert.Equal(t, rec.Body.String(), "yes")
// }

// func TestMgmtServerRoutes(t *testing.T) {
// 	configPath := test.FindConfigPath(t)
// 	e := echo.New()
// 	configureMgmtServer(e, []string{"--config-path=" + configPath})

// 	var context echo.Context
// 	type routeTestType = struct {
// 		name                    string
// 		method                  string
// 		routePath               string
// 		expectedHandlerFilePath string
// 		expectedPathParameters  map[string]string
// 	}
// 	routeTests := make([]routeTestType, 0)

// 	// Healthcheck Routing
// 	routeTests = append(routeTests, routeTestType{
// 		name:                    "healthcheck",
// 		method:                  http.MethodGet,
// 		routePath:               "/hri/healthcheck",
// 		expectedHandlerFilePath: "healthcheck/handler",
// 	})

// 	// Tenants routing
// 	tenantsHandlerPath := "tenants/handler"
// 	routeTests = append(routeTests, []routeTestType{
// 		{
// 			name:                    "tenants - get all",
// 			method:                  http.MethodGet,
// 			routePath:               "/hri/tenants",
// 			expectedHandlerFilePath: tenantsHandlerPath,
// 		},
// 		{
// 			name:                    "tenants - get by id",
// 			method:                  http.MethodGet,
// 			routePath:               "/hri/tenants/testTenant",
// 			expectedHandlerFilePath: tenantsHandlerPath,
// 			expectedPathParameters: map[string]string{
// 				param.TenantId: "testTenant",
// 			},
// 		},
// 		{
// 			name:                    "tenants - create",
// 			method:                  http.MethodPost,
// 			routePath:               "/hri/tenants/testTenant",
// 			expectedHandlerFilePath: tenantsHandlerPath,
// 			expectedPathParameters: map[string]string{
// 				param.TenantId: "testTenant",
// 			},
// 		},
// 		{
// 			name:                    "tenants - delete",
// 			method:                  http.MethodDelete,
// 			routePath:               "/hri/tenants/testTenant",
// 			expectedHandlerFilePath: tenantsHandlerPath,
// 			expectedPathParameters: map[string]string{
// 				param.TenantId: "testTenant",
// 			},
// 		},
// 	}...)

// 	// Batches routing
// 	batchesHandlerPath := "batches/handler"
// 	routeTests = append(routeTests, []routeTestType{
// 		{
// 			name:                    "batch - get all",
// 			method:                  http.MethodGet,
// 			routePath:               "/hri/tenants/testTenant/batches",
// 			expectedHandlerFilePath: batchesHandlerPath,
// 			expectedPathParameters: map[string]string{
// 				param.TenantId: "testTenant",
// 			},
// 		},
// 		{
// 			name:                    "batch - create",
// 			method:                  http.MethodPost,
// 			routePath:               "/hri/tenants/testTenant/batches",
// 			expectedHandlerFilePath: batchesHandlerPath,
// 			expectedPathParameters: map[string]string{
// 				param.TenantId: "testTenant",
// 			},
// 		},
// 		{
// 			name:                    "batch - get by id",
// 			method:                  http.MethodGet,
// 			routePath:               "/hri/tenants/testTenant/batches/testBatch",
// 			expectedHandlerFilePath: batchesHandlerPath,
// 			expectedPathParameters: map[string]string{
// 				param.TenantId: "testTenant",
// 				param.BatchId:  "testBatch",
// 			},
// 		},
// 		{
// 			name:                    "batch - sendComplete",
// 			method:                  http.MethodPut,
// 			routePath:               "/hri/tenants/testTenant/batches/testBatch/action/sendComplete",
// 			expectedHandlerFilePath: batchesHandlerPath,
// 			expectedPathParameters: map[string]string{
// 				param.TenantId: "testTenant",
// 				param.BatchId:  "testBatch",
// 			},
// 		},
// 		{
// 			name:                    "batch - terminate",
// 			method:                  http.MethodPut,
// 			routePath:               "/hri/tenants/testTenant/batches/testBatch/action/terminate",
// 			expectedHandlerFilePath: batchesHandlerPath,
// 			expectedPathParameters: map[string]string{
// 				param.TenantId: "testTenant",
// 				param.BatchId:  "testBatch",
// 			},
// 		},
// 		{
// 			name:                    "batch - processingComplete",
// 			method:                  http.MethodPut,
// 			routePath:               "/hri/tenants/testTenant/batches/testBatch/action/processingComplete",
// 			expectedHandlerFilePath: batchesHandlerPath,
// 			expectedPathParameters: map[string]string{
// 				param.TenantId: "testTenant",
// 				param.BatchId:  "testBatch",
// 			},
// 		},
// 		{
// 			name:                    "batch - fail",
// 			method:                  http.MethodPut,
// 			routePath:               "/hri/tenants/testTenant/batches/testBatch/action/fail",
// 			expectedHandlerFilePath: batchesHandlerPath,
// 			expectedPathParameters: map[string]string{
// 				param.TenantId: "testTenant",
// 				param.BatchId:  "testBatch",
// 			},
// 		},
// 	}...)

// 	// Streams routing
// 	streamsHandlerPath := "streams/handler"
// 	routeTests = append(routeTests, []routeTestType{
// 		{
// 			name:                    "streams - get",
// 			method:                  http.MethodGet,
// 			routePath:               "/hri/tenants/testTenant/streams",
// 			expectedHandlerFilePath: streamsHandlerPath,
// 			expectedPathParameters: map[string]string{
// 				param.TenantId: "testTenant",
// 			},
// 		},
// 		{
// 			name:                    "streams - create",
// 			method:                  http.MethodPost,
// 			routePath:               "/hri/tenants/testTenant/streams/testStream",
// 			expectedHandlerFilePath: streamsHandlerPath,
// 			expectedPathParameters: map[string]string{
// 				param.TenantId: "testTenant",
// 				param.StreamId: "testStream",
// 			},
// 		},
// 		{
// 			name:                    "streams - delete",
// 			method:                  http.MethodDelete,
// 			routePath:               "/hri/tenants/testTenant/streams/testStream",
// 			expectedHandlerFilePath: streamsHandlerPath,
// 			expectedPathParameters: map[string]string{
// 				param.TenantId: "testTenant",
// 				param.StreamId: "testStream",
// 			},
// 		},
// 	}...)

// 	for _, tc := range routeTests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			context = e.NewContext(nil, nil)
// 			e.Router().Find(tc.method, tc.routePath, context)
// 			assertRouteHandlerIsValid(t, context, tc.expectedHandlerFilePath)
// 			for param, _ := range tc.expectedPathParameters {
// 				assert.Equal(t, tc.expectedPathParameters[param], context.Param(param))
// 			}
// 			assert.Equal(t, len(tc.expectedPathParameters), len(context.ParamNames()))
// 		})
// 	}

// 	// The number of routes (including ready/liveness endpoint) should match the number in the echo struct
// 	assert.Equal(t, len(routeTests)+1, len(e.Routes()))
// }

// func assertRouteHandlerIsValid(t *testing.T, context echo.Context, path string) {
// 	// If a route is not defined, Echo will automatically set a "Not found" or "Not Allowed" handler function in the
// 	// context. These handler functions will not panic when called with an empty context. This fact can be used to
// 	// test whether the route resolves to an HRI handler function. The path string is used to check if the expected
// 	// handler file is found in the stack trace after the panic.
// 	err := error(nil)
// 	defer func() {
// 		if r := recover(); r != nil {
// 			stack := string(debug.Stack())
// 			assert.True(t,
// 				strings.Contains(stack, path),
// 				"The stack trace for the route's handler did not reference the expected handler file: %s",
// 				path,
// 			)
// 			return
// 		}
// 		// If the handler call did not panic, then an error was returned. This should be a 404 or a 405
// 		assert.Fail(t, err.Error())
// 	}()
// 	err = context.Handler()(context)
// }
