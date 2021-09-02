/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package auth

const (
	HriIntegrator                string = "hri_data_integrator"
	HriConsumer                  string = "hri_consumer"
	TenantScopePrefix            string = "tenant_"
	MsgAccessTokenMissingScopes         = "The access token must have one of these scopes: hri_consumer, hri_data_integrator"
	MsgIntegratorSubClaimNoMatch        = "The token's sub claim (clientId): %s does not match the data integratorId: %s"
	MsgIntegratorRoleRequired           = "Must have hri_data_integrator role to %s a batch"
	MsgSubClaimRequiredInJwt            = "JWT access token 'sub' claim must be populated"
)
