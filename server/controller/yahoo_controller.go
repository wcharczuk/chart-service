package controller

import (
	"time"

	"github.com/blendlabs/go-web"
	"github.com/wcharczuk/chart-service/server/core"
	"github.com/wcharczuk/chart-service/server/google"
)

// Provider is the pricing provider controller.
type Provider struct{}

// Register registers the controllers routes with the app.
func (p Provider) Register(app *web.App) {
	app.GET("/api/v1/quote/:ticker", p.getQuoteAction)
	app.GET("/api/v1/prices/:ticker", p.getPricesAction)
	app.GET("/api/v1/prices/:ticker/:timeframe", p.getPricesAction)
}

func (p Provider) getQuoteAction(rc *web.Ctx) web.Result {
	ticker, err := rc.RouteParam("ticker")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}
	quote, err := google.GetCurrentPrices([]string{ticker})
	if err != nil {
		return rc.API().InternalError(err)
	}
	return rc.API().Result(quote)
}

func (p Provider) getPricesAction(rc *web.Ctx) web.Result {
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

	hist, err := google.GetHistoricalPrices(ticker, from, to)
	if err != nil {
		return rc.API().InternalError(err)
	}

	return rc.API().Result(hist)
}
