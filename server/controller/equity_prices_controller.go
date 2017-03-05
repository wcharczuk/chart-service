package controller

import (
	"github.com/blendlabs/go-web"
	"github.com/wcharczuk/chart-service/server/model"
)

// EquityPrices is the equity prices controller.
type EquityPrices struct{}

func (ep EquityPrices) getPricesAction(rc *web.Ctx) web.Result {
	ticker, err := rc.RouteParam("ticker")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}

	prices, err := model.GetEquityPrices(ticker)
	if err != nil {
		return rc.API().InternalError(err)
	}

	return rc.API().Result(prices)
}

// Register registers the controllers routes with the app.
func (ep EquityPrices) Register(app *web.App) {
	app.GET("/api/v1/equity.prices/:ticker", ep.getPricesAction)
}
