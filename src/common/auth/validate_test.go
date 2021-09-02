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
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/path"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/coreos/go-oidc"
	"github.com/golang/mock/gomock"
	"log"
	"net/http"
	"os"
	"reflect"
	"testing"
)

// this is for manual testing with a specific OIDC provider (AppID)
func /*Test*/ OidcLib(t *testing.T) {
	const iss = "https://us-south.appid.cloud.ibm.com/oauth/v4/<appId_tenantId>"
	username := os.Getenv("APPID_USERNAME")
	password := os.Getenv("APPID_PASSWORD")

	// First get a token from AppID
	request, err := http.NewRequest("POST", iss+"/token", bytes.NewBuffer([]byte("grant_type=client_credentials")))
	if err != nil {
		t.Errorf("Error creating new http request: %v", err)
	}
	request.SetBasicAuth(username, password)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		t.Errorf("Error executing AppID token POST: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Non 200 from AppID token POST: %v", resp)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Errorf("Error decoding AppID token response: %v", err)
	}

	logger := log.New(os.Stdout, "auth/validate: ", log.Llongfile)
	validator := AuthValidator{}

	// now validate the token
	_, errResp := validator.GetSignedToken(
		map[string]interface{}{
			"issuer":               iss,
			param.OpenWhiskHeaders: map[string]interface{}{"authorization": "Bearer " + body["access_token"].(string)}},
		oidc.NewProvider,
		logger)

	if errResp != nil {
		t.Fatalf("Error: %v", errResp)
	}
}

func testProvider(ctx context.Context, issuer string) (*oidc.Provider, error) {
	return &oidc.Provider{}, nil
}

func TestGetSignedToken(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]interface{}
		expErr      map[string]interface{}
		newProvider NewOidcProvider
	}{
		{"Bad Token",
			map[string]interface{}{
				"issuer":               "https://issuer",
				param.OpenWhiskHeaders: map[string]interface{}{"authorization": "Bearer AEFFASEFLIJPIOJ"},
			},
			response.Error(http.StatusUnauthorized, "Authorization token validation failed: oidc: malformed jwt: square/go-jose: compact JWS format must have three parts"),
			testProvider,
		},
		{"Missing Params",
			map[string]interface{}{
				"issuer":               "https://issuer",
				param.OpenWhiskHeaders: map[string]interface{}{},
			},
			response.Error(http.StatusUnauthorized, "Missing Authorization header"),
			testProvider,
		},
		{"NewProvider Error",
			map[string]interface{}{
				"issuer":               "https://issuer",
				param.OpenWhiskHeaders: map[string]interface{}{"authorization": "Bearer AEFFASEFLIJPIOJ"},
			},
			response.Error(http.StatusInternalServerError, "Failed to create OIDC provider: new provider error"),
			func(ctx context.Context, issuer string) (*oidc.Provider, error) {
				return nil, errors.New("new provider error")
			},
		},
	}

	logger := log.New(os.Stdout, "auth/validate: ", log.Llongfile)
	validator := AuthValidator{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validator.GetSignedToken(tt.params, tt.newProvider, logger)
			expected := fmt.Sprint(tt.expErr)
			if err == nil || expected != fmt.Sprint(err) {
				t.Fatalf("Expected error with bad token, but expected:\n%v\ngot:\n%v", expected, err)
			}
		})
	}
}

