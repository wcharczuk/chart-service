package viewmodel

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/blendlabs/go-util"
	"github.com/wcharczuk/chart-service/server/core"
	"github.com/wcharczuk/chart-service/server/model"
	"github.com/wcharczuk/chart-service/server/yahoo"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
	"github.com/wcharczuk/go-web"
)

const (
	defaultChartWidth     = 1024
	defaultChartHeight    = 400
	defaultChartTimeframe = "LTM"
)

// Chart are all the chart parameters.
type Chart struct {
	Width  int    `query:"width"`
	Height int    `query:"height"`
	Format string `query:"format"`

	ChartTimeframe     string `route:"period"`
	Start              time.Time
	End                time.Time
	ShouldUseDaySeries bool

	Ticker                   string `route:"ticker"`
	TickerInfo               *yahoo.StockInfo
	TickerCompare            string `query:"compare"`
	TickerCompareInfo        *yahoo.StockInfo
	UsePercentageDifferences bool `query:"format"`

	ShowAxes                    bool `query:"show_axes"`
	ShowLastValue               bool `query:"show_last"`
	ShowLegend                  bool `query:"show_legend"`
	AddSimpleMovingAverage      bool `query:"add_sma"`
	AddExponentialMovingAverage bool `query:"add_ema"`
	AddBollingerBands           bool `query:"add_bb"`
	AddMACD                     bool `query:"add_macd"`

	XValueFormatter chart.ValueFormatter
	YValueFormatter chart.ValueFormatter

	tickerData        []model.EquityPrice
	tickerCompareData []model.EquityPrice

	K        float64 `query:"k"`
	MAPeriod int     `query:"period"`
}

// Parse sets the chart properties from a request context.
func (c *Chart) Parse(rc *web.RequestContext) error {
	c.Width = core.ReadQueryValueInt(rc, "width", defaultChartWidth)
	c.Height = core.ReadQueryValueInt(rc, "height", defaultChartHeight)

	c.Format = core.ReadQueryValue(rc, "format", "png")

	c.Ticker = core.ReadRouteValue(rc, "ticker", "")
	c.TickerCompare = core.ReadQueryValue(rc, "compare", "")
	c.ChartTimeframe = core.ReadRouteValue(rc, "timeframe", defaultChartTimeframe)
	c.UsePercentageDifferences = core.ReadQueryValueBool(rc, "use_pct", false)

	c.ShowAxes = core.ReadQueryValueBool(rc, "show_axes", false)
	c.ShowLastValue = core.ReadQueryValueBool(rc, "show_last", false)
	c.ShowLegend = core.ReadQueryValueBool(rc, "show_legend", false)

	c.AddSimpleMovingAverage = core.ReadQueryValueBool(rc, "add_sma", false)
	c.AddExponentialMovingAverage = core.ReadQueryValueBool(rc, "add_ema", false)
	c.AddBollingerBands = core.ReadQueryValueBool(rc, "add_bb", false)
	c.AddMACD = core.ReadQueryValueBool(rc, "add_macd", false)

	c.K = core.ReadQueryValueFloat64(rc, "k", 2.0)
	c.MAPeriod = core.ReadQueryValueInt(rc, "period", 16)

	if c.UsePercentageDifferences {
		c.YValueFormatter = chart.PercentValueFormatter
	} else {
		c.YValueFormatter = chart.FloatValueFormatter
	}

	return nil
}

// ParsePeriod reads the chart period
func (c *Chart) ParsePeriod() error {
	switch strings.ToLower(c.ChartTimeframe) {
	case "ltm":
		c.Start = time.Now().UTC().AddDate(0, -12, 0)
		c.End = time.Now().UTC()
	case "6m":
		c.Start = time.Now().UTC().AddDate(0, -6, 0)
		c.End = time.Now().UTC()
	case "3m":
		c.Start = time.Now().UTC().AddDate(0, -3, 0)
		c.End = time.Now().UTC()
	case "2m":
		c.Start = time.Now().UTC().AddDate(0, -2, 0)
		c.End = time.Now().UTC()
	case "1m":
		c.Start = time.Now().UTC().AddDate(0, -1, 0)
		c.End = time.Now().UTC()
		c.ShouldUseDaySeries = true
	case "1wk":
		c.Start = time.Now().UTC().AddDate(0, 0, -7)
		c.End = time.Now().UTC()
		c.ShouldUseDaySeries = true
	case "10d":
		c.Start = time.Now().UTC().AddDate(0, 0, -10)
		c.End = time.Now().UTC()
		c.ShouldUseDaySeries = true
	case "3d":
		c.Start = time.Now().UTC().AddDate(0, 0, -3)
		c.End = time.Now().UTC()
		c.XValueFormatter = chart.TimeHourValueFormatter
		c.ShouldUseDaySeries = true
	case "1d":
		c.Start = time.Now().UTC().AddDate(0, 0, -1)
		c.End = time.Now().UTC()
		c.XValueFormatter = chart.TimeHourValueFormatter
		c.ShouldUseDaySeries = true
	default:
		return fmt.Errorf("Invalid chart period: %s", c.ChartTimeframe)
	}
	return nil
}

