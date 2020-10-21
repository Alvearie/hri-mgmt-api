/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package path

import (
	"errors"
	"fmt"
	"strings"
)

const ParamOwPath string = "__ow_path"

func ExtractParam(params map[string]interface{}, index int) (string, error) {
	path, ok := params[ParamOwPath].(string)
	if !ok {
		return "", errors.New(fmt.Sprintf("Required parameter '%s' is missing", ParamOwPath))
	}

	pathElements := strings.Split(path, "/")
	if len(pathElements) < index+1 {
		return "", errors.New(fmt.Sprintf("The path is shorter than the requested path parameter; path: %v, requested index: %d", pathElements, index))
	}
	return pathElements[index], nil
}
