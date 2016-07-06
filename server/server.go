package server

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"time"

	"github.com/ajstarks/svgo"
	"github.com/llgcode/draw2d/draw2dimg"
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

	width := 600
	height := 200
	padding := 10

	if widthValue, err := rc.QueryParamInt("width"); err == nil {
		width = widthValue
	}

	if heightValue, err := rc.QueryParamInt("height"); err == nil {
		height = heightValue
	}

	if paddingValue, err := rc.QueryParamInt("padding"); err == nil {
		padding = paddingValue
	}

	format := "svg"
	if formatValue, err := rc.QueryParam("format"); err == nil {
		format = formatValue
	}
	switch format {
	case "svg":
		return drawSVG(rc, prices, width, height, padding)
	case "png":
		return drawPNG(rc, prices, width, height, padding)
	default:
		return rc.API().BadRequest("invalid format type")
	}
}

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

func drawPNG(rc *web.RequestContext, prices []yahoo.HistoricalPrice, width, height, padding int) web.ControllerResult {
	rc.Response.Header().Set("Content-Type", "image/png")

	dest := image.NewRGBA(image.Rect(0, 0, width, height))
	gc := draw2dimg.NewGraphicContext(dest)

	effectiveWidth := width - padding<<1
	effectiveHeight := height - padding<<1

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
	gc.SetFillColor(color.RGBA{R: 0, G: 116, B: 217, A: 255})
	gc.SetLineWidth(3.0)
	gc.MoveTo(float64(x[0]), float64(y[0]))
	for index := 0; index < len(prices); index++ {
		gc.LineTo(float64(x[index]), float64(y[index]))
	}
	gc.Close()
	gc.FillStroke()
	buffer := bytes.NewBuffer([]byte{})
	png.Encode(buffer, dest)
	return rc.Raw(buffer.Bytes())
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
