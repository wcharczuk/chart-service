package server

import (
	"bytes"
	"time"

	"github.com/wcharczuk/chart-service/server/core"
	"github.com/wcharczuk/chart-service/server/yahoo"
	"github.com/wcharczuk/go-chart"
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

	if widthValue, err := rc.QueryParamInt("width"); err == nil {
		width = widthValue
	}

	if heightValue, err := rc.QueryParamInt("height"); err == nil {
		height = heightValue
	}

	xvalues := make([]time.Time, len(prices))
	yvalues := make([]float64, len(prices))
	x := 0
	for index := len(prices) - 1; index >= 0; index-- {
		day := prices[index]
		xvalues[x] = day.Date
		yvalues[x] = day.Close
		x++
	}

	buffer := bytes.NewBuffer([]byte{})
	graph := chart.Chart{
		Title: stock.Name,
		TitleStyle: chart.Style{
			Show: false,
		},
		Width:  width,
		Height: height,
		Background: chart.Style{
			Padding: chart.Box{
				Right:  60.0,
				Bottom: 15.0,
			},
		},
		Axes: chart.Style{
			Show:        false,
			StrokeWidth: 1.0,
		},
		FinalValueLabel: chart.Style{
			Show: true,
		},
		Series: []chart.Series{
			chart.TimeSeries{
				Name: stock.Ticker,
				Style: chart.Style{
					StrokeWidth: 1.0,
				},
				XValues: xvalues,
				YValues: yvalues,
			},
		},
	}

	format := "png"
	if formatValue, err := rc.QueryParam("format"); err == nil {
		format = formatValue
	}
	switch format {
	case "svg":
		return rc.API().BadRequest("svg is not implemented yet")
	case "png":
		rc.Response.Header().Set("Content-Type", "image/png")
		graph.Render(chart.PNG, buffer)
	default:
		return rc.API().BadRequest("invalid format type")
	}
	return rc.Raw(buffer.Bytes())
}

/*
func drawSVG(rc *web.RequestContext, prices []yahoo.HistoricalPrice, width, height, padding int) web.ControllerResult {
	effectiveWidth := width - padding<<1
	effectiveHeight := height - padding<<1

	rc.Response.Header().Set("Content-Type", "image/svg+xml")

	buffer := bytes.NewBuffer([]byte{})
	canvas := svg.New(buffer)
	canvas.Start(width, height)

	var xvalues []time.Time
	var yvalues []float64

	for _, day := range prices {
		xvalues = append(xvalues, day.Date)
		yvalues = append(yvalues, day.Close)
	}

	xRange := core.NewRangeOfTime(effectiveWidth, padding, xvalues...)
	yRange := core.NewRange(effectiveHeight, padding, yvalues...)

	var x []int
	var y []int
	for _, day := range prices {
		x = append(x, xRange.Translate(day.Date))
		y = append(y, yRange.Translate(day.Close))
	}
	canvas.Polyline(x, y, "fill:none;stroke:#0074d9;stroke-width:3")
	canvas.End()

	return rc.Raw(buffer.Bytes())
}
*/

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
