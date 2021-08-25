/**
 * (C) Copyright IBM Corp. 2021
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package model

// Custom Validator Tags
const (
	InjectionCheckValidatorTag string = "injection-check-validator"
	TenantIdValidatorTag       string = "tenantid-validator"
	StreamIdValidatorTag       string = "streamid-validator"
)

// Custom Validation RegEx strings
const (
	InjectionCheckContainsString string = `"=<>[]{}`
)
