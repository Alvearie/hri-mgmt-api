/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package kafka

import (
	"regexp"
	"testing"

	"github.com/Alvearie/hri-mgmt-api/common/config"
	"github.com/stretchr/testify/assert"
)

const (
	validToken = "BEaRer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNjUyMTA4MTQ0LCJleHAiOjI1NTIxMTE3NDR9.XxTTNBtgjX48iCM4FaV_hhhGenzhzrUaTWn6ooepK14" // expires in 2050

)

func TestNewAdminClientFromConfig(t *testing.T) {

	validConfig := config.Config{
		AzKafkaBrokers:    []string{"broker1", "broker2"},
		AzKafkaProperties: map[string]string{"security.protocol": "sasl_ssl"},
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
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			adminClient, err := AzNewAdminClientFromConfig(tc.config, tc.bearerToken)

			if len(tc.expectedError) == 0 {
				assert.NotNil(t, adminClient)
			} else {
				matched, _ := regexp.MatchString(tc.expectedError, err.Body.ErrorDescription)
				if !matched {
					t.Errorf("Returned error did not match expected.\nExpected: %s, Actual: %s", tc.expectedError, err.Body.ErrorDescription)
				}
				assert.Equal(t, tc.expectedErrorCode, err.Code)
			}
		})
	}
}