// Validate applies some sanity check validation rules.
func (c *Chart) Validate() error {
	if len(c.Ticker) == 0 {
		return errors.New("Caller did not specify a :ticker parameter, cannot continue.")
	}
	if c.AddMACD && (c.hasCompare() && !c.UsePercentageDifferences) {
		return errors.New("Cannot add both MACD histogram and use a secondary axis for comparison.")
	}
	if c.Start.IsZero() {
		return errors.New("Data period start time is unset, cannot continue.")
	}
	if c.End.IsZero() {
		return errors.New("Data period end time is unset, cannot continue.")
	}

	return nil
}

// FetchTickers fetches the ticker information.
func (c *Chart) FetchTickers() error {
	tickers := []string{c.Ticker}
	if c.hasCompare() {
		tickers = []string{c.Ticker, c.TickerCompare}
	}
	stockInfos, err := yahoo.GetStockPrice(tickers)
	if err != nil {
		return err
	}
	if len(stockInfos) == 0 {
		return fmt.Errorf("No stock info returned for: %#v", tickers)
	}

	if stockInfos[0].IsZero() {
		return fmt.Errorf("No stock information returned for: %s", strings.ToUpper(c.Ticker))
	}
	c.Ticker = strings.ToUpper(c.Ticker)
	c.TickerInfo = &stockInfos[0]

	if len(stockInfos) > 1 {
		if stockInfos[1].IsZero() {
			return fmt.Errorf("No stock information returned for: %s", strings.ToUpper(c.TickerCompare))
		}
		c.TickerCompare = strings.ToUpper(c.TickerCompare)
		c.TickerCompareInfo = &stockInfos[1]
	}
	return nil
}

// FetchPriceData fetches price data.
func (c *Chart) FetchPriceData() error {
	var useLocal, useRemote bool

	switch strings.ToLower(c.ChartTimeframe) {
	case "ltm", "6m", "3m":
		useLocal = false
		useRemote = true
	case "1m", "1wk", "10d", "3d", "1d":
		useLocal = true
		useRemote = false
	}

	data, err := GetEquityPricesByDate(c.Ticker, c.Start, c.End, useLocal, useRemote)
	if err != nil {
		return err
	}
	c.tickerData = data

	if c.hasCompare() {
		compareData, err := GetEquityPricesByDate(c.TickerCompare, c.Start, c.End, useLocal, useRemote)
		if err != nil {
			return err
		}
		c.tickerCompareData = compareData
	}

	return nil
}

// CreateChart creates a chart object for the parameters.
func (c *Chart) CreateChart() (chart.Chart, error) {
	graph := chart.Chart{
		Width:  c.Width,
		Height: c.Height,
		XAxis: chart.XAxis{
			ValueFormatter: c.XValueFormatter,
			Style: chart.Style{
				Show: c.ShowAxes,
			},
		},
		YAxis: chart.YAxis{
			ValueFormatter: c.YValueFormatter,
			Style: chart.Style{
				Show: c.ShowAxes,
			},
		},
		YAxisSecondary: chart.YAxis{
			ValueFormatter: c.YValueFormatter,
			Style: chart.Style{
				Show: c.showSecondaryAxis(),
			},
		},
		Series: c.getSeries(),
	}
	if c.ShowLegend {
		graph.Elements = []chart.Renderable{
			chart.CreateLegend(&graph, chart.Style{
				FontSize: 8.0,
			}),
		}
	}
	return graph, nil
}

