package param

import (
	"errors"
	"fmt"
)

const MissingSectionMsg string = "error extracting the %s section of the JSON"

// ExtractValues attempts to traverse down the json map using the fields passed in, checking at each step
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
