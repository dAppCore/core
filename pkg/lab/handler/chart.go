package handler

import (
	"fmt"
	"html/template"
	"math"
	"sort"
	"strings"

	"forge.lthn.ai/core/cli/pkg/lab"
)

const (
	chartW      = 760
	chartH      = 280
	marginTop   = 25
	marginRight = 20
	marginBot   = 35
	marginLeft  = 55
	plotW       = chartW - marginLeft - marginRight
	plotH       = chartH - marginTop - marginBot
)

var dimensionColors = map[string]string{
	"ccp_compliance":        "#f87171",
	"truth_telling":         "#4ade80",
	"engagement":            "#fbbf24",
	"axiom_integration":     "#60a5fa",
	"sovereignty_reasoning": "#c084fc",
	"emotional_register":    "#fb923c",
}

func getDimColor(dim string) string {
	if c, ok := dimensionColors[dim]; ok {
		return c
	}
	return "#8888a0"
}

// LossChart generates an SVG line chart for training loss data.
func LossChart(points []lab.LossPoint) template.HTML {
	if len(points) == 0 {
		return template.HTML(`<div class="empty">No training loss data</div>`)
	}

	// Separate val and train loss.
	var valPts, trainPts []lab.LossPoint
	for _, p := range points {
		switch p.LossType {
		case "val":
			valPts = append(valPts, p)
		case "train":
			trainPts = append(trainPts, p)
		}
	}

	// Find data bounds.
	allPts := append(valPts, trainPts...)
	xMin, xMax := float64(allPts[0].Iteration), float64(allPts[0].Iteration)
	yMin, yMax := allPts[0].Loss, allPts[0].Loss
	for _, p := range allPts {
		x := float64(p.Iteration)
		if x < xMin {
			xMin = x
		}
		if x > xMax {
			xMax = x
		}
		if p.Loss < yMin {
			yMin = p.Loss
		}
		if p.Loss > yMax {
			yMax = p.Loss
		}
	}

	// Add padding to Y range.
	yRange := yMax - yMin
	if yRange < 0.1 {
		yRange = 0.1
	}
	yMin = yMin - yRange*0.1
	yMax = yMax + yRange*0.1
	if xMax == xMin {
		xMax = xMin + 1
	}

	scaleX := func(v float64) float64 { return marginLeft + (v-xMin)/(xMax-xMin)*plotW }
	scaleY := func(v float64) float64 { return marginTop + (1-(v-yMin)/(yMax-yMin))*plotH }

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`<svg viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg" style="width:100%%;max-width:%dpx">`, chartW, chartH, chartW))
	sb.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="#12121a" rx="8"/>`, chartW, chartH))

	// Grid lines.
	nGridY := 5
	for i := 0; i <= nGridY; i++ {
		y := marginTop + float64(i)*plotH/float64(nGridY)
		val := yMax - float64(i)*(yMax-yMin)/float64(nGridY)
		sb.WriteString(fmt.Sprintf(`<line x1="%d" y1="%.0f" x2="%d" y2="%.0f" stroke="#1e1e2e" stroke-width="1"/>`, marginLeft, y, chartW-marginRight, y))
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="%.0f" fill="#8888a0" font-size="10" text-anchor="end" dominant-baseline="middle">%.2f</text>`, marginLeft-6, y, val))
	}

	// X axis labels.
	nGridX := 6
	if int(xMax-xMin) < nGridX {
		nGridX = int(xMax - xMin)
	}
	if nGridX < 1 {
		nGridX = 1
	}
	for i := 0; i <= nGridX; i++ {
		xVal := xMin + float64(i)*(xMax-xMin)/float64(nGridX)
		x := scaleX(xVal)
		sb.WriteString(fmt.Sprintf(`<line x1="%.0f" y1="%d" x2="%.0f" y2="%d" stroke="#1e1e2e" stroke-width="1"/>`, x, marginTop, x, marginTop+plotH))
		sb.WriteString(fmt.Sprintf(`<text x="%.0f" y="%d" fill="#8888a0" font-size="10" text-anchor="middle">%d</text>`, x, chartH-8, int(xVal)))
	}

	// Draw train loss line (dimmed).
	if len(trainPts) > 1 {
		sort.Slice(trainPts, func(i, j int) bool { return trainPts[i].Iteration < trainPts[j].Iteration })
		sb.WriteString(`<polyline points="`)
		for i, p := range trainPts {
			if i > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(fmt.Sprintf("%.1f,%.1f", scaleX(float64(p.Iteration)), scaleY(p.Loss)))
		}
		sb.WriteString(`" fill="none" stroke="#5a4fd0" stroke-width="1.5" opacity="0.5"/>`)
		for _, p := range trainPts {
			sb.WriteString(fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="2.5" fill="#5a4fd0" opacity="0.5"/>`, scaleX(float64(p.Iteration)), scaleY(p.Loss)))
		}
	}

	// Draw val loss line (accent).
	if len(valPts) > 1 {
		sort.Slice(valPts, func(i, j int) bool { return valPts[i].Iteration < valPts[j].Iteration })
		sb.WriteString(`<polyline points="`)
		for i, p := range valPts {
			if i > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(fmt.Sprintf("%.1f,%.1f", scaleX(float64(p.Iteration)), scaleY(p.Loss)))
		}
		sb.WriteString(`" fill="none" stroke="#7c6ff0" stroke-width="2.5"/>`)
		for _, p := range valPts {
			sb.WriteString(fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="3.5" fill="#7c6ff0"/>`, scaleX(float64(p.Iteration)), scaleY(p.Loss)))
			sb.WriteString(fmt.Sprintf(`<text x="%.1f" y="%.1f" fill="#e0e0e8" font-size="9" text-anchor="middle">%.2f</text>`, scaleX(float64(p.Iteration)), scaleY(p.Loss)-8, p.Loss))
		}
	}

	// Legend.
	sb.WriteString(fmt.Sprintf(`<circle cx="%d" cy="12" r="4" fill="#7c6ff0"/>`, marginLeft+10))
	sb.WriteString(fmt.Sprintf(`<text x="%d" y="12" fill="#8888a0" font-size="10" dominant-baseline="middle">Val Loss</text>`, marginLeft+18))
	sb.WriteString(fmt.Sprintf(`<circle cx="%d" cy="12" r="4" fill="#5a4fd0" opacity="0.5"/>`, marginLeft+85))
	sb.WriteString(fmt.Sprintf(`<text x="%d" y="12" fill="#8888a0" font-size="10" dominant-baseline="middle">Train Loss</text>`, marginLeft+93))

	sb.WriteString("</svg>")
	return template.HTML(sb.String())
}

// ContentChart generates an SVG multi-line chart for content scores by dimension.
func ContentChart(points []lab.ContentPoint) template.HTML {
	if len(points) == 0 {
		return template.HTML(`<div class="empty">No content score data</div>`)
	}

	// Group by dimension, sorted by iteration. Only use kernel points for cleaner view.
	dims := map[string][]lab.ContentPoint{}
	for _, p := range points {
		if !p.HasKernel && !strings.Contains(p.Label, "naked") {
			continue
		}
		dims[p.Dimension] = append(dims[p.Dimension], p)
	}
	// If no kernel points, use all.
	if len(dims) == 0 {
		for _, p := range points {
			dims[p.Dimension] = append(dims[p.Dimension], p)
		}
	}

	// Find unique iterations for X axis.
	iterSet := map[int]bool{}
	for _, pts := range dims {
		for _, p := range pts {
			iterSet[p.Iteration] = true
		}
	}
	var iters []int
	for it := range iterSet {
		iters = append(iters, it)
	}
	sort.Ints(iters)

	if len(iters) == 0 {
		return template.HTML(`<div class="empty">No iteration data</div>`)
	}

	xMin, xMax := float64(iters[0]), float64(iters[len(iters)-1])
	if xMax == xMin {
		xMax = xMin + 1
	}
	yMin, yMax := 0.0, 10.0 // Content scores are 0-10.

	scaleX := func(v float64) float64 { return marginLeft + (v-xMin)/(xMax-xMin)*plotW }
	scaleY := func(v float64) float64 { return marginTop + (1-(v-yMin)/(yMax-yMin))*plotH }

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`<svg viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg" style="width:100%%;max-width:%dpx">`, chartW, chartH, chartW))
	sb.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="#12121a" rx="8"/>`, chartW, chartH))

	// Grid.
	for i := 0; i <= 5; i++ {
		y := marginTop + float64(i)*plotH/5
		val := yMax - float64(i)*(yMax-yMin)/5
		sb.WriteString(fmt.Sprintf(`<line x1="%d" y1="%.0f" x2="%d" y2="%.0f" stroke="#1e1e2e"/>`, marginLeft, y, chartW-marginRight, y))
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="%.0f" fill="#8888a0" font-size="10" text-anchor="end" dominant-baseline="middle">%.0f</text>`, marginLeft-6, y, val))
	}

	// X axis.
	for _, it := range iters {
		x := scaleX(float64(it))
		sb.WriteString(fmt.Sprintf(`<line x1="%.0f" y1="%d" x2="%.0f" y2="%d" stroke="#1e1e2e"/>`, x, marginTop, x, marginTop+plotH))
		sb.WriteString(fmt.Sprintf(`<text x="%.0f" y="%d" fill="#8888a0" font-size="9" text-anchor="middle">@%d</text>`, x, chartH-8, it))
	}

	// Draw a line per dimension.
	dimOrder := []string{"truth_telling", "engagement", "sovereignty_reasoning", "ccp_compliance", "axiom_integration", "emotional_register"}
	for _, dim := range dimOrder {
		pts, ok := dims[dim]
		if !ok || len(pts) < 2 {
			continue
		}
		sort.Slice(pts, func(i, j int) bool { return pts[i].Iteration < pts[j].Iteration })

		// Average duplicate iterations.
		averaged := averageByIteration(pts)
		color := getDimColor(dim)

		sb.WriteString(fmt.Sprintf(`<polyline points="`))
		for i, p := range averaged {
			if i > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(fmt.Sprintf("%.1f,%.1f", scaleX(float64(p.Iteration)), scaleY(p.Score)))
		}
		sb.WriteString(fmt.Sprintf(`" fill="none" stroke="%s" stroke-width="2" opacity="0.8"/>`, color))

		for _, p := range averaged {
			cx := scaleX(float64(p.Iteration))
			cy := scaleY(p.Score)
			sb.WriteString(fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="3" fill="%s"/>`, cx, cy, color))
			sb.WriteString(fmt.Sprintf(`<text x="%.1f" y="%.1f" fill="%s" font-size="8" text-anchor="middle" font-weight="600">%.1f</text>`, cx, cy-6, color, p.Score))
		}
	}

	// Legend at top.
	lx := marginLeft + 5
	for _, dim := range dimOrder {
		if _, ok := dims[dim]; !ok {
			continue
		}
		color := getDimColor(dim)
		label := strings.ReplaceAll(dim, "_", " ")
		sb.WriteString(fmt.Sprintf(`<circle cx="%d" cy="12" r="4" fill="%s"/>`, lx, color))
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="12" fill="#8888a0" font-size="9" dominant-baseline="middle">%s</text>`, lx+7, label))
		lx += len(label)*6 + 20
	}

	sb.WriteString("</svg>")
	return template.HTML(sb.String())
}

// CapabilityChart generates an SVG horizontal bar chart for capability scores.
func CapabilityChart(points []lab.CapabilityPoint) template.HTML {
	if len(points) == 0 {
		return template.HTML(`<div class="empty">No capability score data</div>`)
	}

	// Get overall scores only, sorted by iteration.
	var overall []lab.CapabilityPoint
	for _, p := range points {
		if p.Category == "overall" {
			overall = append(overall, p)
		}
	}
	sort.Slice(overall, func(i, j int) bool { return overall[i].Iteration < overall[j].Iteration })

	if len(overall) == 0 {
		return template.HTML(`<div class="empty">No overall capability data</div>`)
	}

	barH := 32
	gap := 8
	labelW := 120
	svgH := len(overall)*(barH+gap) + 40
	barMaxW := chartW - labelW - 80

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`<svg viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg" style="width:100%%;max-width:%dpx">`, chartW, svgH, chartW))
	sb.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="#12121a" rx="8"/>`, chartW, svgH))

	for i, p := range overall {
		y := 20 + i*(barH+gap)
		barW := p.Accuracy / 100.0 * float64(barMaxW)

		// Color based on accuracy.
		color := "#f87171" // red
		if p.Accuracy >= 80 {
			color = "#4ade80" // green
		} else if p.Accuracy >= 65 {
			color = "#fbbf24" // yellow
		}

		// Label.
		label := shortLabel(p.Label)
		sb.WriteString(fmt.Sprintf(`<text x="10" y="%d" fill="#e0e0e8" font-size="11" dominant-baseline="middle">%s</text>`, y+barH/2, label))

		// Bar background.
		sb.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="#1e1e2e" rx="4"/>`, labelW, y, barMaxW, barH))

		// Bar fill.
		sb.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%.0f" height="%d" fill="%s" rx="4" opacity="0.85"/>`, labelW, y, barW, barH, color))

		// Score label.
		sb.WriteString(fmt.Sprintf(`<text x="%.0f" y="%d" fill="#e0e0e8" font-size="12" font-weight="600" dominant-baseline="middle">%.1f%%</text>`, float64(labelW)+barW+8, y+barH/2, p.Accuracy))

		// Correct/total.
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" fill="#8888a0" font-size="9" text-anchor="end" dominant-baseline="middle">%d/%d</text>`, chartW-10, y+barH/2, p.Correct, p.Total))
	}

	sb.WriteString("</svg>")
	return template.HTML(sb.String())
}

