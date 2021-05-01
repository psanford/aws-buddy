package console

import (
	"fmt"
	"strings"
)

func FormatTable(rows [][]string) string {
	if len(rows) == 0 {
		return ""
	}
	widths := make([]int, len(rows[0]))
	for _, row := range rows {
		for i, s := range row {
			size := len(s)
			if size > widths[i] {
				widths[i] = size
			}
		}
	}

	var out strings.Builder
	for _, row := range rows {
		for colIdx, s := range row {
			size := len(s)
			pad := widths[colIdx] - size
			out.WriteString(s)
			padStr := strings.Repeat(" ", pad)
			out.WriteString(padStr)
			if colIdx != len(row)-1 {
				out.WriteString(" | ")
			}
		}
		out.WriteString("\n")
	}
	return out.String()
}

func Confirm(prompt string) bool {
	fmt.Print(prompt)
	var result string
	fmt.Scanln(&result)

	return result == "y" || result == "Y" || result == "yes" || result == "Yes"
}
