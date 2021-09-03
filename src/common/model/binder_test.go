/**
 * (C) Copyright IBM Corp. 2021
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package model

import (
	"encoding/json"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

type TestBindStruct struct {
	JsonParam  string `json:"jsonParam"`
	QueryParam string `query:"queryParam"`
	PathParam  string `param:"pathParam"`
}

func TestBind(t *testing.T) {
	b, err := GetBinder()
	assert.NoError(t, err)

	e := echo.New()
	e.Binder = b

	testCases := []struct {
		name                 string
		requestBody          string
		queryParams          [][]string
		pathParamVal         string
		expectedError        error
		expectedBoundRequest interface{}
	}{
		{
			name:        "valid input present",
			requestBody: `{"jsonParam":"is present"}`,
			queryParams: [][]string{
				{"queryParam", "zoop"},
			},
			pathParamVal: "bidoof",
			expectedBoundRequest: TestBindStruct{
				JsonParam:  "is present",
				QueryParam: "zoop",
				PathParam:  "bidoof",
			},
		},
		{
			name:          "bad json (unexpected EOF)",
			requestBody:   `{`,
			expectedError: errors.New("unable to parse request body due to unexpected EOF"),
		},
		{
			name:          "unmarshal to incorrect type",
			requestBody:   `{"jsonParam":6.02214}`,
			expectedError: errors.New(`invalid request param "jsonParam": expected type string, but received type number`),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, getRequestTarget(tc.queryParams), getRequestBody(tc.requestBody))
			request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			recorder := httptest.NewRecorder()
			context := e.NewContext(request, recorder)
			setUpPath(context, tc.pathParamVal)

			var requestStruct TestBindStruct
			err := b.Bind(&requestStruct, context)

			if tc.expectedError == nil {
				if !reflect.DeepEqual(tc.expectedBoundRequest, requestStruct) {
					t.Errorf("Bind\nexpected\n%v\nactual\n%v", tc.expectedBoundRequest, requestStruct)
				}
			} else {
				if !reflect.DeepEqual(tc.expectedError, err) {
					t.Errorf("Bind Error\nexpected\n%v\nactual\n%v", tc.expectedError, err)
				}
			}
		})
	}
}

func TestGetRevisedUnmarshalErr(t *testing.T) {
	testCases := []struct {
		name                 string
		received             string
		field                string
		inputErrMsg          string
		expectedOutputErrMsg string
	}{
		{
			name:                 "expected int, got string",
			received:             "string",
			field:                "expectedRecordCount",
			inputErrMsg:          "Unmarshal type error: expected=int, got=string, field=expectedRecordCount, offset=29, internal=json: cannot unmarshal string into Go struct field SendCompleteRequest.expectedRecordCount of type int",
			expectedOutputErrMsg: "invalid request param \"expectedRecordCount\": expected type int, but received type string",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonErr := json.UnmarshalTypeError{
				Value:  tc.received,
				Type:   reflect.TypeOf(134340),
				Offset: 0,
				Struct: "TestBindStruct",
				Field:  "expectedRecordCount",
			}
			err := echo.NewHTTPError(http.StatusBadRequest, tc.inputErrMsg).SetInternal(&jsonErr)
			assert.Equal(t, tc.expectedOutputErrMsg, getRevisedUnmarshalErr(*err).Error())
		})
	}
}

// Generate a request target based on the test case's query params.
func getRequestTarget(queryParams [][]string) string {
	if queryParams == nil {
		return "/"
	}
	q := make(url.Values)
	for _, paramPair := range queryParams {
		if len(paramPair[1]) > 0 {
			q.Set(paramPair[0], paramPair[1])
		}
	}
	return "/?" + q.Encode()
}

// Generate a request body and handle whether there is a body to pass in or not.
func getRequestBody(body string) io.Reader {
	if body == "" {
		return nil
	}
	return strings.NewReader(body)
}

// Set up the context path and handle whether there is a path param to set or not.
func setUpPath(context echo.Context, pathParamVal string) {
	if pathParamVal != "" {
		context.SetPath("/:pathParam")
		context.SetParamNames("pathParam")
		context.SetParamValues(pathParamVal)
	} else {
		context.SetPath("/")
	}
}
