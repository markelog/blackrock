package stat

import (
	"math"
)

// copy pasta from https://github.com/lukasvermeer/confidence/blob/master/experiment.js

type Variant struct {
	Visits      uint32
	Convertions uint32
}

func G(data []Variant) float64 {
	v := [][]uint32{}
	for _, variant := range data {
		v = append(v, []uint32{(variant.Visits - variant.Convertions), variant.Convertions})
	}
	return gtest(v)
}
func gtest(data [][]uint32) float64 {
	rows := len(data)
	columns := len(data[0])
	row_totals := make([]uint32, rows)
	column_totals := make([]uint32, columns)

	total := uint32(0)

	for i := 0; i < rows; i++ {
		for j := 0; j < columns; j++ {
			entry := data[i][j]
			row_totals[i] += entry
			column_totals[j] += entry
			total += entry
		}
	}
	g_test := float64(0)
	for i := 0; i < rows; i++ {
		for j := 0; j < columns; j++ {
			expected := float64(row_totals[i]) * float64(column_totals[j]) / float64(total)
			seen := float64(data[i][j])
			g_test += 2 * seen * math.Log(seen/expected)
		}
	}

	return g_test
}