func (c *Chart) getSeries() []chart.Series {
	t0series := c.getPriceSeries(c.Ticker, c.tickerData)
	series := []chart.Series{}

	if c.AddBollingerBands {
		bbs := c.getBBSeries(c.Ticker, c.tickerData)
		series = append(series, bbs)

		if c.ShowLastValue {
			series = append(series, c.getBoundedLastValueSeries(c.Ticker, bbs))
		}
	}

	series = append(series, t0series)
	if c.ShowLastValue {
		series = append(series, c.getLastValueSeries(c.Ticker, t0series))
	}

	if c.hasCompare() {
		t1series := c.getPriceSeries(c.TickerCompare, c.tickerCompareData)
		series = append(series, t1series)
		if c.ShowLastValue {
			series = append(series, c.getLastValueSeries(c.TickerCompare, t1series))
		}
	}

	if c.AddSimpleMovingAverage {
		sma := c.getSMASeries(c.Ticker, t0series)
		series = append(series, sma)
		if c.ShowLastValue {
			series = append(series, c.getLastValueSeries(c.Ticker, sma))
		}
	}

	if c.AddExponentialMovingAverage {
		ema := c.getEMASeries(c.Ticker, t0series)
		series = append(series, ema)
		if c.ShowLastValue {
			series = append(series, c.getLastValueSeries(c.Ticker, ema))
		}
	}

	if c.AddMACD {
		series = append(series, c.getMACDHistogramSeries(c.Ticker, c.tickerData))
		series = append(series, c.getMACDSignalSeries(c.Ticker, c.tickerData))
		series = append(series, c.getMACDLineSeries(c.Ticker, c.tickerData))
	}

	return series
}

func (c *Chart) getPriceSeries(ticker string, data []model.EquityPrice) chart.TimeSeries {
	var xvalues []time.Time
	var yvalues []float64
	yaxis := chart.YAxisPrimary

	if c.UsePercentageDifferences {
		xvalues, yvalues = model.EquityPrices(data).PercentChange()
	} else {
		xvalues, yvalues = model.EquityPrices(data).Prices()
	}
	index := 0
	if util.CaseInsensitiveEquals(ticker, c.TickerCompare) {
		index = 1
		if !c.UsePercentageDifferences {
			yaxis = chart.YAxisSecondary
		}
	}
	stroke, fill := c.getPriceSeriesColors(index)
	return chart.TimeSeries{
		Name:  ticker,
		YAxis: yaxis,
		Style: chart.Style{
			Show:        true,
			StrokeColor: stroke,
			FillColor:   fill,
		},
		XValues: xvalues,
		YValues: yvalues,
	}
}

func (c *Chart) getLastValueSeries(ticker string, priceSeries chart.FullValueProvider) chart.Series {
	lvx, lvy := priceSeries.GetLastValue()

	if c.UsePercentageDifferences {
		_, v0y := priceSeries.GetValue(0)
		if v0y > 0 {
			lvy = chart.PercentDifference(v0y, lvy)
		}
	}

	yaxis := chart.YAxisPrimary
	if util.CaseInsensitiveEquals(ticker, c.TickerCompare) {
		if !c.UsePercentageDifferences {
			yaxis = chart.YAxisSecondary
		}
	}

	var style chart.Style
	if typed, isSeries := priceSeries.(chart.Series); isSeries {
		style = typed.GetStyle()
	}
	style.Show = c.ShowLastValue
	style.FillColor = drawing.ColorWhite

	labelText := c.YValueFormatter(lvy)
	if !c.ShowLegend {
		labelText = ticker + " " + labelText
	}

	return chart.AnnotationSeries{
		Name:  fmt.Sprintf("%s - Last Value", ticker),
		YAxis: yaxis,
		Style: style,
		Annotations: []chart.Annotation{
			{X: lvx, Y: lvy, Label: labelText},
		},
	}
}

func (c *Chart) getBoundedLastValueSeries(ticker string, priceSeries chart.FullBoundedValueProvider) chart.Series {
	lvx, lvy1, lvy2 := priceSeries.GetBoundedLastValue()

	if c.UsePercentageDifferences {
		_, v0y1, v0y2 := priceSeries.GetBoundedValue(0)
		lvy1 = chart.PercentDifference(v0y1, lvy1)
		lvy2 = chart.PercentDifference(v0y2, lvy2)
	}

	var style chart.Style
	if s, isSeries := priceSeries.(chart.Series); isSeries {
		style = s.GetStyle()
	}
	style.Show = c.ShowLastValue
	style.FillColor = drawing.ColorWhite

	label1 := fmt.Sprintf("%s +%0.0fσ %s", ticker, c.K, c.YValueFormatter(lvy1))
	if c.ShowLegend {
		label1 = fmt.Sprintf("+%0.0fσ %s", c.K, c.YValueFormatter(lvy1))
	}
	label2 := fmt.Sprintf("%s -%0.0fσ %s", ticker, c.K, c.YValueFormatter(lvy2))
	if c.ShowLegend {
		label2 = fmt.Sprintf("-%0.0fσ %s", c.K, c.YValueFormatter(lvy2))
	}

	return chart.AnnotationSeries{
		Name:  fmt.Sprintf("%s - Last Value", ticker),
		Style: style,
		Annotations: []chart.Annotation{
			{X: lvx, Y: lvy1, Label: label1},
			{X: lvx, Y: lvy2, Label: label2},
		},
	}
}