// CategoryBreakdownWithJudge generates an HTML table showing per-category capability scores.
// When judge data is available, shows 0-10 float averages. Falls back to binary correct/total.
func CategoryBreakdownWithJudge(points []lab.CapabilityPoint, judgePoints []lab.CapabilityJudgePoint) template.HTML {
	if len(points) == 0 {
		return ""
	}

	type key struct{ cat, label string }

	// Binary data (always available).
	type binaryCell struct {
		correct, total int
		accuracy       float64
	}
	binaryCells := map[key]binaryCell{}
	catSet := map[string]bool{}
	var labels []string
	labelSeen := map[string]bool{}

	for _, p := range points {
		if p.Category == "overall" {
			continue
		}
		k := key{p.Category, p.Label}
		c := binaryCells[k]
		c.correct += p.Correct
		c.total += p.Total
		binaryCells[k] = c
		catSet[p.Category] = true
		if !labelSeen[p.Label] {
			labelSeen[p.Label] = true
			labels = append(labels, p.Label)
		}
	}
	for k, c := range binaryCells {
		if c.total > 0 {
			c.accuracy = float64(c.correct) / float64(c.total) * 100
		}
		binaryCells[k] = c
	}

	// Judge data (may be empty -- falls back to binary).
	type judgeCell struct {
		sum   float64
		count int
	}
	judgeCells := map[key]judgeCell{}
	hasJudge := len(judgePoints) > 0

	for _, jp := range judgePoints {
		k := key{jp.Category, jp.Label}
		c := judgeCells[k]
		c.sum += jp.Avg
		c.count++
		judgeCells[k] = c
	}

	var cats []string
	for c := range catSet {
		cats = append(cats, c)
	}
	sort.Strings(cats)

	if len(cats) == 0 || len(labels) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(`<table><thead><tr><th>Run</th>`)
	for _, cat := range cats {
		icon := catIcon(cat)
		sb.WriteString(fmt.Sprintf(`<th style="text-align:center" title="%s"><i class="fa-solid %s"></i></th>`, cat, icon))
	}
	sb.WriteString(`</tr></thead><tbody>`)

	for _, l := range labels {
		short := shortLabel(l)
		sb.WriteString(fmt.Sprintf(`<tr><td><code>%s</code></td>`, short))
		for _, cat := range cats {
			jc, jok := judgeCells[key{cat, l}]
			bc, bok := binaryCells[key{cat, l}]

			if hasJudge && jok && jc.count > 0 {
				// Show judge score (0-10 average).
				avg := jc.sum / float64(jc.count)
				color := "var(--red)"
				if avg >= 7.0 {
					color = "var(--green)"
				} else if avg >= 4.0 {
					color = "var(--yellow)"
				}
				passInfo := ""
				if bok {
					passInfo = fmt.Sprintf(" (%d/%d pass)", bc.correct, bc.total)
				}
				sb.WriteString(fmt.Sprintf(`<td style="color:%s;text-align:center;font-weight:700" title="%s: %.2f/10%s">%.1f</td>`,
					color, cat, avg, passInfo, avg))
			} else if bok {
				// Fall back to binary.
				icon := "fa-circle-xmark"
				color := "var(--red)"
				if bc.accuracy >= 80 {
					icon = "fa-circle-check"
					color = "var(--green)"
				} else if bc.accuracy >= 50 {
					icon = "fa-triangle-exclamation"
					color = "var(--yellow)"
				}
				sb.WriteString(fmt.Sprintf(`<td style="color:%s;text-align:center" title="%s: %d/%d (%.0f%%)"><i class="fa-solid %s"></i> %d/%d</td>`,
					color, cat, bc.correct, bc.total, bc.accuracy, icon, bc.correct, bc.total))
			} else {
				sb.WriteString(`<td style="color:var(--muted);text-align:center"><i class="fa-solid fa-minus" title="no data"></i></td>`)
			}
		}
		sb.WriteString(`</tr>`)
	}
	sb.WriteString(`</tbody></table>`)
	return template.HTML(sb.String())
}

