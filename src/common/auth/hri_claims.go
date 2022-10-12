/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package auth

import (
	"strings"

	"gopkg.in/square/go-jose.v2/jwt"
)

// ClaimsHolder is an interface to support testing
type ClaimsHolder interface {
	Claims(claims interface{}) error
}

type HriClaims struct {
	// Claim information extracted from a JWT access token
	Scope    string   `json:"scope"`
	Subject  string   `json:"sub"`
	Audience []string `json:"aud"`
	// //Added as part of Azure porting
	// Roles []string `json:"roles"`

}

type HriAzClaims struct {
	Audience jwt.Audience `json:"aud"`
	Subject  string       `json:"sub"`

	// Some OAuth services use `scopes` (IBM AppID) and other use `roles` (Azure AD).
	// This extracts both if present and HasScope() searches both
	Scope string   `json:"scope"`
	Roles []string `json:"roles"`
}

func (c HriClaims) HasScope(claim string) bool {
	// split space-delimited scope string into an array
	scopes := strings.Fields(c.Scope)

	for _, val := range scopes {
		if val == claim {
			// token contains claim for this scope
			return true
		}
	}

	return false
}

func (c HriAzClaims) HasRole(claim string) bool {
	// split space-delimited scope string into an array
	scopes := strings.Fields(c.Scope)

	for _, val := range scopes {
		if val == claim {
			// token contains claim for this scope
			return true
		}
	}

	for _, val := range c.Roles {
		if val == claim {
			// token contains claim with this role
			return true
		}
	}

	return false
}
