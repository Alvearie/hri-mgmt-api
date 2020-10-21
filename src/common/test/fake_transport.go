/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package test

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"testing"
)

const (
	// NOTE: the following regex pattern is used to verify that request bodies contain a valid start/end dates
	DatePattern string = "2([0-9][0-9][0-9])-([01][0-9])-([0-3][0-9])T([0-2][0-9]):([0-6][0-9]):([0-6][0-9])Z"
)

type ElasticCall struct {
	RequestQuery       string
	RequestBody        string
	ResponseStatusCode int
	ResponseBody       string
	ResponseErr        error
}

// A fake Transport used to specify the desired behavior of an Elastic Client in unit tests.
// (See create_test.go, get_test.go, get_test_by_id_test.go, and update_status_test.go for usage examples)
//
// New instances of FakeTransport should be created using the NewFakeTransport constructor, optionally in
// combination with the AddCall builder method. i.e.
//     fake := NewFakeTransport(t).AddCall(<path_to_elastic_endpoint>, ElasticCall{...})
//
// The ElasticCall struct defines the expected Body and Query params that should be present when the
// endpoint is called. If either is specified, they will compared to the actual Body and Query params and
// the test will fail if not equal. The Body can optionally be specified as a regex pattern to account for
// dynamic body contents such as date strings (see create_test.go for an example). Keep in mind that if
// the body contains and regex reserved characters, ( i.e. ^, $, ., |, ], [, *, +, ), (, ? ) then these
// chars must be escaped with a "\" (see get_test.go for an example).
type FakeTransport struct {
	t             *testing.T
	expectedCalls map[string][]ElasticCall
}

func NewFakeTransport(t *testing.T) *FakeTransport {
	return &FakeTransport{
		t,
		make(map[string][]ElasticCall),
	}
}

func (ft *FakeTransport) AddCall(requestPath string, call ElasticCall) *FakeTransport {
	// append works on nil slices so this call should always be safe
	ft.expectedCalls[requestPath] = append(ft.expectedCalls[requestPath], call)
	return ft
}

func (ft *FakeTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	// find the next stored call that corresponds to this Elastic endpoint path
	call := ft.lookupCall(request.URL.Path)

	// optionally validate that request query params are as expected
	if call.RequestQuery != "" && call.RequestQuery != request.URL.RawQuery {
		ft.t.Errorf("Incorrect request query params. Expected: %s, Actual: %s", call.RequestQuery, request.URL.RawQuery)
	}

	// optionally validate that the request body is as expected
	ft.checkBody(request.Body, call.RequestBody)

	// convert response status code and body to an http.Response
	httpResp := &http.Response{
		StatusCode: call.ResponseStatusCode,
		Body:       ioutil.NopCloser(strings.NewReader(call.ResponseBody)),
	}

	return httpResp, call.ResponseErr
}

func (ft *FakeTransport) VerifyCalls() {
	// check whether there were any expected calls that were not made
	for key, val := range ft.expectedCalls {
		if len(val) > 0 {
			ft.t.Errorf("Missing %d call(s) to Elastic endpoint: %s", len(val), key)
		}
	}
}

func (ft *FakeTransport) checkBody(body io.ReadCloser, expected string) {
	if expected != "" {
		// convert actual body to a string
		actual := ReaderToString(body)

		// expected body may contain regex patterns so use regex for equality check
		matched, err := regexp.MatchString(expected, actual)
		if err != nil {
			ft.t.Fatalf("Invalid regex string: %s", expected)
		} else if !matched {
			ft.t.Errorf("Incorrect request body.\nExpected: %s,\nActual  : %s", expected, actual)
		}
	}
}

func (ft *FakeTransport) lookupCall(requestPath string) ElasticCall {
	var call ElasticCall

	// find ElasticCalls that correspond to given Elastic endpoint
	calls, ok := ft.expectedCalls[requestPath]
	if !ok || len(calls) == 0 {
		ft.t.Errorf("Unexpected call to Elastic endpoint: %s", requestPath)
	} else {
		// pop the next ElasticCall for this endpoint by first retrieving the next entry
		// and then slicing that entry off of the list of remaining calls
		call = ft.expectedCalls[requestPath][0]
		ft.expectedCalls[requestPath] = ft.expectedCalls[requestPath][1:]
	}

	return call
}

func ReaderToString(reader io.ReadCloser) string {
	result := ""
	if reader != nil {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(reader)
		_ = reader.Close()
		result = buf.String()
	}
	return result
}
