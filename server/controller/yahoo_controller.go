package controller

import (
	"time"

	"github.com/blendlabs/go-web"
	"github.com/wcharczuk/chart-service/server/core"
	"github.com/wcharczuk/chart-service/server/yahoo"
)

// Yahoo is the yahoo controller.
type Yahoo struct{}

func (y Yahoo) getQuoteAction(rc *web.Ctx) web.Result {
	ticker, err := rc.RouteParam("ticker")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}
	quote, err := yahoo.GetStockPrice([]string{ticker})
	if err != nil {
		return rc.API().InternalError(err)
	}
	return rc.API().Result(quote)
}

func (y Yahoo) getPricesAction(rc *web.Ctx) web.Result {
	ticker, err := rc.RouteParam("ticker")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}

	from := time.Now().UTC().AddDate(0, -1, 0)
	to := time.Now().UTC()

	timeframe, err := rc.RouteParam("timeframe")
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

	return rc.API().Result(hist)
}

// Register registers the controllers routes with the app.
func (y Yahoo) Register(app *web.App) {
	app.GET("/api/v1/yahoo.quote/:ticker", y.getQuoteAction)
	app.GET("/api/v1/yahoo.prices/:ticker", y.getPricesAction)
	app.GET("/api/v1/yahoo.prices/:ticker/:timeframe", y.getPricesAction)
}
