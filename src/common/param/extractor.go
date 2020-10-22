/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package param

import (
	"errors"
	"fmt"
)

const (
	MissingSectionMsg    string = "error extracting the %s section of the JSON"
	MissingKafkaFieldMsg string = "error extracting %s from Kafka credentials"
	KafkaResourceId      string = "messagehub"
)

// attempts to traverse down the json map using the fields passed in, checking at each step
// assumes everything is a json object
func ExtractValues(jsonMap map[string]interface{}, fields ...string) (map[string]interface{}, error) {
	rtnPtr := &jsonMap
	for _, field := range fields {
		value, ok := (*rtnPtr)[field].(map[string]interface{})
		if !ok {
			return nil, errors.New(fmt.Sprintf(MissingSectionMsg, field))
		}
		rtnPtr = &value
	}
	return *rtnPtr, nil
}

func ExtractString(params map[string]interface{}, fieldName string) (string, error) {
	creds, err := ExtractValues(params, BoundCreds, KafkaResourceId)
	if err != nil {
		return "", err
	}

	field, ok := creds[fieldName].(string)
	if !ok {
		return "", errors.New(fmt.Sprintf(MissingKafkaFieldMsg, fieldName))
	} else {
		return field, nil
	}
}
