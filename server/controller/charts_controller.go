package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/blendlabs/go-util"
	"github.com/wcharczuk/chart-service/server/core"
	"github.com/wcharczuk/chart-service/server/yahoo"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
	"github.com/wcharczuk/go-web"
)

// ChartsController is the controller that generates charts.
type ChartsController struct{}

func (cc ChartsController) getChartAction(rc *web.RequestContext) web.ControllerResult {
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

	from := time.Now().UTC().AddDate(0, -1, 0)
	to := time.Now().UTC()

	timeframe, err := rc.RouteParameter("timeframe")
	if err == nil {
		from, to, err = core.ParseTimeFrame(timeframe)
		if err != nil {
			return rc.API().BadRequest(err.Error())
		}
	}

	prices, err := yahoo.GetHistoricalPrices(stock.Ticker, from, to)
	if err != nil {
		return rc.API().InternalError(err)
	}

	width := 1024
	height := 400
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

	vx, vy := marshalPrices(prices, usePercentages)

	vf := chart.FloatValueFormatter
	if usePercentages {
		vf = func(v interface{}) string {
			return fmt.Sprintf("%0.2f%%", v)
		}
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
				Annotations: []chart.Annotation{
					chart.Annotation{
						X:     float64(vx[len(vx)-1].Unix()),
						Y:     vy[len(vy)-1],
						Label: fmt.Sprintf("%s %s", stock.Ticker, vf(vy[len(vy)-1])),
					},
				},
			},
		},
	}

	if compareTicker, err := rc.QueryParam("compare"); err == nil {
		comparePrices, err := yahoo.GetHistoricalPrices(compareTicker, from, to)
		if err != nil {
			return rc.API().InternalError(err)
		}
		cx, cy := marshalPrices(comparePrices, usePercentages)

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

		graph.Series = append(graph.Series, chart.AnnotationSeries{
			Name: fmt.Sprintf("%s - Last Value", compareTicker),
			Style: chart.Style{
				Show:        showLastValue,
				StrokeColor: chart.GetDefaultSeriesStrokeColor(1),
			},
			YAxis: yaxis,
			Annotations: []chart.Annotation{
				chart.Annotation{
					X:     float64(cx[len(cx)-1].Unix()),
					Y:     cy[len(cy)-1],
					Label: fmt.Sprintf("%s %s", compareTicker, vf(cy[len(cy)-1])),
				},
			},
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
