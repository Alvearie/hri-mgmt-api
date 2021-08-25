/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package auth

import (
	"strings"
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
