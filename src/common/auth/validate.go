package auth

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/coreos/go-oidc"
)

// Validator Public interface
type BatchValidator interface {
	GetValidatedRoles(requestId string, authorization string, tenant string) (HriAzClaims, *response.ErrorDetailResponse)
}
type TenantValidator interface {
	GetValidatedClaimsForTenant(requestId string, authorization string) *response.ErrorDetailResponse
}

// Added  as part of Azure porting
type theTenantValidator struct {
	issuer      string
	audienceId  string
	providerNew newOidcProvider // this enables unit tests to use a mocked oidcProvider
}

type theBatchValidator struct {
	issuer      string
	audienceId  string
	providerNew newOidcProvider // this enables unit tests to use a mocked oidcProvider
}

// Interfaces cannot be directly created for the auth Provider and IDTokenVerifier, because return types are not
// inferred in Golang. These wrapper structs embed the originals to meet our interface definitions, which enables unit
// testing.

type newOidcProvider func(context.Context, string) (oidcProvider, error)

type oidcProvider interface {
	Verifier(config *oidc.Config) tokenVerifier
}

// Embeds an oidc.Provider and wraps the Verifier method with the correct return type to meet the oidcProvider interface
type theOidcProvider struct {
	*oidc.Provider
}

func (p theOidcProvider) Verifier(config *oidc.Config) tokenVerifier {
	return theTokenVerifier{p.Provider.Verifier(config)}
}

type tokenVerifier interface {
	Verify(ctx context.Context, rawIDToken string) (ClaimsHolder, error)
}

// Embeds an oidc.IDTokenVerifier and wraps the Verify method with the correct return type to meet the tokenVerifier interface
type theTokenVerifier struct {
	*oidc.IDTokenVerifier
}

func (t theTokenVerifier) Verify(ctx context.Context, rawIDToken string) (ClaimsHolder, error) {
	return t.IDTokenVerifier.Verify(ctx, rawIDToken)
}

// Method to create our custom oidcProvider, which embeds an oidc.Provider
func newProvider(ctx context.Context, issuer string) (oidcProvider, error) {
	oidcProvider, err := oidc.NewProvider(ctx, issuer)
	return theOidcProvider{oidcProvider}, err
}

// NewValidator Public default constructor
func NewTenantValidator(issuer string, audienceId string) TenantValidator {
	return theTenantValidator{
		issuer:      issuer,
		audienceId:  audienceId,
		providerNew: newProvider,
	}
}

func NewBatchValidator(issuer string, audienceId string) BatchValidator {
	return theBatchValidator{
		issuer:      issuer,
		audienceId:  audienceId,
		providerNew: newProvider,
	}
}

// Ensures the request has a valid OAuth JWT OIDC compliant access token.
func (v theTenantValidator) getSignedToken(requestId string, authorization string) (ClaimsHolder, *response.ErrorDetailResponse) {
	rawToken := strings.ReplaceAll(strings.ReplaceAll(authorization, "Bearer ", ""), "bearer ", "")

	ctx := context.Background()
	prefix := "auth/validate"
	logger := logwrapper.GetMyLogger(requestId, prefix)

	provider, err := v.providerNew(ctx, v.issuer)
	if err != nil {
		msg := fmt.Sprintf("Failed to create OIDC provider: %s", err.Error())
		logger.Errorln(msg)
		return nil, response.NewErrorDetailResponse(http.StatusInternalServerError, requestId, msg)
	}

	// This verifies the `aud` claim equals the configured audienceId
	verifier := provider.Verifier(&oidc.Config{ClientID: v.audienceId})

	token, err := verifier.Verify(ctx, rawToken)
	if err != nil {
		msg := fmt.Sprintf("Authorization token validation failed: %s", err.Error())
		logger.Errorln(msg)
		errmsg := "Azure AD authentication returned " + strconv.Itoa(http.StatusUnauthorized)

		if strings.Contains(msg, "JWS format must have three parts") || strings.Contains(msg, "malformed jwt") || strings.Contains(msg, "Token Expiry") {
			return nil, response.NewErrorDetailResponse(http.StatusUnauthorized, requestId, errmsg)
		}

		return nil, response.NewErrorDetailResponse(http.StatusUnauthorized, requestId, msg)

	}

	return token, nil
}
func (v theTenantValidator) GetValidatedClaimsForTenant(requestId string, authorization string) *response.ErrorDetailResponse {

	// verify that request has a signed OAuth JWT OIDC-compliant access token
	_, errResp := v.getSignedToken(requestId, authorization)
	if errResp != nil {
		return errResp
	}
	return nil
}