func TestExtractIssuerAndToken(t *testing.T) {
	const issuer = "https://myissuer"
	const token = "LIJIJUGBFFDRYCXWRETYUJBCXCVBNKUYTRF"

	tests := []struct {
		name   string
		params map[string]interface{}
		expErr map[string]interface{}
	}{
		{"Happy Path",
			map[string]interface{}{param.OidcIssuer: issuer, param.OpenWhiskHeaders: map[string]interface{}{"authorization": "Bearer " + token}},
			nil,
		},
		{"Missing Issuer",
			map[string]interface{}{param.OpenWhiskHeaders: map[string]interface{}{"authorization": "Bearer " + token}},
			response.MissingParams(param.OidcIssuer),
		},
		{"Missing OpenWhisk Headers",
			map[string]interface{}{param.OidcIssuer: issuer},
			response.Error(http.StatusInternalServerError, "Unable to extract OpenWhisk headers: error extracting the __ow_headers section of the JSON"),
		},
		{"Missing Authorization",
			map[string]interface{}{param.OidcIssuer: issuer, param.OpenWhiskHeaders: map[string]interface{}{}},
			response.Error(http.StatusUnauthorized, "Missing Authorization header"),
		},
	}

	logger := log.New(os.Stdout, "auth/validate: ", log.Llongfile)
	validator := AuthValidator{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actIssuer, actToken, err := validator.extractIssuerAndToken(tt.params, logger)
			if (err != nil || tt.expErr != nil) && fmt.Sprint(tt.expErr) != fmt.Sprint(err) {
				t.Fatalf("Expected error does not match actual\nexpected:\n%v\nactual:\n%v", tt.expErr, err)
			} else if tt.expErr == nil && (issuer != actIssuer || token != actToken) {
				t.Fatalf("Expected does not match actual\nexpected:\n%s, %s\nactual:\n%s, %s", issuer, token, actIssuer, actToken)
			}
		})
	}
}

func TestCheckTenant(t *testing.T) {
	authorizedTenant := "123"

	tests := []struct {
		name       string
		params     map[string]interface{}
		claims     HriClaims
		statusCode int
	}{
		{
			name:       "Invalid Url Path",
			params:     map[string]interface{}{path.ParamOwPath: "/hri/bad_path"},
			claims:     HriClaims{Scope: "tenant_" + authorizedTenant},
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "Unauthorized Tenant",
			params:     map[string]interface{}{path.ParamOwPath: fmt.Sprintf("/hri/tenants/%s/batches", authorizedTenant)},
			claims:     HriClaims{Scope: "unauthorizedTenant"},
			statusCode: http.StatusUnauthorized,
		},
		{
			name:   "Happy Path",
			params: map[string]interface{}{path.ParamOwPath: fmt.Sprintf("/hri/tenants/%s/batches", authorizedTenant)},
			claims: HriClaims{Scope: "tenant_" + authorizedTenant},
		},
	}

	logger := log.New(os.Stdout, "auth/validate: ", log.Llongfile)
	validator := AuthValidator{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := validator.CheckTenant(tt.params, tt.claims, logger)
			if resp != nil {
				actualStatusCode := resp["error"].(map[string]interface{})["statusCode"]
				if actualStatusCode != tt.statusCode {
					// error response doesn't match expected response
					t.Fatalf("Unexpected err response.\nexpected: %v\nactual  : %v", tt.statusCode, actualStatusCode)
				}
			} else if resp == nil && tt.statusCode != 0 {
				// expected error response, but got none
				t.Fatalf("Expected err response with status code: %v, but got no error response.", tt.statusCode)
			}
		})
	}
}

func TestCheckAudience(t *testing.T) {
	jwtAudienceId := "123"

	tests := []struct {
		name       string
		params     map[string]interface{}
		claims     HriClaims
		statusCode int
	}{
		{
			name:       "No Audience Param",
			params:     map[string]interface{}{},
			claims:     HriClaims{Audience: []string{jwtAudienceId}},
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "Unauthorized Audience",
			params:     map[string]interface{}{param.JwtAudienceId: "Will not match Audience"},
			claims:     HriClaims{Audience: []string{jwtAudienceId}},
			statusCode: http.StatusUnauthorized,
		},
		{
			name:   "Happy Path",
			params: map[string]interface{}{param.JwtAudienceId: jwtAudienceId},
			claims: HriClaims{Audience: []string{jwtAudienceId}},
		},
	}

	logger := log.New(os.Stdout, "auth/validate: ", log.Llongfile)
	validator := AuthValidator{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := validator.CheckAudience(tt.params, tt.claims, logger)
			if resp != nil {
				actualStatusCode := resp["error"].(map[string]interface{})["statusCode"]
				if actualStatusCode != tt.statusCode {
					// error response doesn't match expected response
					t.Fatalf("Unexpected err response.\nexpected: %v\nactual  : %v", tt.statusCode, actualStatusCode)
				}
			} else if resp == nil && tt.statusCode != 0 {
				// expected error response, but got none
				t.Fatalf("Expected err response with status code: %v, but got no error response.", tt.statusCode)
			}
		})
	}
}

