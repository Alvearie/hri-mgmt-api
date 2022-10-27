/*
 * (C) Copyright IBM Corp. 2021
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package batches

import (
	"fmt"
	"net/http"

	"github.com/Alvearie/hri-mgmt-api/batches/status"
	"github.com/Alvearie/hri-mgmt-api/common/auth"
	"github.com/Alvearie/hri-mgmt-api/common/config"
	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/Alvearie/hri-mgmt-api/common/kafka"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

const msgElasticErr string = "error getting Elastic client: %s"
const msgGetByIdErr string = "error getting current Batch Status: %s"

type Handler interface {
	Create(echo.Context) error
	Get(echo.Context) error
	GetById(echo.Context) error
	SendComplete(ctx echo.Context) error
	Terminate(ctx echo.Context) error
	ProcessingComplete(ctx echo.Context) error
	Fail(ctx echo.Context) error
	CreateBatch(echo.Context) error
	GetByBatchId(echo.Context) error
	GetBatch(echo.Context) error
}

type theHandler struct {
	config       config.Config
	jwtValidator auth.Validator

	create             func(string, model.CreateBatch, auth.HriClaims, *elasticsearch.Client, kafka.Writer) (int, interface{})
	get                func(string, model.GetBatch, auth.HriClaims, *elasticsearch.Client) (int, interface{})
	getById            func(string, model.GetByIdBatch, auth.HriClaims, *elasticsearch.Client) (int, interface{})
	getByIdNoAuth      func(string, model.GetByIdBatch, auth.HriClaims, *elasticsearch.Client) (int, interface{})
	sendComplete       func(string, *model.SendCompleteRequest, auth.HriClaims, *elasticsearch.Client, kafka.Writer, status.BatchStatus) (int, interface{})
	terminate          func(string, *model.TerminateRequest, auth.HriClaims, *elasticsearch.Client, kafka.Writer, status.BatchStatus) (int, interface{})
	processingComplete func(string, *model.ProcessingCompleteRequest, auth.HriClaims, *elasticsearch.Client, kafka.Writer, status.BatchStatus) (int, interface{})
	fail               func(string, *model.FailRequest, auth.HriClaims, *elasticsearch.Client, kafka.Writer, status.BatchStatus) (int, interface{})
	//Azure porting
	jwtBatchValidator   auth.BatchValidator
	createBatch         func(string, model.CreateBatch, auth.HriAzClaims, *mongo.Collection, kafka.Writer) (int, interface{})
	getByBatchId        func(string, model.GetByIdBatch, auth.HriAzClaims, *mongo.Collection) (int, interface{})
	getTenantByIdNoAuth func(string, model.GetByIdBatch, auth.HriAzClaims, *mongo.Collection) (int, interface{})
	getBatch            func(string, model.GetBatch, auth.HriAzClaims, *mongo.Collection) (int, interface{})
}

// NewHandler This struct is designed to make unit testing easier. It has function references for the calls to backend
// logic and other classes that reach out to external services like JWT token validation.
func NewHandler(config config.Config) Handler {
	var newHandler Handler

	if config.AuthDisabled {
		newHandler = &theHandler{

			config:              config,
			create:              CreateNoAuth,
			get:                 GetNoAuth,
			getById:             GetByIdNoAuth,
			getByIdNoAuth:       GetByIdNoAuth,
			sendComplete:        SendCompleteNoAuth,
			terminate:           TerminateNoAuth,
			processingComplete:  ProcessingCompleteNoAuth,
			fail:                FailNoAuth,
			getByBatchId:        GetByBatchIdNoAuth,
			getTenantByIdNoAuth: GetByBatchIdNoAuth,
			createBatch:         CreateBatchNoAuth,
			getBatch:            GetBatchNoAuth,
		}

	} else {
		newHandler = &theHandler{
			config:            config,
			jwtValidator:      auth.NewValidator(config.OidcIssuer, config.JwtAudienceId),
			jwtBatchValidator: auth.NewBatchValidator(config.AzOidcIssuer, config.AzJwtAudienceId),

			// The Elastic Client & Kafka Writer creation don't have method references, because they do not reach out to the
			// service until they're used. So, we don't need to mock it for unit testing.
			create:             Create,
			get:                Get,
			getById:            GetById,
			getByIdNoAuth:      GetByIdNoAuth, //Needed for the getCurrentBatchStatus() call for action endpoints
			sendComplete:       SendComplete,
			terminate:          Terminate,
			processingComplete: ProcessingComplete,
			fail:               Fail,

			createBatch:         CreateBatch,
			getByBatchId:        GetByBatchId,
			getTenantByIdNoAuth: GetByBatchIdNoAuth,
			getBatch:            GetBatch,
		}
	}

	return newHandler
}
func (h *theHandler) GetByBatchId(c echo.Context) error {
	requestId := c.Response().Header().Get(echo.HeaderXRequestID)
	prefix := "batches/handler/getById"
	var logger = logwrapper.GetMyLogger(requestId, prefix)

	// bind & validate request body
	var request model.GetByIdBatch
	if err := c.Bind(&request); err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, err.Error()))
	}
	if err := c.Validate(request); err != nil {
		logger.Printf(err.Error())
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, err.Error()))
	}

	mongoClient := mongoApi.GetMongoCollection(h.config.MongoColName)
	if h.config.AuthDisabled == false { //Auth Enabled
		//JWT claims validation

		claims, errResp := h.jwtBatchValidator.GetValidatedRoles(requestId,
			c.Request().Header.Get(echo.HeaderAuthorization), request.TenantId)
		if errResp != nil {
			return c.JSON(errResp.Code, errResp.Body)
		}

		return c.JSON(h.getByBatchId(requestId, request, claims, mongoClient))
	} else {
		logger.Debugln("Auth Disabled - calling GetByBatchIdNoAuth()")
		var emptyClaims = auth.HriAzClaims{}
		return c.JSON(h.getTenantByIdNoAuth(requestId, request, emptyClaims, mongoClient))
	}
}
func (h *theHandler) Create(c echo.Context) error {
	requestId := c.Response().Header().Get(echo.HeaderXRequestID)
	prefix := "batches/handler/create"
	var logger = logwrapper.GetMyLogger(requestId, prefix)

	// bind & validate request body
	var batch model.CreateBatch
	if err := c.Bind(&batch); err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, err.Error()))
	}
	if err := c.Validate(batch); err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, err.Error()))
	}

	esClient, err := elastic.ClientFromConfig(h.config)
	if err != nil {
		msg := fmt.Sprintf(msgElasticErr, err.Error())
		logger.Errorln(msg)
		return c.JSON(http.StatusInternalServerError, response.NewErrorDetail(requestId, msg))
	}

	kafkaWriter, err := kafka.NewWriterFromConfig(h.config)
	if err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusInternalServerError, response.NewErrorDetail(requestId, err.Error()))
	}
	defer kafkaWriter.Close()

	if h.config.AuthDisabled == false { //Auth Enabled
		//JWT claims validation
		claims, errResp := h.jwtValidator.GetValidatedClaims(requestId,
			c.Request().Header.Get(echo.HeaderAuthorization), batch.TenantId)
		if errResp != nil {
			return c.JSON(errResp.Code, errResp.Body)
		}

		return c.JSON(h.create(requestId, batch, claims, esClient, kafkaWriter))
	} else {
		logger.Debugln("Auth Disabled - calling CreateNoAuth()")
		return c.JSON(h.create(requestId, batch, auth.HriClaims{}, esClient, kafkaWriter))
	}
}

func (h *theHandler) GetById(c echo.Context) error {
	requestId := c.Response().Header().Get(echo.HeaderXRequestID)
	prefix := "batches/handler/getById"
	var logger = logwrapper.GetMyLogger(requestId, prefix)

	// bind & validate request body
	var request model.GetByIdBatch
	if err := c.Bind(&request); err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, err.Error()))
	}
	if err := c.Validate(request); err != nil {
		logger.Printf(err.Error())
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, err.Error()))
	}

	esClient, err := elastic.ClientFromConfig(h.config)
	if err != nil {
		msg := fmt.Sprintf(msgElasticErr, err.Error())
		logger.Errorln(msg)
		return c.JSON(http.StatusInternalServerError, response.NewErrorDetail(requestId, msg))
	}

	if h.config.AuthDisabled == false { //Auth Enabled
		//JWT claims validation
		claims, errResp := h.jwtValidator.GetValidatedClaims(requestId,
			c.Request().Header.Get(echo.HeaderAuthorization), request.TenantId)
		if errResp != nil {
			return c.JSON(errResp.Code, response.NewErrorDetail(requestId, errResp.Body.ErrorDescription))
		}

		return c.JSON(h.getById(requestId, request, claims, esClient))
	} else {
		logger.Debugln("Auth Disabled - calling GetByIdNoAuth()")
		var emptyClaims = auth.HriClaims{}
		return c.JSON(h.getById(requestId, request, emptyClaims, esClient))
	}
}

func (h *theHandler) Get(c echo.Context) error {
	requestId := c.Response().Header().Get(echo.HeaderXRequestID)
	prefix := "batches/handler/get"
	var logger = logwrapper.GetMyLogger(requestId, prefix)

	// bind & validate request body
	var request model.GetBatch
	if err := c.Bind(&request); err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, err.Error()))
	}
	if err := c.Validate(request); err != nil {
		logger.Printf(err.Error())
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, err.Error()))
	}

	esClient, err := elastic.ClientFromConfig(h.config)
	if err != nil {
		msg := fmt.Sprintf(msgElasticErr, err.Error())
		logger.Errorln(msg)
		return c.JSON(http.StatusInternalServerError, response.NewErrorDetail(requestId, msg))
	}

	if h.config.AuthDisabled == false { //Auth Enabled
		//JWT claims validation
		claims, errResp := h.jwtValidator.GetValidatedClaims(requestId,
			c.Request().Header.Get(echo.HeaderAuthorization), request.TenantId)
		if errResp != nil {
			return c.JSON(errResp.Code, response.NewErrorDetail(requestId, errResp.Body.ErrorDescription))
		}

		return c.JSON(h.get(requestId, request, claims, esClient))
	} else {
		logger.Debugln("Auth Disabled - calling GetNoAuth()")
		var emptyClaims = auth.HriClaims{}
		return c.JSON(h.get(requestId, request, emptyClaims, esClient))
	}
}

func (h *theHandler) SendComplete(c echo.Context) error {
	requestId := c.Response().Header().Get(echo.HeaderXRequestID)
	prefix := "batches/handler/sendComplete"
	var logger = logwrapper.GetMyLogger(requestId, prefix)

	// bind & validate request body
	var request model.SendCompleteRequest
	if err := c.Bind(&request); err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, err.Error()))
	}
	if err := c.Validate(request); err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, err.Error()))
	}

	request.Validation = h.config.Validation

	esClient, err := elastic.ClientFromConfig(h.config)
	if err != nil {
		msg := fmt.Sprintf(msgElasticErr, err.Error())
		logger.Errorln(msg)
		return c.JSON(http.StatusInternalServerError, response.NewErrorDetail(requestId, msg))
	}

	kafkaWriter, err := kafka.NewWriterFromConfig(h.config)
	if err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusInternalServerError, response.NewErrorDetail(requestId, err.Error()))
	}
	defer kafkaWriter.Close()

	getBatchRequest := model.GetByIdBatch{
		TenantId: request.TenantId,
		BatchId:  request.BatchId,
	}
	var code int
	var body interface{}
	var claims = auth.HriClaims{}
	var errResp *response.ErrorDetailResponse
	if h.config.AuthDisabled == false { //Auth Enabled
		//JWT claims validation
		claims, errResp = h.jwtValidator.GetValidatedClaims(requestId,
			c.Request().Header.Get(echo.HeaderAuthorization), request.TenantId)
		if errResp != nil {
			return c.JSON(errResp.Code, errResp.Body)
		}
		logger.Debugln("Auth Enabled - call SendComplete()")
	} else {
		logger.Debugln("Auth Disabled - call SendCompleteNoAuth()")
	}

	currentStatus, getStatusErr := getCurrentBatchStatus(h, requestId, getBatchRequest, esClient, logger)
	if getStatusErr != nil {
		return c.JSON(getStatusErr.Code, getStatusErr.Body)
	}

	code, body = h.sendComplete(requestId, &request, claims, esClient, kafkaWriter, currentStatus)

	if body != nil {
		return c.JSON(code, body)
	}

	return c.NoContent(code)
}

func (h *theHandler) Terminate(c echo.Context) error {
	requestId := c.Response().Header().Get(echo.HeaderXRequestID)
	prefix := "batches/handler/terminate"
	var logger = logwrapper.GetMyLogger(requestId, prefix)

	// bind & validate request body
	var request model.TerminateRequest
	if err := c.Bind(&request); err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, err.Error()))
	}
	if err := c.Validate(request); err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, err.Error()))
	}

	esClient, err := elastic.ClientFromConfig(h.config)
	if err != nil {
		msg := fmt.Sprintf(msgElasticErr, err.Error())
		logger.Errorln(msg)
		return c.JSON(http.StatusInternalServerError, response.NewErrorDetail(requestId, msg))
	}

	kafkaWriter, err := kafka.NewWriterFromConfig(h.config)
	if err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusInternalServerError, response.NewErrorDetail(requestId, err.Error()))
	}
	defer kafkaWriter.Close()

	getBatchRequest := model.GetByIdBatch{
		TenantId: request.TenantId,
		BatchId:  request.BatchId,
	}
	var code int
	var body interface{}
	var claims = auth.HriClaims{}
	var errResp *response.ErrorDetailResponse
	if h.config.AuthDisabled == false { //Auth Enabled
		//do JWT claims validation
		claims, errResp = h.jwtValidator.GetValidatedClaims(requestId,
			c.Request().Header.Get(echo.HeaderAuthorization), request.TenantId)
		if errResp != nil {
			return c.JSON(errResp.Code, errResp.Body)
		}

		logger.Debugln("Auth Enabled - call Terminate()")
	} else {
		logger.Debugln("Auth Disabled - call TerminateNoAuth()")
	}

	currentStatus, getStatusErr := getCurrentBatchStatus(h, requestId, getBatchRequest, esClient, logger)
	if getStatusErr != nil {
		return c.JSON(getStatusErr.Code, getStatusErr.Body)
	}
	code, body = h.terminate(requestId, &request, claims, esClient, kafkaWriter, currentStatus)

	if body != nil {
		return c.JSON(code, body)
	}

	return c.NoContent(code)
}

func (h *theHandler) ProcessingComplete(c echo.Context) error {
	requestId := c.Response().Header().Get(echo.HeaderXRequestID)
	prefix := "batches/handler/processingComplete"
	var logger = logwrapper.GetMyLogger(requestId, prefix)

	// bind & validate request body
	var request model.ProcessingCompleteRequest
	if err := c.Bind(&request); err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, err.Error()))
	}
	if err := c.Validate(request); err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, err.Error()))
	}

	esClient, err := elastic.ClientFromConfig(h.config)
	if err != nil {
		msg := fmt.Sprintf(msgElasticErr, err.Error())
		logger.Errorln(msg)
		return c.JSON(http.StatusInternalServerError, response.NewErrorDetail(requestId, msg))
	}

	kafkaWriter, err := kafka.NewWriterFromConfig(h.config)
	if err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusInternalServerError, response.NewErrorDetail(requestId, err.Error()))
	}
	defer kafkaWriter.Close()

	getBatchRequest := model.GetByIdBatch{
		TenantId: request.TenantId,
		BatchId:  request.BatchId,
	}
	var code int
	var body interface{}
	var claims = auth.HriClaims{}
	var errResp *response.ErrorDetailResponse
	if h.config.AuthDisabled == false { //Auth Enabled
		//JWT claims validation
		claims, errResp = h.jwtValidator.GetValidatedClaims(requestId,
			c.Request().Header.Get(echo.HeaderAuthorization), request.TenantId)
		if errResp != nil {
			return c.JSON(errResp.Code, errResp.Body)
		}

		logger.Debugln("Auth Enabled - call ProcessingComplete()")
	} else {
		logger.Debugln("Auth Disabled - call ProcessingCompleteNoAuth()")
	}

	currentStatus, getStatusErr := getCurrentBatchStatus(h, requestId, getBatchRequest, esClient, logger)
	if getStatusErr != nil {
		return c.JSON(getStatusErr.Code, getStatusErr.Body)
	}

	code, body = h.processingComplete(requestId, &request, claims, esClient, kafkaWriter, currentStatus)

	if body != nil {
		return c.JSON(code, body)
	}

	return c.NoContent(code)
}

func (h *theHandler) Fail(c echo.Context) error {
	requestId := c.Response().Header().Get(echo.HeaderXRequestID)
	prefix := "batches/handler/fail"
	var logger = logwrapper.GetMyLogger(requestId, prefix)

	// bind & validate request body
	var request model.FailRequest
	if err := c.Bind(&request); err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, err.Error()))
	}
	if err := c.Validate(request); err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, err.Error()))
	}

	getBatchRequest := model.GetByIdBatch{
		TenantId: request.TenantId,
		BatchId:  request.BatchId,
	}
	var code int
	var body interface{}
	esClient, err := elastic.ClientFromConfig(h.config)
	if err != nil {
		msg := fmt.Sprintf(msgElasticErr, err.Error())
		logger.Errorln(msg)
		return c.JSON(http.StatusInternalServerError, response.NewErrorDetail(requestId, msg))
	}

	kafkaWriter, err := kafka.NewWriterFromConfig(h.config)
	if err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusInternalServerError, response.NewErrorDetail(requestId, err.Error()))
	}
	defer kafkaWriter.Close()

	var claims = auth.HriClaims{}
	var errResp *response.ErrorDetailResponse
	if h.config.AuthDisabled == false { //Auth Enabled
		//JWT claims validation
		claims, errResp = h.jwtValidator.GetValidatedClaims(requestId,
			c.Request().Header.Get(echo.HeaderAuthorization), request.TenantId)
		if errResp != nil {
			return c.JSON(errResp.Code, errResp.Body)
		}
		logger.Debugln("Auth Enabled - call Fail()")

	} else {
		logger.Debugln("Auth Disabled - call FailNoAuth()")
	}

	currentStatus, getStatusErr := getCurrentBatchStatus(h, requestId, getBatchRequest, esClient, logger)
	if getStatusErr != nil {
		return c.JSON(getStatusErr.Code, getStatusErr.Body)
	}

	code, body = h.fail(requestId, &request, claims, esClient, kafkaWriter, currentStatus)

	if body != nil {
		return c.JSON(code, body)
	}

	return c.NoContent(code)
}

// get the Current Batch Status --> Need current batch Status for potential "revert Status operation" in updateStatus()
// Note: this call will Always use the empty claims (NoAuth) option for calling GetById()
func getCurrentBatchStatus(h *theHandler, requestId string, getBatchRequest model.GetByIdBatch, esClient *elasticsearch.Client, logger logrus.FieldLogger) (status.BatchStatus, *response.ErrorDetailResponse) {

	var claims = auth.HriClaims{} //Always use the empty claims (NoAuth) option
	getByIdCode, getByIdBody := h.getByIdNoAuth(requestId, getBatchRequest, claims, esClient)
	if getByIdCode != http.StatusOK { //error getting current Batch Info
		var errDetail = getByIdBody.(*response.ErrorDetail)
		newErrMsg := fmt.Sprintf(msgGetByIdErr, errDetail.ErrorDescription)
		logger.Errorln(newErrMsg)

		return status.Unknown, response.NewErrorDetailResponse(getByIdCode, requestId, newErrMsg)
	}

	currentStatus, extractErr := ExtractBatchStatus(getByIdBody)
	if extractErr != nil {
		errMsg := fmt.Sprintf(msgGetByIdErr, extractErr)
		logger.Errorln(errMsg)
		return status.Unknown, response.NewErrorDetailResponse(http.StatusInternalServerError, requestId, errMsg)
	}
	return currentStatus, nil
}

// Added as part of Azure porting
func (h *theHandler) CreateBatch(c echo.Context) error {
	requestId := c.Response().Header().Get(echo.HeaderXRequestID)
	prefix := "batches/handler/create"
	var logger = logwrapper.GetMyLogger(requestId, prefix)

	// bind & validate request body
	var batch model.CreateBatch
	if err := c.Bind(&batch); err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, err.Error()))
	}
	if err := c.Validate(batch); err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, err.Error()))
	}

	kafkaWriter, err := kafka.NewWriterFromAzConfig(h.config)
	if err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusInternalServerError, response.NewErrorDetail(requestId, err.Error()))
	}
	defer kafkaWriter.Close()

	if h.config.AuthDisabled == false { //Auth Enabled
		//JWT claims validation
		claims, errResp := h.jwtBatchValidator.GetValidatedRoles(requestId,
			c.Request().Header.Get(echo.HeaderAuthorization), batch.TenantId)
		if errResp != nil {
			return c.JSON(errResp.Code, errResp.Body)
		}

		return c.JSON(h.createBatch(requestId, batch, claims, mongoApi.GetMongoCollection(h.config.MongoColName), kafkaWriter))
	} else {
		logger.Debugln("Auth Disabled - calling CreateBatchNoAuth()")
		return c.JSON(h.createBatch(requestId, batch, auth.HriAzClaims{}, mongoApi.GetMongoCollection(h.config.MongoColName), kafkaWriter))
	}
}

func (h *theHandler) GetBatch(c echo.Context) error {
	requestId := c.Response().Header().Get(echo.HeaderXRequestID)
	prefix := "batches/handler/get"
	var logger = logwrapper.GetMyLogger(requestId, prefix)

	// bind & validate request body
	var request model.GetBatch
	if err := c.Bind(&request); err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, err.Error()))
	}
	if err := c.Validate(request); err != nil {
		logger.Printf(err.Error())
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, err.Error()))
	}

	if h.config.AuthDisabled == false { //Auth Enabled
		//JWT claims validation
		claims, errResp := h.jwtBatchValidator.GetValidatedRoles(requestId,
			c.Request().Header.Get(echo.HeaderAuthorization), request.TenantId)
		if errResp != nil {
			return c.JSON(errResp.Code, response.NewErrorDetail(requestId, errResp.Body.ErrorDescription))
		}

		return c.JSON(h.getBatch(requestId, request, claims, mongoApi.GetMongoCollection(h.config.MongoColName)))
	} else {
		logger.Debugln("Auth Disabled - calling GetNoAuth()")
		var emptyClaims = auth.HriAzClaims{}
		return c.JSON(h.getBatch(requestId, request, emptyClaims, mongoApi.GetMongoCollection(h.config.MongoColName)))
	}
}
