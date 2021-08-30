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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/config"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	service "github.com/IBM/resource-controller-go-sdk-generator/build/generated"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

// DateTimeFormat magical reference date must be used for some reason
const DateTimeFormat string = "2006-01-02T15:04:05Z"

type ResponseError struct {
	ErrorObj error

	// error code that should be returned by the mgmt-api endpoint
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

func (elasticError ResponseError) LogAndBuildErrorDetail(requestId string, logger logrus.FieldLogger,
	message string) *response.ErrorDetail {
	err := fmt.Errorf("%s: [%d] %w", message, elasticError.Code, elasticError)
	logger.Errorln(err.Error())
	return response.NewErrorDetail(requestId, err.Error())
}

// EncodeQueryBody encodes a map[string]interface{} query body into a byte buffer for the elastic client
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

func ClientFromConfig(config config.Config) (*elasticsearch.Client, error) {
	prefix := "ElasticClient"
	var logger = logwrapper.GetMyLogger("", prefix)
	logger.Debugln("Getting Elastic client...")

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM([]byte(config.ElasticCert))

	esConfig := elasticsearch.Config{
		Addresses: []string{config.ElasticUrl},
		Username:  config.ElasticUsername,
		Password:  config.ElasticPassword,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}
	return fromConfig(esConfig)
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

func CheckElasticIAM(elasticServiceCrn string, bearerToken string, service ResourceControllerService) (int, error) {
	_, response, err := service.GetResourceInstance(context.Background(), bearerToken, elasticServiceCrn)

	if response == nil {
		return http.StatusInternalServerError, err
	} else if response.StatusCode == http.StatusOK {
		return http.StatusOK, nil
	} else if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
		return http.StatusUnauthorized, fmt.Errorf("elastic IAM authentication returned %d : %w", response.StatusCode, err)
	} else if response.StatusCode == http.StatusNotFound {
		return http.StatusInternalServerError, fmt.Errorf("elastic IAM authentication returned 404 : %w", err)
	} else {
		return response.StatusCode, errors.New("Resource controller returned status of: " + response.Status)
	}
}

// ClientFromTransport is used to connect to mocked elastic instance from unit tests
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