func (v theBatchValidator) GetValidatedRoles(requestId string, authorization string, tenant string) (HriAzClaims, *response.ErrorDetailResponse) {
	claims := HriAzClaims{}

	prefix := "auth/getValidatedRoles"
	logger := logwrapper.GetMyLogger(requestId, prefix)
	logger.Infoln(requestId, prefix)
	// verify that request has a signed OAuth JWT OIDC-compliant access token
	token, errResp := v.getSignedToken(requestId, authorization)
	if errResp != nil {
		return claims, errResp
	}

	// extract HRI-related claims from JWT access token
	if err := token.Claims(&claims); err != nil {
		logger.Errorln(err.Error())
		return claims, response.NewErrorDetailResponse(http.StatusUnauthorized, requestId, err.Error())
	}

	// verify that necessary tenant claim exists to access this endpoint's data
	if errResp := v.checkTenantScope(requestId, tenant, claims); errResp != nil {
		return claims, errResp
	}

	return claims, nil
}

func (v theBatchValidator) checkTenantScope(requestId string, tenantId string, claims HriAzClaims) *response.ErrorDetailResponse {
	prefix := "auth/checkTenant"
	logger := logwrapper.GetMyLogger(requestId, prefix)

	// The tenant scope token must have "tenant_" as a prefix
	if !claims.HasRole(TenantScopePrefix + tenantId) {
		// The authorized scopes do not include tenant data
		msg := fmt.Sprintf("Unauthorized tenant access. Tenant '%s' is not included in the authorized roles:tenant_%s.", tenantId, tenantId)
		logger.Errorln(msg)
		return response.NewErrorDetailResponse(http.StatusUnauthorized, requestId, msg)
	}

	// Tenant data included in authorized scopes
	return nil
}

func (v theBatchValidator) getSignedToken(requestId string, authorization string) (ClaimsHolder, *response.ErrorDetailResponse) {
	rawToken := strings.ReplaceAll(strings.ReplaceAll(authorization, "Bearer ", ""), "bearer ", "")

	ctx := context.Background()
	prefix := "auth/validate"
	logger := logwrapper.GetMyLogger(requestId, prefix)

	provider, err := v.providerNew(ctx, v.issuer)
	if err != nil {
		msg := fmt.Sprintf("Failed to create OIDC provider: %s", err.Error())
		logger.Errorln(msg)
		return nil, response.NewErrorDetailResponse(http.StatusInternalServerError, requestId, msg)
	}

	// This verifies the `aud` claim equals the configured audienceId
	verifier := provider.Verifier(&oidc.Config{ClientID: v.audienceId})

	token, err := verifier.Verify(ctx, rawToken)
	if err != nil {
		msg := fmt.Sprintf("Authorization token validation failed: %s", err.Error())
		logger.Errorln(msg)
		errmsg := "Azure AD authentication returned " + strconv.Itoa(http.StatusUnauthorized)

		if strings.Contains(msg, "JWS format must have three parts") || strings.Contains(msg, "malformed jwt") || strings.Contains(msg, "Token Expiry") {
			return nil, response.NewErrorDetailResponse(http.StatusUnauthorized, requestId, errmsg)
		}

		return nil, response.NewErrorDetailResponse(http.StatusUnauthorized, requestId, msg)
	}

	return token, nil
}
