package controller

import (
	"github.com/wcharczuk/chart-service/server/model"
	"github.com/wcharczuk/go-web"
)

// EquityPrices is the equity prices controller.
type EquityPrices struct{}

func (ep EquityPrices) getPricesAction(rc *web.RequestContext) web.ControllerResult {
	ticker, err := rc.RouteParam("ticker")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}

	prices, err := model.GetEquityPrices(ticker)
	if err != nil {
		return rc.API().InternalError(err)
	}

	return rc.API().JSON(prices)
}

// Register registers the controllers routes with the app.
func (ep EquityPrices) Register(app *web.App) {
	app.GET("/api/v1/equity.prices/:ticker", ep.getPricesAction)
}
