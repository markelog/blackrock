package chart

import (
	"fmt"
	"strconv"
	"strings"
)

func makeBar(symbol rune, value int) string {
	s := ""
	for i := int(0); i < value; i++ {
		s += string(symbol)
	}
	return s
}

var multipliers = []string{"", "k", "M", "G", "T", "P"}

func fit(x float64) string {
	div := float64(1)
	var f string
	for _, m := range multipliers {
		f = fmt.Sprintf("%s%s", strconv.FormatFloat(x, 'f', 1, 64), m)
		if len(f) < 8 {
			return f
		}
		div *= float64(1000)
		x /= div

	}

	return f
}

func HorizontalBar(x []float64, y []string, symbol rune, width int, prefix string) string {
	max := float64(0)
	maxLabelWidth := 0
	sum := float64(0)
	for _, v := range x {
		if v > max {
			max = v
		}
		sum += v
	}

	for _, v := range y {
		if len(v) > maxLabelWidth {
			maxLabelWidth = len(v)
		}
	}

	width -= maxLabelWidth + 10 + 8
	lines := []string{}
	pad := fmt.Sprintf("%d", maxLabelWidth)
	for i := 0; i < len(x); i++ {
		v := x[i]
		label := y[i]
		value := int((v / max) * float64(width))

		bar := makeBar(symbol, value)

		line := fmt.Sprintf("%s%-"+pad+"v %8s %6s%% %s", prefix, label, fit(x[i]), fmt.Sprintf("%.2f", 100*(v/sum)), bar)

		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
