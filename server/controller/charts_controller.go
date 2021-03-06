package controller

import (
	"github.com/blendlabs/go-util"
	"github.com/blendlabs/go-web"
	"github.com/wcharczuk/chart-service/server/viewmodel"
	"github.com/wcharczuk/go-chart"
)

// Charts is the controller that generates charts.
type Charts struct{}

func (cc Charts) getChartAction(rc *web.Ctx) web.Result {
	cv := &viewmodel.Chart{}
	err := cv.Parse(rc)
	if err != nil {
		return rc.API().InternalError(err)
	}
	err = cv.ParsePeriod()
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}
	err = cv.Validate()
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}
	err = cv.FetchTickers()
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}
	err = cv.FetchPriceData()
	if err != nil {
		return rc.API().InternalError(err)
	}

	graph, err := cv.CreateChart()
	if err != nil {
		return rc.API().InternalError(err)
	}

	if util.String.CaseInsensitiveEquals(cv.Format, "png") {
		rc.Response.Header().Set("Content-Type", "image/png")
		err := graph.Render(chart.PNG, rc.Response)
		if err != nil {
			if rc.Logger() != nil {
				rc.Logger().Errorf("render error: %s", err.Error())
			}
		}
	} else if util.String.CaseInsensitiveEquals(cv.Format, "svg") {
		rc.Response.Header().Set("Content-Type", "image/svg+xml")
		err := graph.Render(chart.SVG, rc.Response)
		if err != nil {
			if rc.Logger() != nil {
				rc.Logger().Errorf("render error: %s", err.Error())
			}
		}
	}

	return nil
}

// Register registers the controller.
func (cc Charts) Register(app *web.App) {
	app.GET("/stock/chart/:ticker", cc.getChartAction)
	app.GET("/stock/chart/:ticker/:timeframe", cc.getChartAction)
}
