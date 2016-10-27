package server

import (
	logger "github.com/blendlabs/go-logger"
	"github.com/wcharczuk/chart-service/server/controller"
	"github.com/wcharczuk/chart-service/server/core"
	"github.com/wcharczuk/go-web"
)

const (
	// AppName is the name of the app.
	AppName = "chart-service"

	//DateFormat is the default date format.
	DateFormat = "2006-01-02"
)

func rootHandler(rc *web.RequestContext) web.ControllerResult {
	return rc.JSON(map[string]interface{}{"status": "ok!"})
}

func faviconHandler(rc *web.RequestContext) web.ControllerResult {
	rc.Response.Header().Set("Content-Type", "image/png")
	return rc.Raw([]byte{})
}

// Init inits the web app.
func Init() *web.App {
	app := web.New()
	app.SetDiagnostics(logger.NewDiagnosticsAgentFromEnvironment())
	app.Diagnostics().Writer().SetLabel(AppName)
	app.Diagnostics().EnableEvent(logger.EventInfo)
	app.SetPort(core.Config.Port())

	app.GET("/", rootHandler)
	app.GET("/favicon.ico", faviconHandler)
	app.Register(controller.Jobs{})
	app.Register(controller.Charts{})
	app.Register(controller.Equities{})
	app.Register(controller.EquityPrices{})
	app.Register(controller.Yahoo{})

	return app
}
