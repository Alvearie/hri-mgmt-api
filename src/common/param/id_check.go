/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package param

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

//Check that the tenantId contains only lower case characters, numbers, '-', or '_'
func TenantIdCheck(tenantId string) error {
	for _, r := range tenantId {
		if !(unicode.IsLetter(r) && unicode.IsLower(r)) && !unicode.IsNumber(r) && r != '-' && r != '_' {
			return errors.New("TenantId: " + tenantId + " must be lower-case alpha-numeric, '-', or '_'. '" + string(r) + "' is not allowed.")
		}
	}
	return nil
}

//Check that the streamId contains only lower case characters, numbers, '-', '_', and no more than one '.'
func StreamIdCheck(streamId string) error {
	for _, r := range streamId {
		if !(unicode.IsLetter(r) && unicode.IsLower(r)) && !unicode.IsNumber(r) && r != '-' && r != '_' && r != '.' {
			return errors.New("StreamId: " + streamId + " must be lower-case alpha-numeric, '-', or '_', and no more than one '.'. '" + string(r) + "' is not allowed.")
		}
	}
	if strings.Count(streamId, ".") > 1 {
		return errors.New("StreamId: " + streamId + " must be lower-case alpha-numeric, '-', or '_', and no more than one '.'. " + strconv.Itoa(strings.Count(streamId, ".")) + " .'s are not allowed.")
	}
	return nil
}
