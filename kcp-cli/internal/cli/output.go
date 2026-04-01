// Package cli 의 출력 포맷팅 유틸리티이다.
// table, json 형식을 지원하며 글로벌 --output 플래그에 따라 분기한다.
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// formatTable 은 헤더와 행 데이터를 ASCII 테이블로 출력한다
func formatTable(headers []string, rows [][]string) {
	if len(headers) == 0 {
		return
	}

	// 각 컬럼의 최대 너비를 계산한다
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// 구분선 생성
	sep := "+"
	for _, w := range widths {
		sep += strings.Repeat("-", w+2) + "+"
	}

	// 헤더 출력
	fmt.Println(sep)
	fmt.Print("|")
	for i, h := range headers {
		fmt.Printf(" %-*s |", widths[i], h)
	}
	fmt.Println()
	fmt.Println(sep)

	// 행 출력
	for _, row := range rows {
		fmt.Print("|")
		for i := range headers {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			fmt.Printf(" %-*s |", widths[i], cell)
		}
		fmt.Println()
	}
	fmt.Println(sep)
}

// formatJSON 은 데이터를 JSON 형식으로 출력한다
func formatJSON(data interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		fmt.Fprintf(os.Stderr, "JSON 출력 실패: %v\n", err)
	}
}

// formatOutput 은 format 플래그 값에 따라 적절한 형식으로 출력을 분기한다
func formatOutput(format string, headers []string, rows [][]string, data interface{}) {
	switch format {
	case "json":
		formatJSON(data)
	default:
		formatTable(headers, rows)
	}
}
