package chart

import (
	"math"
	"sort"
)

// XAxis represents the horizontal axis.
type XAxis struct {
	Name           string
	Style          Style
	ValueFormatter ValueFormatter
	Range          Range
	Ticks          []Tick
}

// GetName returns the name.
func (xa XAxis) GetName() string {
	return xa.Name
}

// GetStyle returns the style.
func (xa XAxis) GetStyle() Style {
	return xa.Style
}

// GetTicks returns the ticks for a series. It coalesces between user provided ticks and
// generated ticks.
func (xa XAxis) GetTicks(r Renderer, ra Range, vf ValueFormatter) []Tick {
	if len(xa.Ticks) > 0 {
		return xa.Ticks
	}
	return xa.generateTicks(r, ra, vf)
}

func (xa XAxis) generateTicks(r Renderer, ra Range, vf ValueFormatter) []Tick {
	step := xa.getTickStep(r, ra, vf)
	return GenerateTicksWithStep(ra, step, vf)
}

func (xa XAxis) getTickStep(r Renderer, ra Range, vf ValueFormatter) float64 {
	tickCount := xa.getTickCount(r, ra, vf)
	step := ra.Delta() / float64(tickCount)
	return step
}

func (xa XAxis) getTickCount(r Renderer, ra Range, vf ValueFormatter) int {
	fontSize := xa.Style.GetFontSize(DefaultFontSize)
	r.SetFontSize(fontSize)

	// take a cut at determining the 'widest' value.
	l0 := vf(ra.Min)
	ln := vf(ra.Max)
	ll := l0
	if len(ln) > len(l0) {
		ll = ln
	}
	llb := r.MeasureText(ll)
	textWidth := llb.Width
	width := textWidth + DefaultMinimumTickHorizontalSpacing
	count := int(math.Ceil(float64(ra.Domain) / float64(width)))
	return count
}

// Measure returns the bounds of the axis.
func (xa XAxis) Measure(r Renderer, canvasBox Box, ra Range, ticks []Tick) Box {
	defaultFont, _ := GetDefaultFont()
	r.SetFont(xa.Style.GetFont(defaultFont))
	r.SetFontSize(xa.Style.GetFontSize(DefaultFontSize))

	sort.Sort(Ticks(ticks))

	var left, right, top, bottom = math.MaxInt32, 0, math.MaxInt32, 0
	for _, t := range ticks {
		v := t.Value
		lx := ra.Translate(v)
		tb := r.MeasureText(t.Label)

		tx := canvasBox.Left + lx
		ty := canvasBox.Bottom + DefaultXAxisMargin + tb.Height

		top = MinInt(top, canvasBox.Bottom)
		left = MinInt(left, tx-(tb.Width>>1))
		right = MaxInt(right, tx+(tb.Width>>1))
		bottom = MaxInt(bottom, ty)
	}

	return Box{
		Top:    top,
		Left:   left,
		Right:  right,
		Bottom: bottom,
		Width:  right - left,
		Height: bottom - top,
	}
}

// Render renders the axis
func (xa XAxis) Render(r Renderer, canvasBox Box, ra Range, ticks []Tick) {
	tickFontSize := xa.Style.GetFontSize(DefaultFontSize)

	r.SetStrokeColor(xa.Style.GetStrokeColor(DefaultAxisColor))
	r.SetStrokeWidth(xa.Style.GetStrokeWidth(DefaultAxisLineWidth))

	r.MoveTo(canvasBox.Left, canvasBox.Bottom)
	r.LineTo(canvasBox.Right, canvasBox.Bottom)
	r.Stroke()

	r.SetFontColor(xa.Style.GetFontColor(DefaultAxisColor))
	r.SetFontSize(tickFontSize)

	sort.Sort(Ticks(ticks))

	for _, t := range ticks {
		v := t.Value
		lx := ra.Translate(v)
		tb := r.MeasureText(t.Label)
		tx := canvasBox.Left + lx
		ty := canvasBox.Bottom + DefaultXAxisMargin + tb.Height
		r.Text(t.Label, tx-tb.Width>>1, ty)

		r.MoveTo(tx, canvasBox.Bottom)
		r.LineTo(tx, canvasBox.Bottom+DefaultVerticalTickHeight)
		r.Stroke()
	}
}
