package httptransport

import (
	"net/http"

	api "github.com/BetelgeuseTb/betelgeuse-orbitum/pkg/api"
	"github.com/labstack/echo/v4"
	mw "github.com/oapi-codegen/echo-middleware"
)

func NewServer(adapter api.ServerInterface) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	swagger, err := api.GetSwagger()
	if err == nil {
		e.Use(mw.OapiRequestValidator(swagger))
	}

	api.RegisterHandlers(e, adapter)

	e.GET("/healthz", func(c echo.Context) error { return c.String(http.StatusOK, "ok") })
	return e
}
