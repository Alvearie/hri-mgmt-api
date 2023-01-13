package tenants

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Alvearie/hri-mgmt-api/common/auth"
	"github.com/Alvearie/hri-mgmt-api/common/config"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/labstack/echo/v4"
)

type Handler interface {
	//Added as part of Azure porting
	CreateTenant(echo.Context) error
	GetTenantById(echo.Context) error
	GetTenants(echo.Context) error
	DeleteTenant(echo.Context) error
}

// This struct is designed to make unit testing easier. It has function references for the calls to backend
// logic and other methods that reach out to external services like checking Elastic IAM credentials.
type theHandler struct {
	config config.Config
	//Added as part of Azure porting
	createTenant  func(string, string) (int, interface{})
	getTenantById func(string, string) (int, interface{})
	getTenants    func(string) (int, interface{})
	deleteTenant  func(string, string) (int, interface{})
	jwtValidator  auth.TenantValidator
}

func NewHandler(config config.Config) Handler {
	return &theHandler{
		config: config,
		//Added as part of Azure porting
		createTenant:  CreateTenant,
		getTenantById: GetTenantById,
		getTenants:    GetTenants,
		deleteTenant:  DeleteTenant,
		jwtValidator:  auth.NewTenantValidator(config.AzOidcIssuer, config.AzJwtAudienceId),
	}
}

func (h *theHandler) CreateTenant(c echo.Context) error {
	requestId := c.Request().Header.Get(echo.HeaderXRequestID)
	authHeader := c.Request().Header.Get(echo.HeaderAuthorization)

	prefix := "az/tenants/handler/create"
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

	//Adding explicit validation for single char - & _ to restrict tenant creation
	errMessage := "Unable to create a new tenant[" + request.TenantId + "]:[" + strconv.Itoa(http.StatusBadRequest) + "]"
	if strings.HasPrefix(request.TenantId, "_") || strings.HasPrefix(request.TenantId, "-") {
		logger.Errorln(errMessage)
		return c.JSON(http.StatusBadRequest, response.NewErrorDetail(requestId, errMessage))
	}

	//Add JWT Token validation
	errResp := h.jwtValidator.GetValidatedClaimsForTenant(requestId, authHeader)

	if errResp != nil {
		return c.JSON(errResp.Code, errResp.Body)
	}

	return c.JSON(h.createTenant(requestId, request.TenantId))
}

func (h *theHandler) GetTenants(c echo.Context) error {
	requestId := c.Request().Header.Get(echo.HeaderXRequestID)
	prefix := "tenant/handler/GetTenants"
	var logger = logwrapper.GetMyLogger(requestId, prefix)

	logger.Debugln("Start Tenant_List Handler")
	authHeader := c.Request().Header.Get(echo.HeaderAuthorization)

	errResp := h.jwtValidator.GetValidatedClaimsForTenant(requestId, authHeader)
	if errResp != nil {
		return c.JSON(errResp.Code, response.NewErrorDetail(requestId, errResp.Body.ErrorDescription))
	}

	return c.JSON(h.getTenants(requestId))

}

func (h *theHandler) GetTenantById(c echo.Context) error {
	requestId := c.Request().Header.Get(echo.HeaderXRequestID)
	prefix := "tenants/handler/GetTenantById"
	var logger = logwrapper.GetMyLogger(requestId, prefix)

	logger.Debugln("Start Tenant_GetTenantById Handler")
	authHeader := c.Request().Header.Get(echo.HeaderAuthorization)

	//jwtValidator := auth.NewTenantValidator(h.config.AzOidcIssuer, h.config.AzJwtAudienceId)

	//Add JWT Token validation
	errResp := h.jwtValidator.GetValidatedClaimsForTenant(requestId, authHeader)
	tenantId := c.Param(param.TenantId)
	if errResp != nil {
		return c.JSON(errResp.Code, errResp.Body)
	}

	return c.JSON(h.getTenantById(requestId, tenantId))
}

func (h *theHandler) DeleteTenant(c echo.Context) error {
	requestId := c.Request().Header.Get(echo.HeaderXRequestID)
	prefix := "tenant/handler/deleteTenant"
	var logger = logwrapper.GetMyLogger(requestId, prefix)

	authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
	logger.Debugln("Start Tenant_Delete Handler")

	// Extract tenantId param
	tenantId := c.Param(param.TenantId)

	// check bearer token
	//Add JWT Token validation
	errResp := h.jwtValidator.GetValidatedClaimsForTenant(requestId, authHeader)
	if errResp != nil {
		return c.JSON(errResp.Code, errResp.Body)
	}
	code, body := h.deleteTenant(requestId, tenantId)
	if body == nil {
		return c.NoContent(code)
	} else {
		return c.JSON(code, body)
	}
}
