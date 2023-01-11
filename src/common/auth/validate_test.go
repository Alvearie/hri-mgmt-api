package auth

import (
	context "context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"testing"

	response "github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/coreos/go-oidc"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type fakeClaimsHolder struct {
	claims HriAzClaims
	err    error
}

func (f fakeClaimsHolder) Claims(claims interface{}) error {
	*claims.(*HriAzClaims) = f.claims
	return f.err
}

const (
	issuer        = "https://issuer"
	requestId     = "requestId"
	token         = "ASBESESAFSEF"
	authorization = "Bearer " + token
	tenantId      = "tenantId"
	audienceId    = "audienceId"
)

var hriClaims = HriAzClaims{Scope: "tenant_" + tenantId, Subject: "subject", Audience: []string{audienceId}}

func TestGetSignedTokenHappyPath(t *testing.T) {
	// create the mocks
	controller := gomock.NewController(t)
	defer controller.Finish()
	mockProvider := NewMockoidcProvider(controller)
	mocktokenVerifier := NewMocktokenVerifier(controller)

	// define expected calls
	gomock.InOrder(
		mockProvider.
			EXPECT().
			Verifier(&oidc.Config{ClientID: audienceId}).
			Return(mocktokenVerifier),
		mocktokenVerifier.
			EXPECT().
			Verify(gomock.Any(), token).
			Return(fakeClaimsHolder{claims: hriClaims, err: nil}, nil),
	)

	validator := theTenantValidator{
		issuer:     issuer,
		audienceId: audienceId,
		providerNew: func(_ context.Context, s string) (oidcProvider, error) {
			assert.Equal(t, issuer, s)
			return mockProvider, nil
		},
	}

	actClaimsHolder, errResp := validator.getSignedToken(requestId, authorization)
	actClaims := HriAzClaims{}
	actClaimsHolder.Claims(&actClaims)

	assert.Nil(t, errResp)
	assert.Equal(t, hriClaims, actClaims)
}

func TestGetSignedTokenNewProviderError(t *testing.T) {
	validator := theTenantValidator{
		issuer:     issuer,
		audienceId: audienceId,
		providerNew: func(_ context.Context, s string) (oidcProvider, error) {
			return nil, errors.New("Bad issuer url")
		},
	}

	expErr := response.NewErrorDetailResponse(http.StatusInternalServerError, requestId, "Failed to create OIDC provider: Bad issuer url")

	_, err := validator.getSignedToken(requestId, authorization)

	if err == nil || !reflect.DeepEqual(*err, *expErr) {
		t.Fatalf("Unexpected error, expected:\n%v -> %v \nactual:\n%v -> %v", *expErr, *expErr.Body, *err, *err.Body)
	}
}

func TestGetSignedTokenBadToken(t *testing.T) {
	// create the mocks
	controller := gomock.NewController(t)
	defer controller.Finish()
	mockProvider := NewMockoidcProvider(controller)
	mocktokenVerifier := NewMocktokenVerifier(controller)

	// define expected calls
	gomock.InOrder(
		mockProvider.
			EXPECT().
			Verifier(&oidc.Config{ClientID: audienceId}).
			Return(mocktokenVerifier),
		mocktokenVerifier.
			EXPECT().
			Verify(gomock.Any(), token).
			Return(nil, errors.New("oidc: malformed jwt: square/go-jose: compact JWS format must have three parts")),
	)

	validator := theTenantValidator{
		issuer:     issuer,
		audienceId: audienceId,
		providerNew: func(_ context.Context, s string) (oidcProvider, error) {
			assert.Equal(t, issuer, s)
			return mockProvider, nil
		},
	}

	expErrResp := response.NewErrorDetailResponse(http.StatusUnauthorized, requestId,
		"Azure AD authentication returned "+strconv.Itoa(http.StatusUnauthorized))

	_, errResp := validator.getSignedToken(requestId, authorization)

	assert.Equal(t, expErrResp, errResp)
}

func TestGetSignedTokenBadTokenErr(t *testing.T) {
	// create the mocks
	controller := gomock.NewController(t)
	defer controller.Finish()
	mockProvider := NewMockoidcProvider(controller)
	mocktokenVerifier := NewMocktokenVerifier(controller)

	// define expected calls
	gomock.InOrder(
		mockProvider.
			EXPECT().
			Verifier(&oidc.Config{ClientID: audienceId}).
			Return(mocktokenVerifier),
		mocktokenVerifier.
			EXPECT().
			Verify(gomock.Any(), token).
			Return(nil, errors.New("Any other Error")),
	)

	validator := theTenantValidator{
		issuer:     issuer,
		audienceId: audienceId,
		providerNew: func(_ context.Context, s string) (oidcProvider, error) {
			assert.Equal(t, issuer, s)
			return mockProvider, nil
		},
	}

	expErrResp := response.NewErrorDetailResponse(http.StatusUnauthorized, requestId,
		fmt.Sprintf("Authorization token validation failed: %s", "Any other Error"))

	_, errResp := validator.getSignedToken(requestId, authorization)

	assert.Equal(t, expErrResp, errResp)
}

func TestCheckTenantScope(t *testing.T) {
	authorizedTenant := "123"

	tests := []struct {
		name       string
		tenant     string
		claims     HriAzClaims
		statusCode int
	}{
		{
			name:   "Authorized",
			tenant: authorizedTenant,
			claims: HriAzClaims{Scope: "tenant_" + authorizedTenant},
		},
	}

	validator := NewBatchValidator("https://issuer", "audienceId").(theBatchValidator)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator.checkTenantScope(requestId, tt.tenant, tt.claims)

		})
	}
}
func TestGetValidatedClaimsForTenantErr(t *testing.T) {
	authorizedTenant := "123"

	tests := []struct {
		name       string
		tenant     string
		claims     HriAzClaims
		statusCode int
	}{
		{
			name:   "Authorized",
			tenant: authorizedTenant,
			claims: HriAzClaims{Scope: "tenant_" + authorizedTenant},
		},
	}

	validator := NewTenantValidator("https://issuer", "audienceId").(theTenantValidator)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator.GetValidatedClaimsForTenant(requestId, tt.tenant)

		})
	}
}

