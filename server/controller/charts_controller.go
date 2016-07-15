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

	width := defaultChartWidth
	height := defaultChartHeight
	showAxes := true
	showLastValue := true
	usePercentages := false
	useMovingAverages := false
	useLegend := false
	xvf := chart.TimeValueFormatter
	yvf := chart.FloatValueFormatter

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

	if useMovingAveragesValue, err := rc.QueryParam("use_ma"); err == nil {
		useMovingAverages = util.CaseInsensitiveEquals(useMovingAveragesValue, "true")
	}

	if useLegendValue, err := rc.QueryParam("use_legend"); err == nil {
		useLegend = util.CaseInsensitiveEquals(useLegendValue, "true")
	}

	fillColor := drawing.ColorTransparent
	if showAxes {
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

	equityPrices, err := viewmodel.GetEquityPricesByDate(stock.Ticker, from, to)
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
		Name:    stock.Ticker,
		XValues: vx,
		YValues: vy,
		Style: chart.Style{
			Show:        true,
			StrokeColor: chart.GetDefaultSeriesStrokeColor(0),
			FillColor:   fillColor,
		},
	}

	s1ma := &chart.MovingAverageSeries{
		Name: fmt.Sprintf("%s - Mov. Avg.", stock.Ticker),
		Style: chart.Style{
			Show:            useMovingAverages,
			StrokeColor:     drawing.ColorRed,
			StrokeDashArray: []float64{5, 5},
		},
		InnerSeries: s1,
		WindowSize:  16,
	}

	lva := model.EquityPrices(equityPrices).LastValueAnnotation(util.TernaryOfString(useLegend, "", stock.Ticker), yvf)
	if usePercentages {
		lva = model.EquityPrices(equityPrices).LastValueAnnotationPercentChange(util.TernaryOfString(useLegend, "", stock.Ticker), yvf)
	}

	s1as := chart.AnnotationSeries{
		Name: fmt.Sprintf("%s - LV", stock.Ticker),
		Style: chart.Style{
			Show:        showLastValue,
			StrokeColor: chart.GetDefaultSeriesStrokeColor(0),
		},
		Annotations: []chart.Annotation{lva},
	}

	malvx, malvy := s1ma.GetLastValue()
	lvma := chart.Annotation{
		X:     malvx,
		Y:     malvy,
		Label: fmt.Sprintf("%s %s", util.TernaryOfString(useLegend, "", stock.Ticker), yvf(malvy)),
	}

	s1maas := chart.AnnotationSeries{
		Name: fmt.Sprintf("%s - Mov. Avg. LV", stock.Ticker),
		Style: chart.Style{
			Show:        showLastValue && useMovingAverages,
			StrokeColor: drawing.ColorRed,
		},
		Annotations: []chart.Annotation{lvma},
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
			s1,
			s1ma,
			s1as,
			s1maas,
		},
	}
	if useLegend {
		graph.Elements = []chart.Renderable{
			createLegend(&graph),
		}
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
			ValueFormatter: yvf,
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

func createLegend(c *chart.Chart) chart.Renderable {
	return func(r chart.Renderer, cb chart.Box, defaults chart.Style) {

		// DEFAULTS
		legendPadding := 5
		lineTextGap := 5
		lineLengthMinimum := 25

		var labels []string
		var lines []chart.Style
		for _, s := range c.Series {
			if s.GetStyle().IsZero() || s.GetStyle().Show {
				if _, isAnnotationSeries := s.(chart.AnnotationSeries); !isAnnotationSeries {
					labels = append(labels, s.GetName())
					lines = append(lines, s.GetStyle())
				}
			}
		}

		legend := chart.Box{
			Top:  cb.Top, //padding
			Left: cb.Left,
		}

		legendContent := chart.Box{
			Top:  legend.Top + legendPadding,
			Left: legend.Left + legendPadding,
		}

		r.SetFontColor(chart.DefaultTextColor)
		r.SetFontSize(8.0)

		// measure
		for x := 0; x < len(labels); x++ {
			if len(labels[x]) > 0 {
				tb := r.MeasureText(labels[x])
				legendContent.Bottom += (tb.Height() + chart.DefaultMinimumTickVerticalSpacing)
				rowRight := tb.Width() + legendContent.Left + lineLengthMinimum + lineTextGap
				legendContent.Right = chart.MaxInt(legendContent.Right, rowRight)
			}
		}

		legend = legend.Grow(legendContent)
		chart.DrawBox(r, legend, chart.Style{
			FillColor:   drawing.ColorWhite,
			StrokeColor: chart.DefaultAxisColor,
			StrokeWidth: 1.0,
		})

		legendContent.Right = legend.Right - legendPadding
		legendContent.Bottom = legend.Bottom - legendPadding

		ycursor := legendContent.Top
		tx := legendContent.Left
		for x := 0; x < len(labels); x++ {
			if len(labels[x]) > 0 {
				tb := r.MeasureText(labels[x])
				ycursor += tb.Height()

				//r.SetFillColor(chart.DefaultTextColor)
				r.Text(labels[x], tx, ycursor)
				th2 := tb.Height() >> 1

				lx := tx + tb.Width() + lineTextGap
				ly := ycursor - th2
				lx2 := legendContent.Right - legendPadding

				r.SetStrokeColor(lines[x].GetStrokeColor())
				r.SetStrokeWidth(lines[x].GetStrokeWidth())
				r.SetStrokeDashArray(lines[x].GetStrokeDashArray())

				r.MoveTo(lx, ly)
				r.LineTo(lx2, ly)
				r.Stroke()

				ycursor += chart.DefaultMinimumTickVerticalSpacing
			}
		}
	}
}

// Register registers the controller.
func (cc Charts) Register(app *web.App) {
	app.GET("/stock/chart/:ticker", cc.getChartAction)
	app.GET("/stock/chart/:ticker/:timeframe", cc.getChartAction)
}
