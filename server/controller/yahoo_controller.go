package controller

import (
	"time"

	"github.com/wcharczuk/chart-service/server/core"
	"github.com/wcharczuk/chart-service/server/yahoo"
	"github.com/wcharczuk/go-web"
)

// Yahoo is the yahoo controller.
type Yahoo struct{}

func (y Yahoo) getQuoteAction(rc *web.RequestContext) web.ControllerResult {
	ticker, err := rc.RouteParameter("ticker")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}
	quote, err := yahoo.GetStockPrice([]string{ticker})
	if err != nil {
		return rc.API().InternalError(err)
	}
	return rc.API().JSON(quote)
}

func (y Yahoo) getPricesAction(rc *web.RequestContext) web.ControllerResult {
	ticker, err := rc.RouteParameter("ticker")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}

	from := time.Now().UTC().AddDate(0, -1, 0)
	to := time.Now().UTC()

	timeframe, err := rc.RouteParameter("timeframe")
	if err == nil {
		from, to, _, _, err = core.ParseTimeFrame(timeframe)
		if err != nil {
			return rc.API().BadRequest(err.Error())
		}
	}

	hist, err := yahoo.GetHistoricalPrices(ticker, from, to)
	if err != nil {
		return rc.API().InternalError(err)
	}

	return rc.API().JSON(hist)
}

// Register registers the controllers routes with the app.
func (y Yahoo) Register(app *web.App) {
	app.GET("/api/v1/yahoo.quote/:ticker", y.getQuoteAction)
	app.GET("/api/v1/yahoo.prices/:ticker", y.getPricesAction)
	app.GET("/api/v1/yahoo.prices/:ticker/:timeframe", y.getPricesAction)
}
