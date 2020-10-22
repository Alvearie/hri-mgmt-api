/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package path

import "testing"

func TestExtractParamSuccess(t *testing.T) {
	param, err := ExtractParam(map[string]interface{}{ParamOwPath: "/some/rest/type/path"}, 2)
	if err != nil {
		t.Error(err)
	}
	if param != "rest" {
		t.Errorf("Wrong path param; expected: rest, got: %s", param)
	}
}

func TestExtractParamMissingPath(t *testing.T) {
	_, err := ExtractParam(map[string]interface{}{"key": "/some/rest/type/path"}, 2)
	if err == nil || err.Error() != "Required parameter '__ow_path' is missing" {
		t.Errorf("Didn't get expected error when open whisk path parameter was missing; got: %v", err)
	}
}

func TestExtractParamOverIndex(t *testing.T) {
	_, err := ExtractParam(map[string]interface{}{ParamOwPath: "/some/rest/type/path"}, 5)
	if err == nil || err.Error() != "The path is shorter than the requested path parameter; path: [ some rest type path], requested index: 5" {
		t.Errorf("Didn't get expected error when over indexing the path; got: %v", err)
	}
}