// Define a function type with the same signature as the single method in the ClaimsHolder interface
// Any function with this signature can be cast as ClaimsHolderFunc
type ClaimsHolderFunc func(claims interface{}) error

// Make ClaimsHolderFunc satisfy the ClaimsHolder interface
// When ClaimsHolderFunc.Claims is called, the args are simply forwarded on to the underlying ClaimsHolderFunc function
func (f ClaimsHolderFunc) Claims(claims interface{}) error {
	return f(claims)
}

func TestGetValidatedClaimsHappyPath(t *testing.T) {
	// create a mock validator
	controller := gomock.NewController(t)
	defer controller.Finish()
	mockValidator := NewMockValidator(controller)

	// define inputs/outputs
	applicationId := "applicationId"
	params := map[string]interface{}{param.JwtAudienceId: applicationId}
	hriClaims := HriClaims{Scope: "foo bar", Audience: []string{applicationId}}
	claimsHolder := func(c interface{}) error {
		*c.(*HriClaims) = hriClaims
		return nil
	}

	// define expected calls
	gomock.InOrder(
		mockValidator.
			EXPECT().
			GetSignedToken(params, nil, gomock.Any()).
			Return(ClaimsHolderFunc(claimsHolder), nil),
		mockValidator.
			EXPECT().
			CheckTenant(params, hriClaims, gomock.Any()).
			Return(nil),
		mockValidator.
			EXPECT().
			CheckAudience(params, hriClaims, gomock.Any()).
			Return(nil),
	)

	claims, err := GetValidatedClaims(params, mockValidator, nil)

	// we expect to get back a valid set of claims and no error
	expClaims := hriClaims
	var expErr map[string]interface{}

	if !reflect.DeepEqual(err, expErr) {
		t.Fatalf("Unexpected err response.\nexpected: %v\nactual  : %v", expErr, err)
	}
	if !reflect.DeepEqual(claims, expClaims) {
		t.Fatalf("Unexpected claims.\nexpected: %v\nactual  : %v", expClaims, claims)
	}
}

func TestGetValidatedClaimsTokenError(t *testing.T) {
	// create a mock validator
	controller := gomock.NewController(t)
	defer controller.Finish()
	mockValidator := NewMockValidator(controller)

	// define inputs/outputs
	params := map[string]interface{}{}
	badTokenErr := response.Error(999, "bad token")

	// define expected calls
	mockValidator.
		EXPECT().
		GetSignedToken(params, nil, gomock.Any()).
		Return(nil, badTokenErr)

	claims, err := GetValidatedClaims(params, mockValidator, nil)

	// we expect to get back an empty set of claims and a bad token error
	expClaims := HriClaims{}
	expErr := badTokenErr

	if !reflect.DeepEqual(err, expErr) {
		t.Fatalf("Unexpected err response.\nexpected: %v\nactual  : %v", expErr, err)
	}
	if !reflect.DeepEqual(claims, expClaims) {
		t.Fatalf("Unexpected claims.\nexpected: %v\nactual  : %v", expClaims, claims)
	}
}