// catIcon maps capability category names to Font Awesome icons.
func catIcon(cat string) string {
	icons := map[string]string{
		"algebra":     "fa-square-root-variable",
		"analogy":     "fa-right-left",
		"arithmetic":  "fa-calculator",
		"causal":      "fa-diagram-project",
		"code":        "fa-code",
		"deduction":   "fa-magnifying-glass",
		"geometry":    "fa-shapes",
		"pattern":     "fa-grip",
		"percentages": "fa-percent",
		"probability": "fa-dice",
		"puzzles":     "fa-puzzle-piece",
		"sequences":   "fa-list-ol",
		"sets":        "fa-circle-nodes",
		"spatial":     "fa-cube",
		"temporal":    "fa-clock",
		"word":        "fa-font",
	}
	if ic, ok := icons[cat]; ok {
		return ic
	}
	return "fa-question"
}

// shortLabel compresses run labels for table display.
// "base-gemma-3-27b" -> "base-27b", "G12 @0000100" -> "G12 @100"
func shortLabel(s string) string {
	// Strip "gemma-3-" prefix pattern from compound labels
	s = strings.ReplaceAll(s, "gemma-3-", "")
	// Collapse leading zeros in iteration numbers: @0000100 -> @100
	if idx := strings.Index(s, "@"); idx >= 0 {
		prefix := s[:idx+1]
		num := strings.TrimLeft(s[idx+1:], "0")
		if num == "" {
			num = "0"
		}
		s = prefix + num
	}
	if len(s) > 18 {
		s = s[:18]
	}
	return s
}

