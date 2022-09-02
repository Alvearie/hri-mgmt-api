/*
 * (C) Copyright IBM Corp. 2021
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Alvearie/hri-mgmt-api/batches"
	"github.com/Alvearie/hri-mgmt-api/common/config"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/healthcheck"
	mongo "github.com/Alvearie/hri-mgmt-api/mongoApi"
	"github.com/Alvearie/hri-mgmt-api/streams"
	"github.com/Alvearie/hri-mgmt-api/tenants"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/newrelic/go-agent/v3/integrations/nrecho-v4"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
)

func main() {
	e := echo.New()
	retCode, startServer, _ := configureMgmtServer(e, os.Args[1:])
	if retCode != 0 {
		os.Exit(retCode)
	}

	startServer()
	os.Exit(0)
}

func configureMgmtServer(e *echo.Echo, args []string) (int, func(), error) {
	configPath := "C:/hri-mgmnt-api/hri-mgmt-api/config.yml"
	config, err := config.GetConfig(configPath, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR CREATING CONFIG: %v\n", err)
		return 1, nil, err
	}

	//Initialize Logging - NOTE: we are hard-coding the log output location to Stdout
	logCfg, err := logwrapper.Initialize(config.LogLevel, os.Stdout)
	if err != nil {
		msg := "ERROR: Could NOT initialize Logger: %w"
		fmt.Fprintf(os.Stderr, fmt.Errorf(msg, err).Error()+"\n")
		return 3, nil, err //special return code for logging problems
	}
	stdFlds := map[string]string{
		logwrapper.RequestIdField:      "",
		logwrapper.FunctionPrefixField: "SERVE",
	}

	//Log the Server startup
	logger, err := logwrapper.CreateLogger(stdFlds)
	if err != nil {
		msg := "ERROR: Could NOT acquire Logger: " + err.Error()
		fmt.Fprintf(os.Stderr, msg)
		return 3, nil, err //special return code for logging problems
	}
	logger.Infof("HRI serve Startup with minimum Log Level: %s", config.LogLevel)

	// Configure Middleware

	// Normalize incoming requests ending in a slash instead of returning a 404.
	e.Pre(middleware.RemoveTrailingSlash())

	if config.NewRelicEnabled {
		app, err := newrelic.NewApplication(
			newrelic.ConfigAppName(config.NewRelicAppName),
			newrelic.ConfigLicense(config.NewRelicLicenseKey),
			func(config *newrelic.Config) {
				config.HighSecurity = true
			},
		)
		if err != nil {
			logger.Errorf("ERROR CONFIGURING NEW RELIC: %v\n", err)
			return 1, nil, err
		}
		// According to nrecho documentation, this New Relic middleware
		// should be added before any others.
		e.Use(nrecho.Middleware(app))
	}

	e.Use(
		middleware.RequestID(), // Generate a request id on the HTTP response headers
	)
	if logLvlInfoOrLess(logCfg) {
		e.Use(
			middleware.Logger(), // Log every request/response to stdout
		)
	}
	if logLvlDebugOrLess(logCfg) {
		e.Use(
			middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
				logger.Debugf("%s %s '%v' -> %d '%v'",
					c.Request().Method, c.Request().URL, string(reqBody), c.Response().Status, string(resBody))
			}),
		)
	}

	e.Use(ServerHeader)

	// Set custom binder
	customBinder, err := model.GetBinder()
	if err != nil {
		logger.Errorf("ERROR INITIALIZING BINDER: %v\n", err)
		return 1, nil, err
	}
	e.Binder = customBinder

	// Set custom validation
	customValidator, err := model.GetValidator()
	if err != nil {
		logger.Errorf("ERROR INITIALIZING VALIDATOR: %v\n", err)
		return 1, nil, err
	}
	e.Validator = customValidator

	//connect to mongo API
	err = mongo.ConnectFromConfig(config)
	if err != nil {
		e.Logger.Fatal(err)
		os.Exit(2)
	}

	// Prepare the server start function
	startFunc := func() {
		err := error(nil)
		if config.TlsEnabled {
			err = e.StartTLS(":1323", config.TlsCertPath, config.TlsKeyPath)

		} else {
			err = e.Start(":1323")
		}

		if err != nil {
			e.Logger.Fatal(err)
			os.Exit(2)
		}

	}

	// Configure the endpoint routes

	// Endpoint for ready/liveness probes
	e.GET("/alive", func(c echo.Context) error {
		return c.String(http.StatusOK, "yes")
	})

	// Healthcheck routing
	healthcheckHandler := healthcheck.NewHandler(config)
	e.GET("/hri/healthcheck", healthcheckHandler.Healthcheck)

	// Tenants routing
	tenantsHandler := tenants.NewHandler(config)
	e.GET("/hri/tenants", tenantsHandler.Get)
	e.GET(fmt.Sprintf("/hri/tenants/:%s", param.TenantId), tenantsHandler.GetById)
	e.POST(fmt.Sprintf("/hri/tenants/:%s", param.TenantId), tenantsHandler.Create)
	e.DELETE(fmt.Sprintf("/hri/tenants/:%s", param.TenantId), tenantsHandler.Delete)
	//Added as part of Azure porting
	e.POST(fmt.Sprintf("/hri/az/tenants/:%s", param.TenantId), tenantsHandler.CreateTenant)

	// Batches routing
	batchesHandler := batches.NewHandler(config)
	e.GET(fmt.Sprintf("/hri/tenants/:%s/batches/:%s", param.TenantId, param.BatchId), batchesHandler.GetById)
	e.POST(fmt.Sprintf("/hri/tenants/:%s/batches", param.TenantId), batchesHandler.Create)
	e.GET(fmt.Sprintf("/hri/tenants/:%s/batches", param.TenantId), batchesHandler.Get)
	e.PUT(fmt.Sprintf("/hri/tenants/:%s/batches/:%s/action/sendComplete",
		param.TenantId, param.BatchId), batchesHandler.SendComplete)
	e.PUT(fmt.Sprintf("/hri/tenants/:%s/batches/:%s/action/terminate",
		param.TenantId, param.BatchId), batchesHandler.Terminate)
	e.PUT(fmt.Sprintf("/hri/tenants/:%s/batches/:%s/action/processingComplete",
		param.TenantId, param.BatchId), batchesHandler.ProcessingComplete)
	e.PUT(fmt.Sprintf("/hri/tenants/:%s/batches/:%s/action/fail",
		param.TenantId, param.BatchId), batchesHandler.Fail)

	// Streams routing
	streamsHandler := streams.NewHandler(config)
	e.POST(fmt.Sprintf("hri/tenants/:%s/streams/:%s", param.TenantId, param.StreamId), streamsHandler.Create)
	e.DELETE(fmt.Sprintf("hri/tenants/:%s/streams/:%s", param.TenantId, param.StreamId), streamsHandler.Delete)
	e.GET(fmt.Sprintf("/hri/tenants/:%s/streams", param.TenantId), streamsHandler.Get)

	return 0, startFunc, nil
}

func logLvlInfoOrLess(logCfg *logwrapper.LogConfig) bool {
	return logCfg.Level == logrus.InfoLevel || logCfg.Level == logrus.DebugLevel ||
		logCfg.Level == logrus.TraceLevel
}

func logLvlDebugOrLess(logCfg *logwrapper.LogConfig) bool {
	return logCfg.Level == logrus.DebugLevel || logCfg.Level == logrus.TraceLevel
}

func ServerHeader(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'none'; connect-src 'self'; img-src 'self'; style-src 'self';")
		c.Response().Header().Set("X-Content-Type-Options", "nosniff")
		c.Response().Header().Set("X-XSS-Protection", "1")
		return next(c)
	}
}
