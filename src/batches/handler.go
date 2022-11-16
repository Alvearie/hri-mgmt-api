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
	"github.com/Alvearie/hri-mgmt-api/common/kafka"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

const msgGetByIdErr string = "error getting current Batch Status: %s"

type Handler interface {
	CreateBatch(echo.Context) error
	GetByBatchId(echo.Context) error
	GetBatch(echo.Context) error
	SendStatusComplete(ctx echo.Context) error
	SendFail(ctx echo.Context) error
	TerminateBatch(ctx echo.Context) error
	ProcessingCompleteBatch(ctx echo.Context) error
}

type theHandler struct {
	config config.Config

	jwtBatchValidator       auth.BatchValidator
	createBatch             func(string, model.CreateBatch, auth.HriAzClaims, *mongo.Collection, kafka.Writer) (int, interface{})
	getByBatchId            func(string, model.GetByIdBatch, auth.HriAzClaims, *mongo.Collection) (int, interface{})
	getTenantByIdNoAuth     func(string, model.GetByIdBatch, auth.HriAzClaims, *mongo.Collection) (int, interface{})
	getBatch                func(string, model.GetBatch, auth.HriAzClaims, *mongo.Collection) (int, interface{})
	sendStatusComplete      func(string, *model.SendCompleteRequest, auth.HriAzClaims, *mongo.Collection, kafka.Writer, status.BatchStatus) (int, interface{})
	sendFail                func(string, *model.FailRequest, auth.HriAzClaims, *mongo.Collection, kafka.Writer, status.BatchStatus) (int, interface{})
	terminateBatch          func(string, *model.TerminateRequest, auth.HriAzClaims, *mongo.Collection, kafka.Writer, status.BatchStatus) (int, interface{})
	processingCompleteBatch func(string, *model.ProcessingCompleteRequest, auth.HriAzClaims, *mongo.Collection, kafka.Writer, status.BatchStatus) (int, interface{})
}

// NewHandler This struct is designed to make unit testing easier. It has function references for the calls to backend
// logic and other classes that reach out to external services like JWT token validation.
func NewHandler(config config.Config) Handler {
	var newHandler Handler

	if config.AuthDisabled {
		newHandler = &theHandler{

			config: config,

			getByBatchId:            GetByBatchIdNoAuth,
			getTenantByIdNoAuth:     GetByBatchIdNoAuth,
			createBatch:             CreateBatchNoAuth,
			getBatch:                GetBatchNoAuth,
			sendStatusComplete:      SendStatusCompleteNoAuth,
			sendFail:                SendFailNoAuth,
			terminateBatch:          TerminateBatchNoAuth,
			processingCompleteBatch: ProcessingCompleteBatchNoAuth,
		}

	} else {
		newHandler = &theHandler{
			config: config,

			jwtBatchValidator: auth.NewBatchValidator(config.AzOidcIssuer, config.AzJwtAudienceId),

			createBatch:             CreateBatch,
			getByBatchId:            GetByBatchId,
			getTenantByIdNoAuth:     GetByBatchIdNoAuth,
			getBatch:                GetBatch,
			sendStatusComplete:      SendStatusComplete,
			sendFail:                SendFail,
			terminateBatch:          TerminateBatch,
			processingCompleteBatch: ProcessingCompleteBatch,
		}
	}

	return newHandler
}

func (h *theHandler) SendFail(c echo.Context) error {
	requestId := c.Response().Header().Get(echo.HeaderXRequestID)
	prefix := "batches/handler/sendfail"
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

	mongoClient := mongoApi.GetMongoCollection(h.config.MongoColName)

	kafkaWriter, err := kafka.NewWriterFromAzConfig(h.config)
	if err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusInternalServerError, response.NewErrorDetail(requestId, err.Error()))
	}
	defer kafkaWriter.Close()

	var claims = auth.HriAzClaims{}
	var errResp *response.ErrorDetailResponse
	if h.config.AuthDisabled == false { //Auth Enabled
		//JWT claims validation
		claims, errResp = h.jwtBatchValidator.GetValidatedRoles(requestId,
			c.Request().Header.Get(echo.HeaderAuthorization), request.TenantId)
		if errResp != nil {
			return c.JSON(errResp.Code, errResp.Body)
		}
		logger.Debugln("Auth Enabled - call SendFail()")

	} else {
		logger.Debugln("Auth Disabled - call FailNoAuth()")
	}

	currentStatus, getStatusErr := getBatchStatus(h, requestId, getBatchRequest, mongoClient, logger)
	if getStatusErr != nil {
		return c.JSON(getStatusErr.Code, getStatusErr.Body)
	}

	code, body = h.sendFail(requestId, &request, claims, mongoClient, kafkaWriter, currentStatus)

	if body != nil {
		return c.JSON(code, body)
	}

	return c.NoContent(code)
}

