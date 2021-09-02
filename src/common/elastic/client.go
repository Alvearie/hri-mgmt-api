/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package elastic

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	service "github.com/IBM/resource-controller-go-sdk-generator/build/generated"
	"github.com/elastic/go-elasticsearch/v7"
	"log"
	"net/http"
	"strings"
)

// magical reference date must be used for some reason
const DateTimeFormat string = "2006-01-02T15:04:05Z"

type ResponseError struct {
	ErrorObj error

	// error code that should be returned by the hri-mgmt-api endpoint
	Code int

	// Elastic error type
	ErrorType string

	// Elastic root_cause type
	RootCause string
}

func (elasticError ResponseError) Error() string {
	if elasticError.ErrorObj == nil {
		// An error was built, but no error information was provided.
		// Return a generic error message with the statusCode
		err := fmt.Errorf(msgUnexpectedErr, elasticError.Code)
		return err.Error()
	}

	return elasticError.ErrorObj.Error()
}

func (elasticError ResponseError) LogAndBuildApiResponse(logger *log.Logger, message string) map[string]interface{} {
	err := fmt.Errorf("%s: %w", message, elasticError)
	logger.Printf("%v", err)
	return response.Error(elasticError.Code, fmt.Sprintf("%v", err))
}

// encodes a map[string]interface{} query body into a byte buffer for the elastic client
func EncodeQueryBody(queryBody map[string]interface{}) (*bytes.Buffer, error) {
	var encodedBuffer bytes.Buffer
	encoder := json.NewEncoder(&encodedBuffer)
	encoder.SetEscapeHTML(false)

	err := encoder.Encode(queryBody)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to encode query body as byte buffer: %s", err.Error()))
	}
	return &encodedBuffer, nil
}

// used to connect to actual elastic instance from serverless framework
func ClientFromParams(params map[string]interface{}) (*elasticsearch.Client, error) {

	httpsCreds, err := param.ExtractValues(params, param.BoundCreds, "databases-for-elasticsearch", "connection", "https")
	if err != nil {
		return nil, err
	}

	cert, err := param.ExtractValues(httpsCreds, "certificate")
	if err != nil {
		return nil, err
	}

	base64Cert, ok := cert["certificate_base64"].(string)
	if !ok {
		return nil, errors.New("error extracting the certificate from Elastic credentials")
	}

	hostsInterface, ok := httpsCreds["composed"].([]interface{})
	if !ok {
		return nil, errors.New("error extracting the host from Elastic credentials")
	}
	hosts := make([]string, len(hostsInterface))
	for i, entry := range hostsInterface {
		hosts[i] = entry.(string)
	}

	// decode base64-encoded cert and add it to pool
	caCert, err := base64.StdEncoding.DecodeString(base64Cert)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	config := elasticsearch.Config{
		Addresses: hosts,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}

	return fromConfig(config)
}

type ResourceControllerService interface {
	GetResourceInstance(ctx context.Context, authorization string, id string) (service.ResourceInstance, *http.Response, error)
}

func CreateResourceControllerService() ResourceControllerService {
	conf := service.NewConfiguration()
	client := service.NewAPIClient(conf)
	service := client.ResourceInstancesApi
	return service
}

func CheckElasticIAM(params map[string]interface{}, service ResourceControllerService) (bool, error) {
	//get deployment id from params

	SDKCreds, err := param.ExtractValues(params, param.BoundCreds, "databases-for-elasticsearch", "instance_administration_api")
	if err != nil {
		return false, err
	}

	deployment_id, ok := SDKCreds["deployment_id"].(string)
	if !ok {
		return false, errors.New("error extracting the deployment Id from administration api")
	}

	//get bearer token from params
	ow_headers, err := param.ExtractValues(params, "__ow_headers")
	if err != nil {
		return false, err
	}
	bearerToken, authOk := ow_headers["authorization"].(string)
	if !authOk {
		return false, errors.New(fmt.Sprintf("Required parameter authorization is missing"))
	}

	_, response, err := service.GetResourceInstance(context.Background(), bearerToken, deployment_id)

	if err != nil {
		return false, err
	} else if response.StatusCode == http.StatusOK {
		return true, nil
	} else {
		return false, errors.New("Resource controller returned status of: " + response.Status)
	}
}

// used to connect to mocked elastic instance from unit tests
func ClientFromTransport(transport http.RoundTripper) (*elasticsearch.Client, error) {
	config := elasticsearch.Config{Transport: transport}
	return fromConfig(config)
}

func fromConfig(config elasticsearch.Config) (*elasticsearch.Client, error) {
	client, err := elasticsearch.NewClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func IndexFromTenantId(tenantId string) string {
	return tenantId + "-batches"
}

func TenantIdFromIndex(tenantIndex string) string {
	return strings.TrimSuffix(tenantIndex, "-batches")
}

func TenantsFromIndices(body []map[string]interface{}) map[string]interface{} {
	var indices []interface{}
	for _, s := range body {
		idMap := make(map[string]interface{})
		index := s["index"]
		if strings.HasSuffix(index.(string), "-batches") {
			idMap["id"] = strings.TrimSuffix(index.(string), "-batches")
			indices = append(indices, idMap)
		}
	}

	tenantsMap := make(map[string]interface{})
	tenantsMap["results"] = indices
	return tenantsMap
}