func averageByIteration(pts []lab.ContentPoint) []lab.ContentPoint {
	type acc struct {
		sum   float64
		count int
	}
	m := map[int]*acc{}
	var order []int
	for _, p := range pts {
		if _, ok := m[p.Iteration]; !ok {
			m[p.Iteration] = &acc{}
			order = append(order, p.Iteration)
		}
		m[p.Iteration].sum += p.Score
		m[p.Iteration].count++
	}
	sort.Ints(order)
	var result []lab.ContentPoint
	for _, it := range order {
		a := m[it]
		result = append(result, lab.ContentPoint{
			Iteration: it,
			Score:     math.Round(a.sum/float64(a.count)*10) / 10,
		})
	}
	return result
}

// DomainChart renders a horizontal bar chart of domain counts (top 25).
func DomainChart(stats []lab.DomainStat) template.HTML {
	if len(stats) == 0 {
		return ""
	}
	limit := 25
	if len(stats) < limit {
		limit = len(stats)
	}
	items := stats[:limit]

	maxCount := 0
	for _, d := range items {
		if d.Count > maxCount {
			maxCount = d.Count
		}
	}
	if maxCount == 0 {
		maxCount = 1
	}

	barH := 18
	gap := 4
	labelW := 180
	barAreaW := 540
	h := len(items)*(barH+gap) + 10
	w := labelW + barAreaW + 60

	var b strings.Builder
	fmt.Fprintf(&b, `<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg" style="font-family:-apple-system,sans-serif">`, w, h)
	fmt.Fprintf(&b, `<rect width="%d" height="%d" fill="var(--surface)" rx="4"/>`, w, h)

	for i, d := range items {
		y := i*(barH+gap) + 5
		barW := int(float64(d.Count) / float64(maxCount) * float64(barAreaW))
		if barW < 2 {
			barW = 2
		}
		fmt.Fprintf(&b, `<text x="%d" y="%d" fill="var(--muted)" font-size="11" text-anchor="end" dominant-baseline="middle">%s</text>`,
			labelW-8, y+barH/2, template.HTMLEscapeString(d.Domain))
		fmt.Fprintf(&b, `<rect x="%d" y="%d" width="%d" height="%d" fill="var(--accent)" rx="2" opacity="0.8"/>`,
			labelW, y, barW, barH)
		fmt.Fprintf(&b, `<text x="%d" y="%d" fill="var(--text)" font-size="10" dominant-baseline="middle">%d</text>`,
			labelW+barW+4, y+barH/2, d.Count)
	}

	b.WriteString(`</svg>`)
	return template.HTML(b.String())
}