func TestGetValidatedClaimsExtractionError(t *testing.T) {
	// create a mock validator
	controller := gomock.NewController(t)
	defer controller.Finish()
	mockValidator := NewMockValidator(controller)

	// define inputs/outputs
	params := map[string]interface{}{}
	badClaimsHolderErr := "bad claims holder"
	claimsHolder := func(c interface{}) error {
		return errors.New(badClaimsHolderErr)
	}

	// define expected calls
	mockValidator.
		EXPECT().
		GetSignedToken(params, nil, gomock.Any()).
		Return(ClaimsHolderFunc(claimsHolder), nil)

	claims, err := GetValidatedClaims(params, mockValidator, nil)

	// we expect to get back an empty set of claims and a bad claims error
	expClaims := HriClaims{}
	expErr := response.Error(http.StatusUnauthorized, badClaimsHolderErr)

	if !reflect.DeepEqual(err, expErr) {
		t.Fatalf("Unexpected err response.\nexpected: %v\nactual  : %v", expErr, err)
	}
	if !reflect.DeepEqual(claims, expClaims) {
		t.Fatalf("Unexpected claims.\nexpected: %v\nactual  : %v", expClaims, claims)
	}
}

func TestGetValidatedClaimsTenantError(t *testing.T) {
	// create a mock validator
	controller := gomock.NewController(t)
	defer controller.Finish()
	mockValidator := NewMockValidator(controller)

	// define inputs/outputs
	params := map[string]interface{}{}
	hriClaims := HriClaims{Scope: "foo bar"}
	claimsHolder := func(c interface{}) error {
		*c.(*HriClaims) = hriClaims
		return nil
	}
	badTenantErr := response.Error(999, "bad tenant")

	// define expected calls
	gomock.InOrder(
		mockValidator.
			EXPECT().
			GetSignedToken(params, nil, gomock.Any()).
			Return(ClaimsHolderFunc(claimsHolder), nil),
		mockValidator.
			EXPECT().
			CheckTenant(params, hriClaims, gomock.Any()).
			Return(badTenantErr),
	)

	claims, err := GetValidatedClaims(params, mockValidator, nil)

	// we expect to get back a valid set of claims and a bad tenant error
	expClaims := hriClaims
	expErr := badTenantErr

	if !reflect.DeepEqual(err, expErr) {
		t.Fatalf("Unexpected err response.\nexpected: %v\nactual  : %v", expErr, err)
	}
	if !reflect.DeepEqual(claims, expClaims) {
		t.Fatalf("Unexpected claims.\nexpected: %v\nactual  : %v", expClaims, claims)
	}
}

func TestGetValidatedClaimsApplicationError(t *testing.T) {
	// create a mock validator
	controller := gomock.NewController(t)
	defer controller.Finish()
	mockValidator := NewMockValidator(controller)

	// define inputs/outputs
	applicationId := "applicationId"
	params := map[string]interface{}{param.JwtAudienceId: applicationId}
	hriClaims := HriClaims{Scope: "foo bar", Audience: []string{applicationId}}
	claimsHolder := func(c interface{}) error {
		*c.(*HriClaims) = hriClaims
		return nil
	}
	badApplicationErr := response.Error(999, "bad audience")

	// define expected calls
	gomock.InOrder(
		mockValidator.
			EXPECT().
			GetSignedToken(params, nil, gomock.Any()).
			Return(ClaimsHolderFunc(claimsHolder), nil),
		mockValidator.
			EXPECT().
			CheckTenant(params, hriClaims, gomock.Any()).
			Return(nil),
		mockValidator.
			EXPECT().
			CheckAudience(params, hriClaims, gomock.Any()).
			Return(badApplicationErr),
	)

	claims, err := GetValidatedClaims(params, mockValidator, nil)

	// we expect to get back a valid set of claims and a bad tenant error
	expClaims := hriClaims
	expErr := badApplicationErr

	if !reflect.DeepEqual(err, expErr) {
		t.Fatalf("Unexpected err response.\nexpected: %v\nactual  : %v", expErr, err)
	}
	if !reflect.DeepEqual(claims, expClaims) {
		t.Fatalf("Unexpected claims.\nexpected: %v\nactual  : %v", expClaims, claims)
	}
}
