/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package kafka

import (
	"github.com/Alvearie/hri-mgmt-api/common/config"
	"github.com/stretchr/testify/assert"
	"net/http"
	"regexp"
	"testing"
)

const (
	validToken   = "BEaRer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNjUyMTA4MTQ0LCJleHAiOjI1NTIxMTE3NDR9.XxTTNBtgjX48iCM4FaV_hhhGenzhzrUaTWn6ooepK14" // expires in 2050
	expiredToken = "bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNDUyMTA4MTQ0LCJleHAiOjE1NTIxMTE3NDR9.JCYxVQmkSoHtmcpl_AjIH_SD2fDDQvldwYyCU0xQcYw"
	invalidToken = "bearer INVALID"

	expiredTokenErrMsg = "Must supply an unexpired token:.*"
	invalidTokenErrMsg = "unexpected error parsing bearer token:.*"
)

func TestNewAdminClientFromConfig(t *testing.T) {

	validConfig := config.Config{
		KafkaBrokers: []string{"broker1", "broker2"},
	}

	testCases := []struct {
		name              string
		config            config.Config
		bearerToken       string
		success           bool
		expectedError     string
		expectedErrorCode int
	}{
		{
			name:        "happy path",
			config:      validConfig,
			bearerToken: validToken,
		}, {
			name:              "invalid bearer token",
			config:            validConfig,
			bearerToken:       invalidToken,
			expectedError:     invalidTokenErrMsg,
			expectedErrorCode: http.StatusUnauthorized,
		}, {
			name:              "expired bearer token",
			config:            validConfig,
			bearerToken:       expiredToken,
			expectedError:     expiredTokenErrMsg,
			expectedErrorCode: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			adminClient, errCode, err := NewAdminClientFromConfig(tc.config, tc.bearerToken)

			if len(tc.expectedError) == 0 {
				assert.NotNil(t, adminClient)
			} else {
				matched, _ := regexp.MatchString(tc.expectedError, err.Error())
				if !matched {
					t.Errorf("Returned error did not match expected.\nExpected: %s, Actual: %s", tc.expectedError, err.Error())
				}
				assert.Equal(t, tc.expectedErrorCode, errCode)
			}
		})
	}
}
