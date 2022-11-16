/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package elastic

/*
const msgClientErr string = "elasticsearch client error: %w"
const MsgNilResponse string = "elasticsearch client returned nil response without an error"
const msgParseErr string = "error parsing the Elasticsearch response body: %w"
const msgUnexpectedErr string = "unexpected Elasticsearch %d error"
const msgEmptyResultErr string = "unexpected empty result"

func DecodeBody(res *esapi.Response, elasticClientError error) (map[string]interface{}, *ResponseError) {
	err := checkForClientErr(res, elasticClientError)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	var body map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		err = fmt.Errorf(msgParseErr, err)
		return nil, &ResponseError{ErrorObj: err, Code: http.StatusInternalServerError}
	}

	if res.IsError() {
		if errBody, ok := body["error"].(map[string]interface{}); ok {
			return nil, getErrorFromElasticErrorResponse(errBody, res.StatusCode)
		} else if errMsg, ok := body["error"].(string); ok {
			return nil, &ResponseError{ErrorObj: fmt.Errorf(errMsg), Code: res.StatusCode}
		}

		// Elastic returned an error code, but no error information was returned in the response. This is normal Elastic
		//   behavior for some endpoints, most notably when a document isn't found using the Get Document API.
		return body, &ResponseError{ErrorObj: nil, Code: res.StatusCode}
	}

	return body, nil
}

func getErrorFromElasticErrorResponse(elasticError map[string]interface{}, statusCode int) *ResponseError {
	errorType := elasticError["type"].(string)
	err := fmt.Errorf("%s: %s", errorType, elasticError["reason"].(string))

	if rootCauses, ok := elasticError["root_cause"].([]interface{}); ok && len(rootCauses) > 0 {
		// The elastic error returned a root_cause section with more detailed error information.

		// Elastic's root_cause field contains a list of root causes. However, only one root cause is usually present,
		//   so only the first root cause will be returned in the response.
		rootCause := rootCauses[0].(map[string]interface{})
		rootCauseType := rootCause["type"].(string)

		if rootCauseType != errorType {
			// Elastic will often duplicate the error "type" and "reason" in the root_cause section, so the additional error
			//   information is not added if both error and root_error types are identical.
			rootCauseErr := fmt.Errorf("%s: %s", rootCauseType, rootCause["reason"])
			return &ResponseError{ErrorObj: fmt.Errorf("%v: %w", err, rootCauseErr), Code: statusCode,
				ErrorType: errorType, RootCause: rootCauseType}
		}
	}

	return &ResponseError{ErrorObj: err, Code: statusCode, ErrorType: errorType}
}

func DecodeBodyFromJsonArray(res *esapi.Response, err error) ([]map[string]interface{}, *ResponseError) {
	clientErr := checkForClientErr(res, err)
	if clientErr != nil {
		return nil, clientErr
	}

	defer res.Body.Close()

	if res.IsError() {
		// If an error occurred, then the response is (probably) an elastic error
		_, elasticErr := DecodeBody(res, err)
		if elasticErr != nil && len(elasticErr.ErrorType) != 0 {
			// DecodeBody was able to extract error information
			return nil, elasticErr
		}

		// DecodeBody wasn't able to extract error information, or the response was in some unanticipated format.
		return nil, &ResponseError{ErrorObj: fmt.Errorf(msgUnexpectedErr, res.StatusCode), Code: http.StatusInternalServerError}
	}

	var body []map[string]interface{}
	if parseErr := json.NewDecoder(res.Body).Decode(&body); parseErr != nil {
		return nil, &ResponseError{ErrorObj: fmt.Errorf(msgParseErr, parseErr), Code: http.StatusInternalServerError}
	}

	return body, nil
}

func DecodeFirstArrayElement(res *esapi.Response, err error) (map[string]interface{}, *ResponseError) {
	decoded, errResp := DecodeBodyFromJsonArray(res, err)
	if errResp != nil {
		return nil, errResp
	}

	if len(decoded) == 0 {
		return nil, &ResponseError{ErrorObj: fmt.Errorf(msgEmptyResultErr), Code: http.StatusInternalServerError}
	}

	return decoded[0], nil
}

func checkForClientErr(res *esapi.Response, err error) *ResponseError {
	if err != nil {
		return &ResponseError{ErrorObj: fmt.Errorf(msgClientErr, err), Code: http.StatusInternalServerError}
	} else if res == nil {
		return &ResponseError{ErrorObj: fmt.Errorf(MsgNilResponse), Code: http.StatusInternalServerError}
	}
	return nil
}
*/
