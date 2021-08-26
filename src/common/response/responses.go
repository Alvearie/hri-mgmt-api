/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package response

type ErrorDetail struct {
	ErrorEventId     string `json:"errorEventId"`
	ErrorDescription string `json:"errorDescription"`
}

func NewErrorDetail(requestId string, description string) *ErrorDetail {
	errorDetail := ErrorDetail{
		ErrorEventId:     requestId,
		ErrorDescription: description,
	}
	return &errorDetail
}

type ErrorDetailResponse struct {
	Code int
	Body *ErrorDetail
}

func NewErrorDetailResponse(code int, requestId string, description string) *ErrorDetailResponse {
	return &ErrorDetailResponse{
		Code: code,
		Body: &ErrorDetail{
			ErrorEventId:     requestId,
			ErrorDescription: description,
		},
	}
}
