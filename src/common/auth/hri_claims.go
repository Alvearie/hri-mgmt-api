/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package auth

import (
	"gopkg.in/square/go-jose.v2/jwt"
	"strings"
)

// ClaimsHolder is an interface to support testing
type ClaimsHolder interface {
	Claims(claims interface{}) error
}

/*
HriClaims is used to extract 'claims' from a JWT access token that are needed
by processing logic. It also abstracts differences between various authorization
services like IBM AppID and Azure AD.
*/
type HriClaims struct {
	// jwt.Audience can marshal string or []string json types
	Audience jwt.Audience `json:"aud"`
	Subject  string       `json:"sub"`

	// Some OAuth services use `scopes` (IBM AppID) and other use `roles` (Azure AD).
	// This extracts both if present and HasScope() searches both
	Scope string   `json:"scope"`
	Roles []string `json:"roles"`
}

// HasScope searches both Scope and Roles for the specified scope
func (c HriClaims) HasScope(scope string) bool {
	// split space-delimited scope string into an array
	scopes := strings.Fields(c.Scope)

	for _, val := range scopes {
		if val == scope {
			// token contains claim with this scope
			return true
		}
	}

	for _, val := range c.Roles {
		if val == scope {
			// token contains claim with this scope
			return true
		}
	}

	return false
}
