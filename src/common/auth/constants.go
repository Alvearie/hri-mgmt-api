/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package auth

const (
	HriIntegrator                string = "hri_data_integrator"
	HriInternal                  string = "hri_internal"
	HriConsumer                  string = "hri_consumer"
	NoAuthFakeIntegrator         string = "NoAuthUnkIntegrator"
	TenantScopePrefix            string = "tenant_"
	MsgAccessTokenMissingScopes         = "The access token must have one of these scopes: hri_consumer, hri_data_integrator"
	MsgIntegratorSubClaimNoMatch        = "The token's sub claim (clientId): %s does not match the data integratorId: %s"
	MsgIntegratorRoleRequired           = "Must have hri_data_integrator role to %s a batch"
	MsgInternalRoleRequired             = "Must have hri_internal role to mark a batch as %s"
	MsgSubClaimRequiredInJwt            = "JWT access token 'sub' claim must be populated."
	Hri                          string = "hri_"
	DataIntegrator               string = "_data_integrator"
	DataConsumer                 string = "_data_consumer"
	DataInternal                 string = "_data_internal"
)
