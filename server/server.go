package server

import (
	logger "github.com/blendlabs/go-logger"
	"github.com/blendlabs/go-request"
	"github.com/blendlabs/go-web"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/chart-service/server/controller"
	"github.com/wcharczuk/chart-service/server/core"
)

const (
	// AppName is the name of the app.
	AppName = "chart-service"

	//DateFormat is the default date format.
	DateFormat = "2006-01-02"
)

func rootHandler(rc *web.Ctx) web.Result {
	return rc.JSON().Result(map[string]interface{}{"status": "ok!"})
}

func faviconHandler(rc *web.Ctx) web.Result {
	rc.Response.Header().Set("Content-Type", "image/png")
	return rc.Raw([]byte{})
}

// Init inits the web app.
func Init() *web.App {
	app := web.New()
	app.SetLogger(logger.NewFromEnvironment())
	logger.SetDefault(app.Logger())
	app.SetName(AppName)
	app.Logger().EnableEvent(logger.EventInfo)
	app.SetPort(core.Config.Port())

	app.GET("/", rootHandler)
	app.GET("/favicon.ico", faviconHandler)
	app.Register(controller.Jobs{})
	app.Register(controller.Charts{})
	app.Register(controller.Equities{})
	app.Register(controller.EquityPrices{})
	app.Register(controller.Provider{})

	app.OnStart(func(_ *web.App) error {
		if app.Logger().IsEnabled(logger.EventDebug) {
			app.Logger().AddEventListener(spiffy.EventFlagQuery, spiffy.NewPrintStatementListener())
			app.Logger().AddEventListener(spiffy.EventFlagExecute, spiffy.NewPrintStatementListener())
			app.Logger().AddEventListener(request.Event, request.NewOutgoingListener(request.WriteOutgoingRequest))
		}
		return nil
	})

	return app
}
