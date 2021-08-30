/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/coreos/go-oidc"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"reflect"
	"testing"
)

// this is for manual testing with a specific OIDC provider (AppID)
func /*Test*/ OidcLib(t *testing.T) {
	const iss = "https://us-south.appid.cloud.ibm.com/oauth/v4/12345"
	const audienceId = "b8f85fbe-b00a-4296-b54b-e9ec09a5b2f3"
	username := os.Getenv("APPID_USERNAME")
	password := os.Getenv("APPID_PASSWORD")

	// First get a token from AppID
	request, err := http.NewRequest("POST", iss+"/token", bytes.NewBuffer([]byte("grant_type=client_credentials&audience="+audienceId)))
	if err != nil {
		t.Fatalf("Error creating new http request: %v", err)
	}
	request.SetBasicAuth(username, password)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		t.Fatalf("Error executing AppID token POST: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("Non 200 from AppID token POST: %v", resp)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Error decoding AppID token response: %v", err)
	}

	validator := theValidator{
		issuer:      iss,
		audienceId:  audienceId,
		providerNew: newProvider,
	}

	_, errResp := validator.getSignedToken(requestId, body["access_token"].(string))

	if errResp != nil {
		t.Fatalf("Error: %v", errResp)
	}
}

type fakeClaimsHolder struct {
	claims HriClaims
	err    error
}

func (f fakeClaimsHolder) Claims(claims interface{}) error {
	*claims.(*HriClaims) = f.claims
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

var hriClaims = HriClaims{Scope: "tenant_" + tenantId, Subject: "subject", Audience: []string{audienceId}}

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

	validator := theValidator{
		issuer:     issuer,
		audienceId: audienceId,
		providerNew: func(_ context.Context, s string) (oidcProvider, error) {
			assert.Equal(t, issuer, s)
			return mockProvider, nil
		},
	}

	actClaimsHolder, errResp := validator.getSignedToken(requestId, authorization)
	actClaims := HriClaims{}
	actClaimsHolder.Claims(&actClaims)

	assert.Nil(t, errResp)
	assert.Equal(t, hriClaims, actClaims)
}

func TestGetSignedTokenNewProviderError(t *testing.T) {
	validator := theValidator{
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

	validator := theValidator{
		issuer:     issuer,
		audienceId: audienceId,
		providerNew: func(_ context.Context, s string) (oidcProvider, error) {
			assert.Equal(t, issuer, s)
			return mockProvider, nil
		},
	}

	expErrResp := response.NewErrorDetailResponse(http.StatusUnauthorized, requestId, "Authorization token validation failed: oidc: malformed jwt: square/go-jose: compact JWS format must have three parts")

	_, errResp := validator.getSignedToken(requestId, authorization)

	assert.Equal(t, *expErrResp, *errResp)
}

func TestCheckTenant(t *testing.T) {
	authorizedTenant := "123"

	tests := []struct {
		name       string
		tenant     string
		claims     HriClaims
		statusCode int
	}{
		{
			name:   "Authorized",
			tenant: authorizedTenant,
			claims: HriClaims{Scope: "tenant_" + authorizedTenant},
		},
		{
			name:       "Unauthorized Tenant",
			tenant:     authorizedTenant,
			claims:     HriClaims{Scope: "unauthorizedTenant"},
			statusCode: http.StatusUnauthorized,
		},
	}

	validator := NewValidator("https://issuer", "audienceId").(theValidator)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := validator.checkTenant(requestId, tt.tenant, tt.claims)
			if resp != nil {
				assert.Equal(t, tt.statusCode, resp.Code)
			} else if resp == nil && tt.statusCode != 0 {
				// expected error response, but got none
				t.Fatalf("Expected err response with status code: %v, but got no error response.", tt.statusCode)
			}
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

	validator := theValidator{
		issuer:     issuer,
		audienceId: audienceId,
		providerNew: func(_ context.Context, s string) (oidcProvider, error) {
			assert.Equal(t, issuer, s)
			return mockProvider, nil
		},
	}

	claims, err := validator.GetValidatedClaims(requestId, authorization, tenantId)

	assert.Nil(t, err)
	assert.Equal(t, hriClaims, claims)
}

func TestGetValidatedClaimsTokenError(t *testing.T) {
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

	validator := theValidator{
		issuer:     issuer,
		audienceId: audienceId,
		providerNew: func(_ context.Context, s string) (oidcProvider, error) {
			assert.Equal(t, issuer, s)
			return mockProvider, nil
		},
	}

	claims, err := validator.GetValidatedClaims(requestId, authorization, tenantId)

	// we expect to get back an empty set of claims and a bad token error
	expClaims := HriClaims{}
	expErrResp := response.NewErrorDetailResponse(http.StatusUnauthorized, requestId, "Authorization token validation failed: oidc: malformed jwt: square/go-jose: compact JWS format must have three parts")

	if !reflect.DeepEqual(err, expErrResp) {
		t.Fatalf("Unexpected err response.\nexpected: %v\nactual  : %v", expErrResp, err)
	}
	assert.Equal(t, expClaims, claims)
}

func TestGetValidatedClaimsExtractionError(t *testing.T) {
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
			Return(fakeClaimsHolder{claims: HriClaims{}, err: errors.New(badClaimsHolderErr)}, nil),
	)

	validator := theValidator{
		issuer:     issuer,
		audienceId: audienceId,
		providerNew: func(_ context.Context, s string) (oidcProvider, error) {
			assert.Equal(t, issuer, s)
			return mockProvider, nil
		},
	}

	claims, err := validator.GetValidatedClaims(requestId, authorization, tenantId)

	// we expect to get back an empty set of claims and a bad claims error
	expClaims := HriClaims{}
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

	validator := theValidator{
		issuer:     issuer,
		audienceId: audienceId,
		providerNew: func(_ context.Context, s string) (oidcProvider, error) {
			assert.Equal(t, issuer, s)
			return mockProvider, nil
		},
	}

	claims, err := validator.GetValidatedClaims(requestId, authorization, "wrongTenantId")

	expErrResp := response.NewErrorDetailResponse(http.StatusUnauthorized, requestId, "Unauthorized tenant access. Tenant 'wrongTenantId' is not included in the authorized scopes: tenant_tenantId.")

	if !reflect.DeepEqual(err, expErrResp) {
		t.Fatalf("Unexpected err response.\nexpected: %v -> %v \nactual  : %v -> %v ", *expErrResp, *expErrResp.Body, *err, *err.Body)
	}
	assert.Equal(t, hriClaims, claims)
}
