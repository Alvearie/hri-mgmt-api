/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package auth

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestHasScope(t *testing.T) {
	scopeToFind := "hri_data_integrator"

	tests := []struct {
		name     string
		scope    string
		expected bool
	}{
		{
			name:     "Empty String",
			expected: false,
		},
		{
			name:     "No Spaces String",
			scope:    "hri_other" + scopeToFind + "hri_consumer",
			expected: false,
		},
		{
			name:     "Comma-Delimited String",
			scope:    "hri_other, " + scopeToFind + ", hri_consumer",
			expected: false,
		},
		{
			name:     "Not In String",
			scope:    "hri_consumer",
			expected: false,
		},
		{
			name:     "Partially In String",
			scope:    "hri integrator hri_consumer",
			expected: false,
		},
		{
			name:     "Single scope",
			scope:    scopeToFind,
			expected: true,
		},
		{
			name:     "In String",
			scope:    scopeToFind + " hri_consumer",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := HriClaims{Scope: tt.scope}
			actual := claims.HasScope(scopeToFind)
			if actual != tt.expected {
				t.Fatalf("Unexpected result.\nexpected: %v\nactual  :\n%v", tt.expected, actual)
			}
		})
	}
}

func TestUnmarshalClaims(t *testing.T) {
	testScope := "testScope"
	testSubject := "testSubject"
	testAudience := "testAudience"
	jwtTokenJson := fmt.Sprintf(`
	{
		"scope": "%s",
		"sub": "%s",
		"aud": ["%s"]
	}`, testScope, testSubject, testAudience)

	claims := HriClaims{}
	err := json.Unmarshal([]byte(jwtTokenJson), &claims)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !reflect.DeepEqual(claims.Scope, testScope) {
		t.Fatalf("Unexpected result.\nexpected: %v\nactual  : %v", testScope, claims.Scope)
	}

	if !reflect.DeepEqual(claims.Subject, testSubject) {
		t.Fatalf("Unexpected result.\nexpected: %v\nactual  : %v", testSubject, claims.Subject)
	}

	if !reflect.DeepEqual(claims.Audience, []string{testAudience}) {
		t.Fatalf("Unexpected result.\nexpected: %v\nactual  : %v", []string{testAudience}, claims.Audience)
	}
}