func (h *theHandler) SendStatusComplete(c echo.Context) error {
	requestId := c.Response().Header().Get(echo.HeaderXRequestID)
	prefix := "batches/handler/sendStatusComplete"
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

	mongoClient := mongoApi.GetMongoCollection(h.config.MongoColName)

	kafkaWriter, err := kafka.NewWriterFromAzConfig(h.config)
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
	var claims = auth.HriAzClaims{}
	var errResp *response.ErrorDetailResponse

	if h.config.AuthDisabled == false { //Auth Enabled
		//JWT claims validation
		claims, errResp = h.jwtBatchValidator.GetValidatedRoles(requestId,
			c.Request().Header.Get(echo.HeaderAuthorization), request.TenantId)
		if errResp != nil {
			return c.JSON(errResp.Code, errResp.Body)
		}
		logger.Debugln("Auth Enabled - call SendComplete()")
	} else {
		logger.Debugln("Auth Disabled - call SendCompleteNoAuth()")
	}

	currentStatus, getStatusErr := getBatchStatus(h, requestId, getBatchRequest, mongoClient, logger)
	if getStatusErr != nil {
		return c.JSON(getStatusErr.Code, getStatusErr.Body)
	}

	code, body = h.sendStatusComplete(requestId, &request, claims, mongoClient, kafkaWriter, currentStatus)

	if body != nil {
		return c.JSON(code, body)
	}

	return c.NoContent(code)
}

// get the Current Batch Status --> Need current batch Status for potential "revert Status operation" in updateBatchStatus()
// Note: this call will Always use the empty claims (NoAuth) option for calling getTenantByIdNoAuth()
func getBatchStatus(h *theHandler, requestId string, getBatchRequest model.GetByIdBatch, mongoClient *mongo.Collection, logger logrus.FieldLogger) (status.BatchStatus, *response.ErrorDetailResponse) {

	var claims = auth.HriAzClaims{} //Always use the empty claims (NoAuth) option
	getByIdCode, responseBody := h.getTenantByIdNoAuth(requestId, getBatchRequest, claims, mongoClient)
	if getByIdCode != http.StatusOK { //error getting current Batch Info
		var errDetail = responseBody.(*response.ErrorDetail)
		newErrMsg := fmt.Sprintf(msgGetByIdErr, errDetail.ErrorDescription)
		logger.Errorln(newErrMsg)

		return status.Unknown, response.NewErrorDetailResponse(getByIdCode, requestId, newErrMsg)
	}

	currentStatus, extractErr := ExtractBatchStatus(responseBody)
	if extractErr != nil {
		errMsg := fmt.Sprintf(msgGetByIdErr, extractErr)
		logger.Errorln(errMsg)
		return status.Unknown, response.NewErrorDetailResponse(http.StatusInternalServerError, requestId, errMsg)
	}

	return currentStatus, nil
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

func (h *theHandler) TerminateBatch(c echo.Context) error {
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

	kafkaWriter, err := kafka.NewWriterFromAzConfig(h.config)
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
	var claims = auth.HriAzClaims{}
	var errResp *response.ErrorDetailResponse
	if !h.config.AuthDisabled { //Auth Enabled
		//do JWT claims validation
		claims, errResp = h.jwtBatchValidator.GetValidatedRoles(requestId,
			c.Request().Header.Get(echo.HeaderAuthorization), request.TenantId)
		if errResp != nil {
			return c.JSON(errResp.Code, errResp.Body)
		}
		logger.Debugln("Auth Enabled - call Terminate()")
	} else {
		logger.Debugln("Auth Disabled - call TerminateNoAuth()")
	}

	currentStatus, getStatusErr := getBatchStatus(h, requestId, getBatchRequest, mongoApi.GetMongoCollection(h.config.MongoColName), logger)
	if getStatusErr != nil {
		return c.JSON(getStatusErr.Code, getStatusErr.Body)
	}
	code, body = h.terminateBatch(requestId, &request, claims, mongoApi.GetMongoCollection(h.config.MongoColName), kafkaWriter, currentStatus)

	if body != nil {
		return c.JSON(code, body)
	}

	return c.NoContent(code)
}

func (h *theHandler) ProcessingCompleteBatch(c echo.Context) error {
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

	kafkaWriter, err := kafka.NewWriterFromAzConfig(h.config)
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
	var claims = auth.HriAzClaims{}
	var errResp *response.ErrorDetailResponse
	if !h.config.AuthDisabled { //Auth Enabled
		//JWT claims validation
		claims, errResp = h.jwtBatchValidator.GetValidatedRoles(requestId,
			c.Request().Header.Get(echo.HeaderAuthorization), request.TenantId)
		if errResp != nil {
			return c.JSON(errResp.Code, errResp.Body)
		}

		logger.Debugln("Auth Enabled - call ProcessingComplete()")
	} else {
		logger.Debugln("Auth Disabled - call ProcessingCompleteNoAuth()")
	}

	currentStatus, getStatusErr := getBatchStatus(h, requestId, getBatchRequest, mongoApi.GetMongoCollection(h.config.MongoColName), logger)
	if getStatusErr != nil {
		return c.JSON(getStatusErr.Code, getStatusErr.Body)
	}

	code, body = h.processingCompleteBatch(requestId, &request, claims, mongoApi.GetMongoCollection(h.config.MongoColName), kafkaWriter, currentStatus)

	if body != nil {
		return c.JSON(code, body)
	}

	return c.NoContent(code)
}
