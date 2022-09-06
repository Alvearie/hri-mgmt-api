/*
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package tenants

import (
	"net/http"

	"github.com/Alvearie/hri-mgmt-api/common/auth"
	"github.com/Alvearie/hri-mgmt-api/common/config"
	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

type Handler interface {
	Create(echo.Context) error
	Get(echo.Context) error
	GetById(echo.Context) error
	Delete(echo.Context) error
	//Added as part of Azure porting
	CreateTenant(echo.Context) error
	GetTenantById(echo.Context) error
}

// This struct is designed to make unit testing easier. It has function references for the calls to backend
// logic and other methods that reach out to external services like checking Elastic IAM credentials.
type theHandler struct {
	config          config.Config
	checkElasticIAM func(string, string, elastic.ResourceControllerService) (int, error)
	// The Elastic Client creation doesn't have a method reference, because it does not reach out to the Elastic
	// cluster until it's used. So, we don't need to mock it for unit testing.
	create  func(string, string, *elasticsearch.Client) (int, interface{})
	get     func(string, *elasticsearch.Client) (int, interface{})
	getById func(string, string, *elasticsearch.Client) (int, interface{})
	delete  func(string, string, *elasticsearch.Client) (int, interface{})
	//Added as part of Azure porting
	createTenant  func(string, string, *mongo.Collection) (int, interface{})
	getTenantById func(string, string, *mongo.Collection) (int, interface{})
	//jwtValidator auth.Validator
}

func NewHandler(config config.Config) Handler {
	return &theHandler{
		config:          config,
		checkElasticIAM: elastic.CheckElasticIAM,
		create:          Create,
		get:             Get,
		getById:         GetById,
		delete:          Delete,
		//Added as part of Azure porting
		createTenant:  CreateTenant,
		getTenantById: GetTenantById,
	}
}

func (h *theHandler) Create(c echo.Context) error {
	requestId := c.Request().Header.Get(echo.HeaderXRequestID)
	authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
	prefix := "tenants/handler/create"
	var logger = logwrapper.GetMyLogger(requestId, prefix)

	// bind & validate request body
	var request model.CreateTenant
	if err := c.Bind(&request); err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, err.Error()))
	}
	if err := c.Validate(request); err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, err.Error()))
	}

	// check bearer token
	service := elastic.CreateResourceControllerService()
	code, err := h.checkElasticIAM(h.config.ElasticServiceCrn, authHeader, service)
	if err != nil {
		logger.Errorln(err.Error())
		return c.JSON(code, response.NewErrorDetail(requestId, err.Error()))
	}

	esClient, err := elastic.ClientFromConfig(h.config)
	if err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusInternalServerError, response.NewErrorDetail(requestId, err.Error()))
	}

	return c.JSON(h.create(requestId, request.TenantId, esClient))
}

func (h *theHandler) CreateTenant(c echo.Context) error {
	requestId := c.Request().Header.Get(echo.HeaderXRequestID)
	authHeader := c.Request().Header.Get(echo.HeaderAuthorization)

	prefix := "az/tenants/handler/create"
	var logger = logwrapper.GetMyLogger(requestId, prefix)

	jwtValidator := auth.NewTenantValidator(h.config.AzOidcIssuer, h.config.AzJwtAudienceId)
	// bind & validate request body
	var request model.CreateTenant
	if err := c.Bind(&request); err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, err.Error()))
	}
	if err := c.Validate(request); err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, err.Error()))
	}

	//Add JWT Token validation

	errResp := jwtValidator.GetValidatedClaimsForTenant(requestId, authHeader)

	if errResp != nil {
		return c.JSON(errResp.Code, errResp.Body)
	}

	return c.JSON(h.createTenant(requestId, request.TenantId, mongoApi.GetMongoCollection(h.config.MongoColName)))
}

func (h *theHandler) Get(c echo.Context) error {
	requestId := c.Request().Header.Get(echo.HeaderXRequestID)
	prefix := "tenant/handler/get"
	var logger = logwrapper.GetMyLogger(requestId, prefix)

	logger.Debugln("Start Tenant_Get Handler")
	authHeader := c.Request().Header.Get(echo.HeaderAuthorization)

	// check bearer token
	service := elastic.CreateResourceControllerService()
	code, err := h.checkElasticIAM(h.config.ElasticServiceCrn, authHeader, service)
	if err != nil {
		return c.JSON(code, response.NewErrorDetail(requestId, err.Error()))
	}

	esClient, err := elastic.ClientFromConfig(h.config)
	if err != nil {
		msg := err.Error()
		logger.Errorln(msg)
		return c.JSON(http.StatusInternalServerError, response.NewErrorDetail(requestId, msg))
	}

	return c.JSON(h.get(requestId, esClient))
}

func (h *theHandler) GetById(c echo.Context) error {
	requestId := c.Request().Header.Get(echo.HeaderXRequestID)
	prefix := "tenant/handler/GetById"
	var logger = logwrapper.GetMyLogger(requestId, prefix)

	logger.Debugln("Start Tenant_GetById Handler")
	authHeader := c.Request().Header.Get(echo.HeaderAuthorization)

	// check bearer token
	service := elastic.CreateResourceControllerService()
	code, err := h.checkElasticIAM(h.config.ElasticServiceCrn, authHeader, service)
	if err != nil {
		msg := err.Error()
		logger.Errorln(msg)
		return c.JSON(code, response.NewErrorDetail(requestId, msg))
	}

	tenantId := c.Param(param.TenantId)
	esClient, err := elastic.ClientFromConfig(h.config)
	if err != nil {
		msg := err.Error()
		logger.Errorln(msg)
		return c.JSON(http.StatusInternalServerError, response.NewErrorDetail(requestId, msg))
	}

	return c.JSON(h.getById(requestId, tenantId, esClient))
}

func (h *theHandler) GetTenantById(c echo.Context) error {
	requestId := c.Request().Header.Get(echo.HeaderXRequestID)
	prefix := "az/tenants/handler/GetTenantById"
	var logger = logwrapper.GetMyLogger(requestId, prefix)

	logger.Debugln("Start Tenant_GetTenantById Handler")
	authHeader := c.Request().Header.Get(echo.HeaderAuthorization)

	jwtValidator := auth.NewTenantValidator(h.config.AzOidcIssuer, h.config.AzJwtAudienceId)

	//Add JWT Token validation
	errResp := jwtValidator.GetValidatedClaimsForTenant(requestId, authHeader)
	tenantId := c.Param(param.TenantId)
	if errResp != nil {
		return c.JSON(errResp.Code, errResp.Body)
	}

	return c.JSON(h.getTenantById(requestId, tenantId, mongoApi.GetMongoCollection(h.config.MongoColName)))
}

func (h *theHandler) Delete(c echo.Context) error {
	requestId := c.Request().Header.Get(echo.HeaderXRequestID)
	prefix := "tenant/handler/delete"
	var logger = logwrapper.GetMyLogger(requestId, prefix)

	authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
	logger.Debugln("Start Tenant_Delete Handler")

	// Extract tenantId param
	tenantId := c.Param(param.TenantId)

	// check bearer token
	service := elastic.CreateResourceControllerService()
	code, err := h.checkElasticIAM(h.config.ElasticServiceCrn, authHeader, service)
	if err != nil {
		msg := err.Error()
		logger.Errorln(msg)
		return c.JSON(code, response.NewErrorDetail(requestId, msg))
	}

	esClient, err := elastic.ClientFromConfig(h.config)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.NewErrorDetail(requestId, err.Error()))
	}

	code, body := h.delete(requestId, tenantId, esClient)
	if body == nil {
		return c.NoContent(code)
	} else {
		return c.JSON(code, body)
	}
}
