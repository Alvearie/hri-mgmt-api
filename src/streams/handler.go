package streams

import (
	"fmt"
	"net/http"

	"github.com/Alvearie/hri-mgmt-api/common/auth"
	configPkg "github.com/Alvearie/hri-mgmt-api/common/config"
	"github.com/Alvearie/hri-mgmt-api/common/kafka"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/labstack/echo/v4"
)

const (
	MissingHeaderMsg = "missing header 'Authorization'"
)

type Handler interface {
	//Azure porting
	CreateStream(echo.Context) error
	GetStream(echo.Context) error
	DeleteStream(echo.Context) error
}

// This struct is designed to make unit testing easier. It has function references for the calls to backend
// logic and other methods that reach out to external services like JWT token validation.
type theHandler struct {
	config       configPkg.Config
	createStream func(model.CreateStreamsRequest, string, string, bool, string, kafka.KafkaAdmin) ([]string, int, error)
	getStream    func(string, string, kafka.KafkaAdmin) (int, interface{})
	deleteStream func(string, []string, kafka.KafkaAdmin) (int, error)
	jwtValidator auth.TenantValidator
}

func NewHandler(config configPkg.Config) Handler {
	return &theHandler{
		config:       config,
		createStream: CreateStream,
		getStream:    GetStream,
		deleteStream: DeleteStream,
		jwtValidator: auth.NewTenantValidator(config.AzOidcIssuer, config.AzJwtAudienceId),
	}
}

func (h *theHandler) CreateStream(c echo.Context) error {
	requestId := c.Response().Header().Get(echo.HeaderXRequestID)
	prefix := "streams/create/handler"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debug("Start Handler-Create")

	bearerTokens := c.Request().Header[echo.HeaderAuthorization]
	if len(bearerTokens) == 0 {
		return c.JSON(http.StatusUnauthorized, response.NewErrorDetail(requestId, MissingHeaderMsg))
	}

	//Add JWT Token validation
	authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
	errResp := h.jwtValidator.GetValidatedClaimsForTenant(requestId, authHeader)

	if errResp != nil {
		return c.JSON(errResp.Code, errResp.Body)
	}
	//END

	service, errDetail := kafka.AzNewAdminClientFromConfig(h.config, bearerTokens[0])
	if errDetail != nil {
		logger.Errorln(errDetail.Body.ErrorDescription)
		return c.JSON(errDetail.Code, response.NewErrorDetail(requestId, errDetail.Body.ErrorDescription))
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

	createdTopics, returnCode, createError := h.createStream(request, request.TenantId, request.StreamId, h.config.Validation, requestId, service)
	if returnCode != http.StatusCreated {
		if len(createdTopics) > 0 {
			_, deleteError := h.deleteStream(requestId, createdTopics, service)
			if deleteError != nil {
				msg := fmt.Sprintf("%s\n%s", createError.Error(), deleteError)
				return c.JSON(returnCode, response.NewErrorDetail(requestId, msg))
			}
		}
		return c.JSON(returnCode, response.NewErrorDetail(requestId, createError.Error()))
	}

	respBody := map[string]interface{}{param.StreamId: request.StreamId}
	return c.JSON(returnCode, respBody)
}

func (h *theHandler) DeleteStream(c echo.Context) error {
	requestId := c.Response().Header().Get(echo.HeaderXRequestID)
	prefix := "streams/delete/handler"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debug("Start Handler Delete")

	bearerTokens := c.Request().Header[echo.HeaderAuthorization]
	if len(bearerTokens) == 0 {
		return c.JSON(http.StatusUnauthorized, response.NewErrorDetail(requestId, MissingHeaderMsg))
	}

	service, errDetail := kafka.AzNewAdminClientFromConfig(h.config, bearerTokens[0])
	if errDetail != nil {
		logger.Errorln(errDetail.Body.ErrorDescription)
		return c.JSON(errDetail.Code, response.NewErrorDetail(requestId, errDetail.Body.ErrorDescription))
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

	//Add JWT Token validation
	authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
	errResp := h.jwtValidator.GetValidatedClaimsForTenant(requestId, authHeader)

	if errResp != nil {
		return c.JSON(errResp.Code, errResp.Body)
	}
	//END
	inTopicName, notificationTopicName, outTopicName, invalidTopicName := kafka.CreateTopicNames(request.TenantId, request.StreamId)
	streamNames := []string{inTopicName, notificationTopicName}
	if h.config.Validation {
		streamNames = append(streamNames, outTopicName, invalidTopicName)
	}

	returnCode, err := h.deleteStream(requestId, streamNames, service)
	if err != nil {
		return c.JSON(returnCode, response.NewErrorDetail(requestId, err.Error()))
	}

	return c.NoContent(returnCode)
}

func (h *theHandler) GetStream(c echo.Context) error {
	requestId := c.Response().Header().Get(echo.HeaderXRequestID)
	prefix := "streams/list/handler"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debug("Start Handler List Streams")

	bearerTokens := c.Request().Header[echo.HeaderAuthorization]
	if len(bearerTokens) == 0 {
		logger.Errorln(MissingHeaderMsg)
		return c.JSON(http.StatusUnauthorized, response.NewErrorDetail(requestId, MissingHeaderMsg))
	}

	service, err := kafka.AzNewAdminClientFromConfig(h.config, bearerTokens[0])
	if err != nil {
		logger.Errorln(err.Body.ErrorDescription)
		return c.JSON(err.Code, response.NewErrorDetail(requestId, err.Body.ErrorDescription))
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

	//Add JWT Token validation
	authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
	errResp := h.jwtValidator.GetValidatedClaimsForTenant(requestId, authHeader)

	if errResp != nil {
		return c.JSON(errResp.Code, errResp.Body)
	}
	//END

	return c.JSON(h.getStream(requestId, request.TenantId, service))
}
