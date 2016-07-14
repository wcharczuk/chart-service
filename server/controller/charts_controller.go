package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/blendlabs/go-util"
	"github.com/wcharczuk/chart-service/server/core"
	"github.com/wcharczuk/chart-service/server/model"
	"github.com/wcharczuk/chart-service/server/viewmodel"
	"github.com/wcharczuk/chart-service/server/yahoo"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
	"github.com/wcharczuk/go-web"
)

const (
	defaultChartWidth  = 1024
	defaultChartHeight = 400
)

// Charts is the controller that generates charts.
type Charts struct{}

func (cc Charts) getChartAction(rc *web.RequestContext) web.ControllerResult {
	ticker, err := rc.RouteParameter("ticker")
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}

	stockInfos, err := yahoo.GetStockPrice([]string{ticker})
	if err != nil {
		return rc.API().InternalError(err)
	}
	if len(stockInfos) == 0 || stockInfos[0].IsZero() {
		return rc.API().BadRequest(fmt.Sprintf("Ticker `%s` could not be found.", ticker))
	}

	stock := stockInfos[0]

	from := time.Now().UTC().AddDate(0, -1, 0)
	to := time.Now().UTC()

	timeframe, err := rc.RouteParameter("timeframe")
	if err == nil {
		from, to, err = core.ParseTimeFrame(timeframe)
		if err != nil {
			return rc.API().BadRequest(err.Error())
		}
	}

	equityPrices, err := viewmodel.GetEquityPricesByDate(stock.Ticker, from, to)
	if err != nil {
		return rc.API().InternalError(err)
	}

	width := defaultChartWidth
	height := defaultChartHeight
	showAxes := true
	showLastValue := true
	usePercentages := false

	if widthValue, err := rc.QueryParamInt("width"); err == nil {
		width = widthValue
	}

	if heightValue, err := rc.QueryParamInt("height"); err == nil {
		height = heightValue
	}

	if showAxesValue, err := rc.QueryParam("show_axes"); err == nil {
		showAxes = util.CaseInsensitiveEquals(showAxesValue, "true")
	}

	if showLastValueValue, err := rc.QueryParam("show_last"); err == nil {
		showLastValue = util.CaseInsensitiveEquals(showLastValueValue, "true")
	}

	if usePercentagesValue, err := rc.QueryParam("use_pct"); err == nil {
		usePercentages = util.CaseInsensitiveEquals(usePercentagesValue, "true")
	}

	fillColor := drawing.ColorTransparent
	if showAxes {
		fillColor = chart.GetDefaultSeriesStrokeColor(0).WithAlpha(64)
	}

	var vx []time.Time
	var vy []float64
	if usePercentages {
		vx, vy = model.EquityPrices(equityPrices).PercentChange()
	} else {
		vx, vy = model.EquityPrices(equityPrices).Prices()
	}

	vf := chart.FloatValueFormatter
	if usePercentages {
		vf = func(v interface{}) string {
			return fmt.Sprintf("%0.2f%%", v)
		}
	}

	lva := model.EquityPrices(equityPrices).LastValueAnnotation(stock.Ticker, vf)
	if usePercentages {
		lva = model.EquityPrices(equityPrices).LastValueAnnotationPercentChange(stock.Ticker, vf)
	}

	graph := chart.Chart{
		Title: stock.Name,
		TitleStyle: chart.Style{
			Show: false,
		},
		Width:  width,
		Height: height,
		XAxis: chart.XAxis{
			Style: chart.Style{
				Show: showAxes,
			},
		},
		YAxis: chart.YAxis{
			ValueFormatter: vf,
			Style: chart.Style{
				Show: showAxes,
			},
		},
		Series: []chart.Series{
			chart.TimeSeries{
				Name:    stock.Ticker,
				XValues: vx,
				YValues: vy,
				Style: chart.Style{
					Show:        true,
					StrokeColor: chart.GetDefaultSeriesStrokeColor(0),
					FillColor:   fillColor,
				},
			},
			chart.AnnotationSeries{
				Name: fmt.Sprintf("%s - Last Value", stock.Ticker),
				Style: chart.Style{
					Show:        showLastValue,
					StrokeColor: chart.GetDefaultSeriesStrokeColor(0),
				},
				Annotations: []chart.Annotation{lva},
			},
		},
	}

	if compareTicker, err := rc.QueryParam("compare"); err == nil {
		compareEquityPrices, err := viewmodel.GetEquityPricesByDate(compareTicker, from, to)
		if err != nil {
			return rc.API().InternalError(err)
		}

		var cx []time.Time
		var cy []float64
		if usePercentages {
			cx, cy = model.EquityPrices(compareEquityPrices).PercentChange()
		} else {
			cx, cy = model.EquityPrices(compareEquityPrices).Prices()
		}

		graph.YAxisSecondary = chart.YAxis{
			ValueFormatter: vf,
			Style: chart.Style{
				Show: showAxes && !usePercentages,
			},
		}

		compareFillColor := drawing.ColorTransparent
		if showAxes {
			compareFillColor = chart.GetDefaultSeriesStrokeColor(1).WithAlpha(64)
		}

		yaxis := chart.YAxisPrimary
		if !usePercentages {
			yaxis = chart.YAxisSecondary
		}

		graph.Series = append([]chart.Series{chart.TimeSeries{
			Name:    compareTicker,
			XValues: cx,
			YValues: cy,
			YAxis:   yaxis,
			Style: chart.Style{
				Show:        true,
				StrokeColor: chart.GetDefaultSeriesStrokeColor(1),
				FillColor:   compareFillColor,
			},
		}}, graph.Series...)

		clva := model.EquityPrices(compareEquityPrices).LastValueAnnotation(compareTicker, vf)
		if usePercentages {
			clva = model.EquityPrices(compareEquityPrices).LastValueAnnotationPercentChange(compareTicker, vf)
		}

		graph.Series = append(graph.Series, chart.AnnotationSeries{
			Name: fmt.Sprintf("%s - Last Value", compareTicker),
			Style: chart.Style{
				Show:        showLastValue,
				StrokeColor: chart.GetDefaultSeriesStrokeColor(1),
			},
			YAxis:       yaxis,
			Annotations: []chart.Annotation{clva},
		})
	}

	format := "png"
	if formatValue, err := rc.QueryParam("format"); err == nil {
		format = formatValue
	}
	switch format {
	case "svg":
		rc.Response.Header().Set("Content-Type", "image/svg+xml")
		err = graph.Render(chart.SVG, rc.Response)
	case "png":
		rc.Response.Header().Set("Content-Type", "image/png")
		err = graph.Render(chart.PNG, rc.Response)
	default:
		return rc.API().BadRequest("invalid format type")
	}
	if err != nil {
		return rc.API().InternalError(err)
	}
	rc.Response.WriteHeader(http.StatusOK)
	return nil
}

// Register registers the controller.
func (cc Charts) Register(app *web.App) {
	app.GET("/stock/chart/:ticker", cc.getChartAction)
	app.GET("/stock/chart/:ticker/:timeframe", cc.getChartAction)
}