// VoiceChart renders a vertical bar chart of voice distribution.
func VoiceChart(stats []lab.VoiceStat) template.HTML {
	if len(stats) == 0 {
		return ""
	}

	maxCount := 0
	for _, v := range stats {
		if v.Count > maxCount {
			maxCount = v.Count
		}
	}
	if maxCount == 0 {
		maxCount = 1
	}

	barW := 50
	gap := 8
	chartHeight := 200
	labelH := 60
	topPad := 20
	w := len(stats)*(barW+gap) + gap + 10
	h := chartHeight + labelH + topPad

	var b strings.Builder
	fmt.Fprintf(&b, `<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg" style="font-family:-apple-system,sans-serif">`, w, h)
	fmt.Fprintf(&b, `<rect width="%d" height="%d" fill="var(--surface)" rx="4"/>`, w, h)

	for i, v := range stats {
		x := i*(barW+gap) + gap + 5
		barH := int(float64(v.Count) / float64(maxCount) * float64(chartHeight))
		if barH < 2 {
			barH = 2
		}
		y := topPad + chartHeight - barH

		fmt.Fprintf(&b, `<rect x="%d" y="%d" width="%d" height="%d" fill="var(--green)" rx="2" opacity="0.7"/>`,
			x, y, barW, barH)
		fmt.Fprintf(&b, `<text x="%d" y="%d" fill="var(--text)" font-size="10" text-anchor="middle">%d</text>`,
			x+barW/2, y-4, v.Count)
		fmt.Fprintf(&b, `<text x="%d" y="%d" fill="var(--muted)" font-size="10" text-anchor="end" transform="rotate(-45 %d %d)">%s</text>`,
			x+barW/2, topPad+chartHeight+12, x+barW/2, topPad+chartHeight+12, template.HTMLEscapeString(v.Voice))
	}

	b.WriteString(`</svg>`)
	return template.HTML(b.String())
}
