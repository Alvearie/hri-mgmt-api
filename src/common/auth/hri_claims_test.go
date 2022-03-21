/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package auth

import (
	"encoding/json"
	"fmt"
	"gopkg.in/square/go-jose.v2/jwt"
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

const (
	testAudience = "testAudience"
	testSubject  = "testSubject"
	testScope1   = "testScope1"
	testScope2   = "testScope2"
)

var azureToken = fmt.Sprintf(`
{
  "aud": "%s",
  "iss": "https://sts.windows.net/677b4c1e-702a-4acf-bb3f-15bf9c98e79c/",
  "iat": 1628196759,
  "nbf": 1628196759,
  "exp": 1628200659,
  "aio": "E2ZgYOi+3mpzznqdz3QDD6Nn1QtTAQ==",
  "appid": "a507f324-4084-4e91-90c9-75558b5272c0",
  "appidacr": "1",
  "idp": "https://sts.windows.net/677b4c1e-702a-4acf-bb3f-15bf9c98e79c/",
  "oid": "2c3ce727-9504-4782-a8e4-5a35306adfc7",
  "rh": "0.AVAAHkx7Zypwz0q7PxW_nJjnnCTzB6WEQJFOkMl1VYtScsBQAAA.",
  "sub": "%s",
  "roles": [
    "extra",
    "%s",
    "%s"
  ],
  "tid": "677b4c1e-702a-4acf-bb3f-15bf9c98e79c",
  "uti": "nBs2yJvFE0eOXVdkJGjdAg",
  "ver": "1.0"
}`, testAudience, testSubject, testScope1, testScope2)

var appidToken = fmt.Sprintf(`
{
  "iss": "https://us-south.appid.cloud.ibm.com/oauth/v4/2189304b-38dc-49b4-9734-b94b52ea6c28",
  "exp": 1572529268,
  "aud": [
    "%s"
  ],
  "sub": "%s",
  "email_verified": true,
  "amr": [
    "cloud_directory"
  ],
  "iat": 1572525668,
  "tenant": "2189304b-38dc-49b4-9734-b94b52ea6c28",
  "scope": "openid appid_default appid_readprofile appid_readuserattr appid_writeuserattr appid_authenticated %s %s"
}`, testAudience, testSubject, testScope1, testScope2)

var basicTokenWithScopes = fmt.Sprintf(`
{
  "aud": ["%s"],
  "sub": "%s",
  "scope": "%s %s"
}`, testAudience, testSubject, testScope1, testScope2)

var basicTokenWithRoles = fmt.Sprintf(`
{
  "aud": "%s",
  "sub": "%s",
  "roles": ["%s","%s"]
}`, testAudience, testSubject, testScope1, testScope2)

var basicTokenWithScopesAndRoles = fmt.Sprintf(`
{
  "aud": "%s",
  "sub": "%s",
  "scope": "%s",
  "roles": ["%s"]
}`, testAudience, testSubject, testScope1, testScope2)

func TestUnmarshalClaims(t *testing.T) {

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "Azure token",
			token: azureToken,
		},
		{
			name:  "AppId token",
			token: appidToken,
		},
		{
			name:  "Basic token with scopes",
			token: basicTokenWithScopes,
		},
		{
			name:  "Basic token with roles",
			token: basicTokenWithRoles,
		},
		{
			name:  "Basic token with scopes and roles",
			token: basicTokenWithScopesAndRoles,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := HriClaims{}
			err := json.Unmarshal([]byte(tt.token), &claims)
			if err != nil {
				t.Fatal(err.Error())
			}

			if !reflect.DeepEqual(claims.Audience, jwt.Audience{testAudience}) {
				t.Fatalf("Unexpected result.\nexpected: %v\nactual  : %v", []string{testAudience}, claims.Audience)
			}

			if !reflect.DeepEqual(claims.Subject, testSubject) {
				t.Fatalf("Unexpected result.\nexpected: %v\nactual  : %v", testSubject, claims.Subject)
			}

			for _, scope := range []string{testScope1, testScope2} {
				if !claims.HasScope(scope) {
					t.Fatalf("Expected claims to have scope '%s'. claims: %v", scope, claims)
				}
			}

			if claims.HasScope("admin") {
				t.Fatalf("Expected claims to NOT have scope 'admin'. claims: %v", claims)
			}
		})
	}
}
