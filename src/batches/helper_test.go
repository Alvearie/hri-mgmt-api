/*
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package batches

import (
	"fmt"
	"reflect"
	"regexp"
)

func RequestCompareScriptTest(expectedUpdateRequest map[string]interface{}, actualUpdateRequest map[string]interface{}) error {

	matches, err := regexp.MatchString(expectedUpdateRequest["script"].(map[string]interface{})["source"].(string), actualUpdateRequest["script"].(map[string]interface{})["source"].(string))
	if err != nil || !matches {
		return err
	}

	return nil //update requests are equivalent

}

func RequestCompareWithMetadataTest(expectedUpdateRequest map[string]interface{}, actualUpdateRequest map[string]interface{}) error {

	if !reflect.DeepEqual(expectedUpdateRequest["script"].(map[string]interface{})["lang"].(string), actualUpdateRequest["script"].(map[string]interface{})["lang"].(string)) {
		return fmt.Errorf("expected lang field of script does not match actual lang field of script")
	}

	if !reflect.DeepEqual(expectedUpdateRequest["script"].(map[string]interface{})["params"].(map[string]interface{}), actualUpdateRequest["script"].(map[string]interface{})["params"].(map[string]interface{})) {
		return fmt.Errorf("expected params field of script does not match actual params field of script")
	}

	return nil //update requests are equivalent

}
