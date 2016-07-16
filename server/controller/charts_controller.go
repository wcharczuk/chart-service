package controller

import (
	"fmt"
	"net/http"
	"strings"
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
	stockTicker := strings.ToUpper(stock.Ticker)

	from := time.Now().UTC().AddDate(0, -1, 0)
	to := time.Now().UTC()

	width := defaultChartWidth
	height := defaultChartHeight
	showAxes := true
	showLastValue := true
	usePercentages := false
	useSimpleMovingAverage := false
	useExpMovingAverage := false
	useLegend := false
	useBollingerBands := false
	xvf := chart.TimeValueFormatter
	yvf := chart.FloatValueFormatter

	smaSize := 16
	emaSigma := 0.1818

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

	if useSimpleMovingAverageValue, err := rc.QueryParam("use_sma"); err == nil {
		useSimpleMovingAverage = util.CaseInsensitiveEquals(useSimpleMovingAverageValue, "true")
	}

	if useExpMovingAverageValue, err := rc.QueryParam("use_ema"); err == nil {
		useExpMovingAverage = util.CaseInsensitiveEquals(useExpMovingAverageValue, "true")
	}

	if useLegendValue, err := rc.QueryParam("use_legend"); err == nil {
		useLegend = util.CaseInsensitiveEquals(useLegendValue, "true")
	}

	if useBollingerBandsValue, err := rc.QueryParam("use_bbs"); err == nil {
		useBollingerBands = util.CaseInsensitiveEquals(useBollingerBandsValue, "true")
	}

	if smaSizeValue, err := rc.QueryParamInt("window"); err == nil {
		smaSize = smaSizeValue
	}

	if emaSigmaValue, err := rc.QueryParamFloat64("sigma"); err == nil {
		emaSigma = emaSigmaValue
	}

	fillColor := drawing.ColorTransparent
	if showAxes && !useBollingerBands {
		fillColor = chart.GetDefaultSeriesStrokeColor(0).WithAlpha(64)
	}

	timeframe, err := rc.RouteParameter("timeframe")
	if err == nil {
		from, to, xvf, yvf, err = core.ParseTimeFrame(timeframe)
		if err != nil {
			return rc.API().BadRequest(err.Error())
		}
	}

	if usePercentages {
		yvf = chart.PercentValueFormatter
	}

	equityPrices, err := viewmodel.GetEquityPricesByDate(stockTicker, from, to)
	if err != nil {
		return rc.API().InternalError(err)
	}
	var vx []time.Time
	var vy []float64
	if usePercentages {
		vx, vy = model.EquityPrices(equityPrices).PercentChange()
	} else {
		vx, vy = model.EquityPrices(equityPrices).Prices()
	}

	s1 := chart.TimeSeries{
		Name:    stockTicker,
		XValues: vx,
		YValues: vy,
		Style: chart.Style{
			Show:        true,
			StrokeColor: chart.GetDefaultSeriesStrokeColor(0),
			FillColor:   fillColor,
		},
	}

	lva := model.EquityPrices(equityPrices).LastValueAnnotation(util.TernaryOfString(useLegend, "", stockTicker), yvf)
	if usePercentages {
		lva = model.EquityPrices(equityPrices).LastValueAnnotationPercentChange(util.TernaryOfString(useLegend, "", stockTicker), yvf)
	}

	s1as := chart.AnnotationSeries{
		Name: fmt.Sprintf("%s - LV", stockTicker),
		Style: chart.Style{
			Show:        showLastValue,
			StrokeColor: chart.GetDefaultSeriesStrokeColor(0),
		},
		Annotations: []chart.Annotation{lva},
	}

	s1sma := &chart.SimpleMovingAverageSeries{
		Name: fmt.Sprintf("%s - Mov. Avg.", stockTicker),
		Style: chart.Style{
			Show:            useSimpleMovingAverage,
			StrokeColor:     drawing.ColorRed,
			StrokeDashArray: []float64{5, 5},
		},
		InnerSeries: s1,
		WindowSize:  smak,
	}
	smax, smay := s1sma.GetLastValue()
	lvssma := chart.Annotation{
		X:     smax,
		Y:     smay,
		Label: fmt.Sprintf("%s %s", util.TernaryOfString(useLegend, "", stockTicker), yvf(smay)),
	}
	s1smaas := chart.AnnotationSeries{
		Name: fmt.Sprintf("%s - Mov. Avg. LV", stockTicker),
		Style: chart.Style{
			Show:        showLastValue && useSimpleMovingAverage,
			StrokeColor: drawing.ColorRed,
		},
		Annotations: []chart.Annotation{lvssma},
	}

	s1ema := &chart.ExponentialMovingAverageSeries{
		Name: fmt.Sprintf("%s - Mov. Avg.", stockTicker),
		Style: chart.Style{
			Show:            useExpMovingAverage,
			StrokeColor:     drawing.ColorBlue,
			StrokeDashArray: []float64{5, 5},
		},
		InnerSeries: s1,
		Sigma:       emaSigma,
	}
	emax, emay := s1ema.GetLastValue()
	lvsema := chart.Annotation{
		X:     emax,
		Y:     emay,
		Label: fmt.Sprintf("%s %s", util.TernaryOfString(useLegend, "", stockTicker), yvf(emay)),
	}
	s1emaas := chart.AnnotationSeries{
		Name: fmt.Sprintf("%s - Mov. Avg. LV", stockTicker),
		Style: chart.Style{
			Show:        showLastValue && useExpMovingAverage,
			StrokeColor: drawing.ColorRed,
		},
		Annotations: []chart.Annotation{lvsema},
	}

	s1bbs := &chart.BollingerBandsSeries{
		Name: fmt.Sprintf("%s - Bol. Bands", stockTicker),
		Style: chart.Style{
			Show:        useBollingerBands,
			StrokeColor: chart.GetDefaultSeriesStrokeColor(0).WithAlpha(48),
			FillColor:   chart.GetDefaultSeriesStrokeColor(0).WithAlpha(32),
		},
		InnerSeries: s1,
		K:           2.0,
		WindowSize:  16,
	}

	graph := chart.Chart{
		Title: stock.Name,
		TitleStyle: chart.Style{
			Show: false,
		},
		Width:  width,
		Height: height,
		XAxis: chart.XAxis{
			ValueFormatter: xvf,
			Style: chart.Style{
				Show: showAxes,
			},
		},
		YAxis: chart.YAxis{
			ValueFormatter: yvf,
			Style: chart.Style{
				Show: showAxes,
			},
		},
		Series: []chart.Series{
			s1bbs,
			s1,
			s1as,
			s1sma,
			s1smaas,
			s1ema,
			s1emaas,
		},
	}
	if useLegend {
		graph.Elements = []chart.Renderable{
			chart.CreateLegend(&graph, chart.Style{
				FontSize: 8.0,
			}),
		}
	}

	if compareTicker, err := rc.QueryParam("compare"); err == nil {
		compareTicker = strings.ToUpper(compareTicker)
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
			ValueFormatter: yvf,
			Style: chart.Style{
				Show: showAxes && !usePercentages,
			},
		}

		compareFillColor := drawing.ColorTransparent
		if showAxes && !useBollingerBands {
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

		clva := model.EquityPrices(compareEquityPrices).LastValueAnnotation(util.TernaryOfString(useLegend, "", compareTicker), yvf)
		if usePercentages {
			clva = model.EquityPrices(compareEquityPrices).LastValueAnnotationPercentChange(util.TernaryOfString(useLegend, "", compareTicker), yvf)
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