func TestGetValidatedClaimsHappyPath(t *testing.T) {
	// create the mocks
	controller := gomock.NewController(t)
	defer controller.Finish()
	mockProvider := NewMockoidcProvider(controller)
	mocktokenVerifier := NewMocktokenVerifier(controller)

	// define expected calls
	gomock.InOrder(
		mockProvider.
			EXPECT().
			Verifier(&oidc.Config{ClientID: audienceId}).
			Return(mocktokenVerifier),
		mocktokenVerifier.
			EXPECT().
			Verify(gomock.Any(), token).
			Return(fakeClaimsHolder{claims: hriClaims, err: nil}, nil),
	)

	validator := theTenantValidator{
		issuer:     issuer,
		audienceId: audienceId,
		providerNew: func(_ context.Context, s string) (oidcProvider, error) {
			assert.Equal(t, issuer, s)
			return mockProvider, nil
		},
	}

	err := validator.GetValidatedClaimsForTenant(requestId, authorization)

	assert.Nil(t, err)

}

func TestGetValidatedRolesHappyPath(t *testing.T) {
	// create the mocks
	controller := gomock.NewController(t)
	defer controller.Finish()
	mockProvider := NewMockoidcProvider(controller)
	mocktokenVerifier := NewMocktokenVerifier(controller)

	// define expected calls
	gomock.InOrder(
		mockProvider.
			EXPECT().
			Verifier(&oidc.Config{ClientID: audienceId}).
			Return(mocktokenVerifier),
		mocktokenVerifier.
			EXPECT().
			Verify(gomock.Any(), token).
			Return(fakeClaimsHolder{claims: hriClaims, err: nil}, nil),
	)

	validator := theBatchValidator{
		issuer:     issuer,
		audienceId: audienceId,
		providerNew: func(_ context.Context, s string) (oidcProvider, error) {
			assert.Equal(t, issuer, s)
			return mockProvider, nil
		},
	}

	claims, err := validator.GetValidatedRoles(requestId, authorization, tenantId)

	assert.Nil(t, err)
	assert.Equal(t, hriClaims, claims)
}

func TestGetValidatedRolesTokenError(t *testing.T) {
	// create the mocks
	controller := gomock.NewController(t)
	defer controller.Finish()
	mockProvider := NewMockoidcProvider(controller)
	mocktokenVerifier := NewMocktokenVerifier(controller)

	// define expected calls
	gomock.InOrder(
		mockProvider.
			EXPECT().
			Verifier(&oidc.Config{ClientID: audienceId}).
			Return(mocktokenVerifier),
		mocktokenVerifier.
			EXPECT().
			Verify(gomock.Any(), token).
			Return(nil, errors.New("oidc: malformed jwt: square/go-jose: compact JWS format must have three parts")),
	)

	validator := theBatchValidator{
		issuer:     issuer,
		audienceId: audienceId,
		providerNew: func(_ context.Context, s string) (oidcProvider, error) {
			assert.Equal(t, issuer, s)
			return mockProvider, nil
		},
	}

	claims, err := validator.GetValidatedRoles(requestId, authorization, tenantId)

	// we expect to get back an empty set of claims and a bad token error
	expClaims := HriAzClaims{}
	expErrResp := response.NewErrorDetailResponse(http.StatusUnauthorized, requestId, "Azure AD authentication returned 401")

	if !reflect.DeepEqual(err, expErrResp) {
		t.Fatalf("Unexpected err response.\nexpected: %v\nactual  : %v", expErrResp, err)
	}
	assert.Equal(t, expClaims, claims)
}

