/*
 * (C) Copyright IBM Corp. 2021
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package healthcheck

import (
	"net/http"

	configPkg "github.com/Alvearie/hri-mgmt-api/common/config"
	"github.com/Alvearie/hri-mgmt-api/common/kafka"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

type Handler interface {
	HriHealthcheck(echo.Context) error
}

// This struct is designed to make unit testing easier. It has function references for the calls to backend
// logic and other methods that reach out to external services like creating the Kafka partition reader.
type theHandler struct {
	config         configPkg.Config
	hriHealthcheck func(string, *mongo.Collection, kafka.HealthChecker) (int, *response.ErrorDetail)
}

func NewHandler(config configPkg.Config) Handler {
	return &theHandler{
		config:         config,
		hriHealthcheck: GetCheck,
	}
}

func (h *theHandler) HriHealthcheck(c echo.Context) error {
	//get Logger instance
	requestId := c.Response().Header().Get(echo.HeaderXRequestID)
	prefix := "hrihealthcheck/handler"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debug("Start HRI Healthcheck Handler")

	healthChecker, err := kafka.HriHealthChecker(h.config)
	if err != nil {
		logger.Errorln(err.Error())
		return c.JSON(http.StatusInternalServerError, response.NewErrorDetail(requestId, err.Error()))
	}
	defer healthChecker.Close()

	code, errorDetail := h.hriHealthcheck(requestId, mongoApi.GetMongoCollection(h.config.MongoColName), healthChecker)
	if errorDetail != nil {
		return c.JSON(http.StatusInternalServerError, errorDetail)
	}
	return c.NoContent(code)
}
