/*
 * (C) Copyright IBM Corp. 2021
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package streams

import (
	"fmt"
	configPkg "github.com/Alvearie/hri-mgmt-api/common/config"
	"github.com/Alvearie/hri-mgmt-api/common/kafka"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler interface {
	Create(echo.Context) error
	Get(echo.Context) error
	Delete(echo.Context) error
}

// This struct is designed to make unit testing easier. It has function references for the calls to backend
// logic and other methods that reach out to external services like JWT token validation.
type theHandler struct {
	config configPkg.Config
	create func(model.CreateStreamsRequest, string, string, bool, string, kafka.KafkaAdmin) ([]string, int, error)
	delete func(string, []string, kafka.KafkaAdmin) (int, error)
	get    func(string, string, kafka.KafkaAdmin) (int, interface{})
}

func NewHandler(config configPkg.Config) Handler {
	return &theHandler{
		config: config,
		create: Create,
		get:    Get,
		delete: Delete,
	}
}

func (h *theHandler) Create(c echo.Context) error {
	requestId := c.Response().Header().Get(echo.HeaderXRequestID)
	prefix := "streams/create/handler"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debug("Start Handler-Create")

	bearerTokens := c.Request().Header[echo.HeaderAuthorization]
	if len(bearerTokens) == 0 {
		return c.JSON(http.StatusUnauthorized, response.NewErrorDetail(requestId, "missing header 'Authorization'"))
	}
	service, err := kafka.NewAdminClientFromConfig(h.config, bearerTokens[0])
	if err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusInternalServerError, response.NewErrorDetail(requestId, err.Error()))
	}

	// bind & validate request body
	var request model.CreateStreamsRequest
	if err := c.Bind(&request); err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, err.Error()))
	}
	if err := c.Validate(request); err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, err.Error()))
	}

	createdTopics, returnCode, createError := h.create(request, request.TenantId, request.StreamId, h.config.Validation, requestId, service)
	if returnCode != http.StatusCreated {
		_, deleteError := h.delete(requestId, createdTopics, service)
		if deleteError != nil {
			msg := fmt.Sprintf("%s\n%s", createError.Error(), deleteError)
			return c.JSON(returnCode, response.NewErrorDetail(requestId, msg))
		} else {
			return c.JSON(returnCode, response.NewErrorDetail(requestId, createError.Error()))
		}
	}

	respBody := map[string]interface{}{param.StreamId: request.StreamId}
	return c.JSON(returnCode, respBody)
}

func (h *theHandler) Delete(c echo.Context) error {
	requestId := c.Response().Header().Get(echo.HeaderXRequestID)
	prefix := "streams/delete/handler"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debug("Start Handler Delete")

	bearerTokens := c.Request().Header[echo.HeaderAuthorization]
	if len(bearerTokens) == 0 {
		return c.JSON(http.StatusUnauthorized, response.NewErrorDetail(requestId, "missing header 'Authorization'"))
	}

	service, err := kafka.NewAdminClientFromConfig(h.config, bearerTokens[0])
	if err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusInternalServerError, response.NewErrorDetail(requestId, err.Error()))
	}

	// bind & validate request body
	var request model.DeleteStreamRequest
	if err := c.Bind(&request); err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, err.Error()))
	}
	if err := c.Validate(request); err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, err.Error()))
	}

	inTopicName, notificationTopicName, outTopicName, invalidTopicName := kafka.CreateTopicNames(request.TenantId, request.StreamId)
	streamNames := []string{inTopicName, notificationTopicName}
	if h.config.Validation {
		streamNames = append(streamNames, outTopicName, invalidTopicName)
	}

	returnCode, err := h.delete(requestId, streamNames, service)
	if err != nil {
		return c.JSON(returnCode, response.NewErrorDetail(requestId, err.Error()))
	}

	return c.NoContent(returnCode)
}

func (h *theHandler) Get(c echo.Context) error {
	requestId := c.Response().Header().Get(echo.HeaderXRequestID)
	prefix := "streams/list/handler"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debug("Start Handler List Streams")

	bearerTokens := c.Request().Header[echo.HeaderAuthorization]
	if len(bearerTokens) == 0 {
		msg := "missing header 'Authorization'"
		logger.Errorln(msg)
		return c.JSON(http.StatusUnauthorized, response.NewErrorDetail(requestId, msg))
	}

	service, err := kafka.NewAdminClientFromConfig(h.config, bearerTokens[0])
	if err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusInternalServerError, response.NewErrorDetail(requestId, err.Error()))
	}

	// bind & validate request body
	var request model.GetStreamRequest
	if err := c.Bind(&request); err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, err.Error()))
	}
	if err := c.Validate(request); err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, err.Error()))
	}

	return c.JSON(h.get(requestId, request.TenantId, service))
}
