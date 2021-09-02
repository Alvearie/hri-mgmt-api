/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package auth

import (
	"context"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/path"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/coreos/go-oidc"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
)

var Logger *log.Logger

// NOTE: Do to the tightly coupled nature of several classes in the oidc library, the test code coverage is below 90%.
// Interfaces cannot be created for the Provider and IDTokenVerifier, because return types cannot be inferred in Golang.

// type to support testing
type NewOidcProvider func(context.Context, string) (*oidc.Provider, error)

// interface to support testing
type Validator interface {
	GetSignedToken(params map[string]interface{}, providerNew NewOidcProvider, logger *log.Logger) (ClaimsHolder, map[string]interface{})
	CheckTenant(params map[string]interface{}, claims HriClaims, logger *log.Logger) map[string]interface{}
	CheckAudience(params map[string]interface{}, claims HriClaims, logger *log.Logger) map[string]interface{}
}

// struct that implements the Validator interface
type AuthValidator struct{}

// Ensures the request has a valid OAuth JWT OIDC compliant access token.
// Eventually this will also check the token scopes against a provided list.
// Returns a non-nil web action http response map on error.
func (v AuthValidator) GetSignedToken(params map[string]interface{}, providerNew NewOidcProvider, logger *log.Logger) (ClaimsHolder, map[string]interface{}) {
	issuer, rawToken, errResp := v.extractIssuerAndToken(params, logger)
	if errResp != nil {
		return nil, errResp
	}

	ctx := context.Background()

	provider, err := providerNew(ctx, issuer)
	if err != nil {
		msg := fmt.Sprintf("Failed to create OIDC provider: %s", err.Error())
		logger.Printf(msg)
		return nil, response.Error(http.StatusInternalServerError, msg)
	}

	// Since this is the protected resource, we don't have a client id and will accept tokens for all clients
	verifier := provider.Verifier(&oidc.Config{SkipClientIDCheck: true})

	token, err := verifier.Verify(ctx, rawToken)
	if err != nil {
		msg := fmt.Sprintf("Authorization token validation failed: %s", err.Error())
		logger.Printf(msg)
		return nil, response.Error(http.StatusUnauthorized, msg)
	}

	return token, nil
}

func (v AuthValidator) extractIssuerAndToken(params map[string]interface{}, logger *log.Logger) (string, string, map[string]interface{}) {

	// validate that required input params are present
	validator := param.ParamValidator{}
	errResp := validator.Validate(
		params,
		param.Info{param.OidcIssuer, reflect.String},
	)
	if errResp != nil {
		logger.Printf("Missing OIDC issuer parameter: %s", errResp)
		return "", "", errResp
	}

	// extract the Authorization Bearer token
	headers, err := param.ExtractValues(params, param.OpenWhiskHeaders)
	if err != nil {
		msg := fmt.Sprintf("Unable to extract OpenWhisk headers: %s", err.Error())
		logger.Printf(msg)
		return "", "", response.Error(http.StatusInternalServerError, msg)
	}

	auth, ok := headers["authorization"].(string)
	if !ok {
		auth, ok = headers["Authorization"].(string)
		if !ok {
			return "", "", response.Error(http.StatusUnauthorized, "Missing Authorization header")
		}
	}
	rawToken := strings.ReplaceAll(strings.ReplaceAll(auth, "Bearer ", ""), "bearer ", "")

	return params[param.OidcIssuer].(string), rawToken, nil
}

func (v AuthValidator) CheckTenant(params map[string]interface{}, claims HriClaims, logger *log.Logger) map[string]interface{} {
	// extract tenantId from URL path
	tenantId, err := path.ExtractParam(params, param.TenantIndex)
	if err != nil {
		logger.Println(err.Error())
		return response.Error(http.StatusBadRequest, err.Error())
	}

	// The tenant scope token must have "tenant_" as a prefix
	if !claims.HasScope(TenantScopePrefix + tenantId) {
		// The authorized scopes do not include tenant data
		msg := fmt.Sprintf("Unauthorized tenant access. Tenant '%s' is not included in the authorized scopes: %v.", tenantId, claims.Scope)
		logger.Println(msg)
		return response.Error(http.StatusUnauthorized, msg)
	}

	// Tenant data included in authorized scopes
	return nil
}

func (v AuthValidator) CheckAudience(params map[string]interface{}, claims HriClaims, logger *log.Logger) map[string]interface{} {
	// validate that required input params are present
	validator := param.ParamValidator{}
	errResp := validator.Validate(
		params,
		param.Info{param.JwtAudienceId, reflect.String},
	)
	if errResp != nil {
		logger.Printf("Missing %s parameter: %s", param.JwtAudienceId, errResp)
		return errResp
	}

	for _, jwtAudience := range claims.Audience {
		if jwtAudience == params[param.JwtAudienceId] {
			// JWT Audience matches JWT Audience Id
			return nil
		}
	}

	msg := fmt.Sprintf(
		"Unauthorized tenant access. The JWT Audience '%s' does not match the JWT Audience Id '%s'.",
		claims.Audience, params[param.JwtAudienceId])
	logger.Println(msg)
	return response.Error(http.StatusUnauthorized, msg)
}

func GetValidatedClaims(params map[string]interface{}, validator Validator, providerNew NewOidcProvider) (HriClaims, map[string]interface{}) {
	logger := GetLogger()
	claims := HriClaims{}

	// verify that request has a signed OAuth JWT OIDC-compliant access token
	token, errResp := validator.GetSignedToken(params, providerNew, logger)
	if errResp != nil {
		return claims, errResp
	}

	// extract HRI-related claims from JWT access token
	if err := token.Claims(&claims); err != nil {
		logger.Println(err.Error())
		return claims, response.Error(http.StatusUnauthorized, err.Error())
	}

	// verify that necessary tenant claim exists to access this endpoint's data
	if errResp := validator.CheckTenant(params, claims, logger); errResp != nil {
		return claims, errResp
	}

	// verify that the JWT Audience matches the JWT Audience Id
	if errResp := validator.CheckAudience(params, claims, logger); errResp != nil {
		return claims, errResp
	}

	return claims, nil
}

func GetLogger() *log.Logger {
	if Logger == nil {
		Logger = log.New(os.Stdout, "auth/validate: ", log.Llongfile)
	}

	return Logger
}
