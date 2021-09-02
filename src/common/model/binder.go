/**
 * (C) Copyright IBM Corp. 2021
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"regexp"
	"strings"
)

type CustomBinder struct {
	// DefaultBinder is the default implementation of the Binder interface.
	defaultBinder *echo.DefaultBinder
}

const unmarshalErrRegex = `code=400, message=Unmarshal type error: expected=.+, got=.+, field=[a-zA-Z]+, offset=\d+, internal=json: cannot unmarshal .+ into Go struct field .+ of type .+`

func GetBinder() (*CustomBinder, error) {
	// custom binder for binding request payload.
	db := &echo.DefaultBinder{}
	custom := CustomBinder{
		defaultBinder: db,
	}
	return &custom, nil
}

// Bind is a wrapper function for echo's defaultBinder's Bind function.  The only difference in functionality is that
// if defaultBinder.Bind throws an error, that error message is reconstructed into a more user-friendly error message,
// thus a new and improved error is returned instead.
func (cb CustomBinder) Bind(i interface{}, c echo.Context) error {
	// Note:  Query parameters are only bound for GET/DELETE methods
	binderErr := cb.defaultBinder.Bind(i, c)
	if binderErr == nil {
		return binderErr
	}

	fullErrMsg := binderErr.Error()

	if fullErrMsg == "code=400, message=unexpected EOF, internal=unexpected EOF" {
		return errors.New("unable to parse request body due to unexpected EOF")
	}

	var matchUnmarshalErr = regexp.MustCompile(unmarshalErrRegex)
	if matchUnmarshalErr.MatchString(fullErrMsg) {
		echoErr, ok := binderErr.(*echo.HTTPError)
		if ok {
			return getRevisedUnmarshalErr(*echoErr)
		}
	}

	return binderErr
}

// Example)  Original Err: "Unmarshal type error: expected=int, got=string, field=expectedRecordCount, offset=29, internal=json: cannot unmarshal string into Go struct field SendCompleteRequest.expectedRecordCount of type int"
//                New Err: "invalid request param \"expectedRecordCount\": expected type int, but received type string"
func getRevisedUnmarshalErr(err echo.HTTPError) error {
	internal := err.Internal.(*json.UnmarshalTypeError)
	newErrStr := fmt.Sprintf(
		`invalid request param "%s": expected type %s, but received type %s`,
		internal.Field,
		internal.Type.String(),
		internal.Value,
	)
	return errors.New(newErrStr)
}

func getRevisedUnmarshalOLD(fullErrMsg string) error {
	newErrStr := ""
	errMsgDetails := strings.Split(fullErrMsg, "Unmarshal type error: ")[1]
	errComponents := strings.Split(errMsgDetails, ", ")

	for _, str := range errComponents {
		pair := strings.Split(str, "=")
		key, val := pair[0], pair[1]
		switch key {
		case "expected":
			newErrStr = newErrStr + "expected type " + val + ", "
		case "got":
			newErrStr = newErrStr + "but received type " + val
		case "field":
			newErrStr = "invalid request param \"" + val + "\": " + newErrStr
		}
	}
	return errors.New(newErrStr)
}
