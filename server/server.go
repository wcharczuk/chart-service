package server

import (
	"time"

	"github.com/wcharczuk/chart-service/server/core"
	"github.com/wcharczuk/chart-service/server/yahoo"
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

func stockHandler(rc *web.RequestContext) web.ControllerResult {
	ticker, err := rc.RouteParameter("ticker")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}

	stockInfos, err := yahoo.GetStockPrice([]string{ticker})
	if err != nil {
		return rc.API().InternalError(err)
	}
	if len(stockInfos) == 0 || stockInfos[0].IsZero() {
		return rc.Raw([]byte{})
	}

	stock := stockInfos[0]

	return rc.Raw([]byte{})
}

func apiStockPriceHistoricalHandler(rc *web.RequestContext) web.ControllerResult {
	ticker, err := rc.RouteParameter("ticker")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}

	from := time.Now().UTC().AddDate(0, -1, 0)
	to := time.Now().UTC()

	timeframe, err := rc.RouteParameter("timeframe")
	if err == nil {
		from, to, err = core.ParseTimeFrame(timeframe)
		if err != nil {
			return rc.API().BadRequest(err.Error())
		}
	} else {
		fromParam, err := rc.QueryParamTime("from", DateFormat)
		if err == nil {
			from = fromParam
		}

		toParam, err := rc.QueryParamTime("to", DateFormat)
		if err == nil {
			to = toParam
		}
	}

	prices, err := yahoo.GetHistoricalPrices(ticker, from, to)
	if err != nil {
		return rc.API().InternalError(err)
	}
	return rc.API().JSON(prices)
}

// Init inits the web app.
func Init() *web.App {
	app := web.New()
	app.SetName(AppName)
	app.SetPort(Config.Port())

	if Config.IsProduction() {
		app.SetLogger(web.NewStandardOutputErrorLogger())
	} else {
		app.SetLogger(web.NewStandardOutputLogger())
	}

	app.GET("/", rootHandler)
	app.GET("/stock/chart/:ticker", stockHandler)
	app.GET("/stock/chart/:ticker/:timeframe", stockHandler)
	app.GET("/api/v1/stock/prices/:ticker", apiStockPriceHistoricalHandler)
	app.GET("/api/v1/stock/prices/:ticker/:timeframe", apiStockPriceHistoricalHandler)
	return app
}
