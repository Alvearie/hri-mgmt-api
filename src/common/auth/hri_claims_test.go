/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package auth

import (
	"testing"
)

func TestHasRole(t *testing.T) {
	scopeToFind := "hri_data_integrator"

	tests := []struct {
		name     string
		scope    string
		roles    []string
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
		{
			name:     "In String",
			roles:    []string{scopeToFind},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := HriAzClaims{Scope: tt.scope, Roles: tt.roles}
			actual := claims.HasRole(scopeToFind)
			if actual != tt.expected {
				t.Fatalf("Unexpected result.\nexpected: %v\nactual  :\n%v", tt.expected, actual)
			}
		})
	}
}

func TestGetAuthRole(t *testing.T) {

	tests := []struct {
		name     string
		role     string
		expected string
	}{
		{
			name:     "Default",
			role:     "Test String",
			expected: "Test String",
		},
		{
			name:     "HriIntegrator",
			role:     HriIntegrator,
			expected: Hri + TenantScopePrefix + tenantId + DataIntegrator,
		},
		{
			name:     "HriConsumer",
			role:     HriConsumer,
			expected: Hri + TenantScopePrefix + tenantId + DataConsumer,
		},
		{
			name:     "HriInternal",
			role:     HriInternal,
			expected: Hri + TenantScopePrefix + tenantId + DataInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			actual := GetAuthRole(tenantId, tt.role)
			if actual != tt.expected {
				t.Fatalf("Unexpected result.\nexpected: %v\nactual  :\n%v", tt.expected, actual)
			}
		})
	}
}
