package test

import (
	"net/http"
	"net/http/httptest"

	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/labstack/echo/v4"
)

const (
	ValidTenantId string = "134340"
	ValidBatchId  string = "test-batch"
	ValidStreamId string = "porcypine-stream"
)

// GetTestServer Create an echo server that is already loaded with our binding & validation system.
func GetTestServer() *echo.Echo {
	e := echo.New()
	customBinder, _ := model.GetBinder()
	e.Binder = customBinder
	customValidator, _ := model.GetValidator()
	e.Validator = customValidator
	return e
}

// PrepareHeadersContextRecorder Set headers, create and return a context and recorder.
func PrepareHeadersContextRecorder(request *http.Request, e *echo.Echo) (echo.Context, *httptest.ResponseRecorder) {
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recorder := httptest.NewRecorder()
	ctx := e.NewContext(request, recorder)
	return ctx, recorder
}