func TestGetValidatedRolesExtractionError(t *testing.T) {
	// create the mocks
	controller := gomock.NewController(t)
	defer controller.Finish()
	mockProvider := NewMockoidcProvider(controller)
	mocktokenVerifier := NewMocktokenVerifier(controller)

	badClaimsHolderErr := "bad claims holder"

	// define expected calls
	gomock.InOrder(
		mockProvider.
			EXPECT().
			Verifier(&oidc.Config{ClientID: audienceId}).
			Return(mocktokenVerifier),
		mocktokenVerifier.
			EXPECT().
			Verify(gomock.Any(), token).
			Return(fakeClaimsHolder{claims: HriAzClaims{}, err: errors.New(badClaimsHolderErr)}, nil),
	)

	validator := theBatchValidator{
		issuer:     issuer,
		audienceId: audienceId,
		providerNew: func(_ context.Context, s string) (oidcProvider, error) {
			assert.Equal(t, issuer, s)
			return mockProvider, nil
		},
	}

	claims, err := validator.GetValidatedRoles(requestId, authorization, tenantId)

	// we expect to get back an empty set of claims and a bad claims error
	expClaims := HriAzClaims{}
	expErrResp := response.NewErrorDetailResponse(http.StatusUnauthorized, requestId, "bad claims holder")

	if !reflect.DeepEqual(err, expErrResp) {
		t.Fatalf("Unexpected err response.\nexpected: %v -> %v \nactual  : %v -> %v ", *expErrResp, *expErrResp.Body, *err, *err.Body)
	}
	assert.Equal(t, expClaims, claims)
}

func TestGetValidatedClaimsTenantError(t *testing.T) {
	// create the mocks
	controller := gomock.NewController(t)
	defer controller.Finish()
	mockProvider := NewMockoidcProvider(controller)
	mocktokenVerifier := NewMocktokenVerifier(controller)

	// define expected calls
	gomock.InOrder(
		mockProvider.
			EXPECT().
			Verifier(&oidc.Config{ClientID: audienceId}).
			Return(mocktokenVerifier),
		mocktokenVerifier.
			EXPECT().
			Verify(gomock.Any(), token).
			Return(fakeClaimsHolder{claims: hriClaims, err: nil}, nil),
	)

	validator := theBatchValidator{
		issuer:     issuer,
		audienceId: audienceId,
		providerNew: func(_ context.Context, s string) (oidcProvider, error) {
			assert.Equal(t, issuer, s)
			return mockProvider, nil
		},
	}

	claims, err := validator.GetValidatedRoles(requestId, authorization, "wrongTenantId")

	expErrResp := response.NewErrorDetailResponse(http.StatusUnauthorized, requestId, "Unauthorized tenant access. Tenant 'wrongTenantId' is not included in the authorized roles:tenant_wrongTenantId.")

	if !reflect.DeepEqual(err, expErrResp) {
		t.Fatalf("Unexpected err response.\nexpected: %v -> %v \nactual  : %v -> %v ", *expErrResp, *expErrResp.Body, *err, *err.Body)
	}
	assert.Equal(t, hriClaims, claims)
}

func TestGetSignedTokenBatchNewProviderError(t *testing.T) {
	validator := theBatchValidator{
		issuer:     issuer,
		audienceId: audienceId,
		providerNew: func(_ context.Context, s string) (oidcProvider, error) {
			return nil, errors.New("Bad issuer url")
		},
	}

	expErr := response.NewErrorDetailResponse(http.StatusInternalServerError, requestId, "Failed to create OIDC provider: Bad issuer url")

	_, err := validator.getSignedToken(requestId, authorization)

	if err == nil || !reflect.DeepEqual(*err, *expErr) {
		t.Fatalf("Unexpected error, expected:\n%v -> %v \nactual:\n%v -> %v", *expErr, *expErr.Body, *err, *err.Body)
	}
}

func TestGetSignedTokenBatchBadTokenErr(t *testing.T) {
	// create the mocks
	controller := gomock.NewController(t)
	defer controller.Finish()
	mockProvider := NewMockoidcProvider(controller)
	mocktokenVerifier := NewMocktokenVerifier(controller)

	// define expected calls
	gomock.InOrder(
		mockProvider.
			EXPECT().
			Verifier(&oidc.Config{ClientID: audienceId}).
			Return(mocktokenVerifier),
		mocktokenVerifier.
			EXPECT().
			Verify(gomock.Any(), token).
			Return(nil, errors.New("Any other Error")),
	)

	validator := theBatchValidator{
		issuer:     issuer,
		audienceId: audienceId,
		providerNew: func(_ context.Context, s string) (oidcProvider, error) {
			assert.Equal(t, issuer, s)
			return mockProvider, nil
		},
	}

	expErrResp := response.NewErrorDetailResponse(http.StatusUnauthorized, requestId,
		fmt.Sprintf("Authorization token validation failed: %s", "Any other Error"))

	_, errResp := validator.getSignedToken(requestId, authorization)

	assert.Equal(t, expErrResp, errResp)
}