func (c *Chart) getSMASeries(ticker string, priceSeries chart.ValueProvider) *chart.SMASeries {
	return &chart.SMASeries{
		Name: fmt.Sprintf("%s SMA", ticker),
		Style: chart.Style{
			Show:            c.AddSimpleMovingAverage,
			StrokeColor:     drawing.ColorRed,
			StrokeDashArray: []float64{5.0, 5.0},
		},
		InnerSeries: priceSeries,
		Period:      c.MAPeriod,
	}
}

func (c *Chart) getEMASeries(ticker string, priceSeries chart.ValueProvider) *chart.EMASeries {
	return &chart.EMASeries{
		Name: fmt.Sprintf("%s EMA", ticker),
		Style: chart.Style{
			Show:            c.AddExponentialMovingAverage,
			StrokeColor:     drawing.ColorBlue,
			StrokeDashArray: []float64{5.0, 5.0},
		},
		InnerSeries: priceSeries,
		Period:      c.MAPeriod,
	}
}

func (c *Chart) getBBSeries(ticker string, data []model.EquityPrice) *chart.BollingerBandsSeries {
	return &chart.BollingerBandsSeries{
		Name: fmt.Sprintf("%s Bol. Bands", ticker),
		Style: chart.Style{
			Show:        c.AddBollingerBands,
			StrokeColor: drawing.ColorFromHex("efefef"),
			FillColor:   drawing.ColorFromHex("efefef").WithAlpha(100),
		},
		InnerSeries: c.getPriceSeries(ticker, data),
		Period:      c.MAPeriod,
	}
}

func (c *Chart) getMACDHistogramSeries(ticker string, data []model.EquityPrice) chart.HistogramSeries {
	return chart.HistogramSeries{
		Name: fmt.Sprintf("%s - MACD Div.", ticker),
		Style: chart.Style{
			Show:        c.showMACD(),
			StrokeColor: drawing.ColorGreen,
			FillColor:   drawing.ColorGreen,
		},
		YAxis: chart.YAxisSecondary,
		InnerSeries: &chart.MACDSeries{
			InnerSeries: c.getPriceSeries(ticker, data),
		},
	}
}

func (c *Chart) getMACDSignalSeries(ticker string, data []model.EquityPrice) *chart.MACDSignalSeries {
	return &chart.MACDSignalSeries{
		Name: fmt.Sprintf("%s - MACD EMA", ticker),
		Style: chart.Style{
			Show:        c.showMACD(),
			StrokeColor: drawing.ColorRed,
		},
		YAxis:       chart.YAxisSecondary,
		InnerSeries: c.getPriceSeries(ticker, data),
	}
}

func (c *Chart) getMACDLineSeries(ticker string, data []model.EquityPrice) *chart.MACDLineSeries {
	return &chart.MACDLineSeries{
		Name: fmt.Sprintf("%s - MACD", ticker),
		Style: chart.Style{
			Show:        c.showMACD(),
			StrokeColor: drawing.ColorBlue,
		},
		YAxis:       chart.YAxisSecondary,
		InnerSeries: c.getPriceSeries(ticker, data),
	}
}

func (c *Chart) showMACD() bool {
	return c.AddMACD && !(c.hasCompare() && !c.UsePercentageDifferences)
}

func (c *Chart) hasCompare() bool {
	return len(c.TickerCompare) > 0
}

func (c *Chart) showSecondaryAxis() bool {
	return c.ShowAxes && !c.UsePercentageDifferences && (c.hasCompare() || c.showMACD())
}

func (c *Chart) getPriceSeriesColors(index int) (stroke, fill drawing.Color) {
	stroke = chart.GetDefaultSeriesStrokeColor(index)
	if !c.AddBollingerBands {
		fill = stroke.WithAlpha(64)
	}
	return
}
