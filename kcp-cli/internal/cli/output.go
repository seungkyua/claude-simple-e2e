// Package cli 의 출력 포맷팅 유틸리티이다.
// table, json 형식을 지원하며 글로벌 --output 플래그에 따라 분기한다.
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mattn/go-runewidth"
)

// displayWidth 는 문자열의 터미널 표시 폭을 반환한다.
// 한국어/한자 등 전각 문자는 2칸, ASCII는 1칸으로 계산한다.
func displayWidth(s string) int {
	return runewidth.StringWidth(s)
}

// padRight 는 문자열을 지정된 표시 폭에 맞춰 오른쪽에 공백을 채운다.
// 전각 문자의 폭을 올바르게 처리한다.
func padRight(s string, width int) string {
	sw := displayWidth(s)
	if sw >= width {
		return s
	}
	return s + strings.Repeat(" ", width-sw)
}

// formatTable 은 헤더와 행 데이터를 ASCII 테이블로 출력한다.
// 한국어 등 전각 문자의 표시 폭을 올바르게 처리하여 테이블이 깨지지 않는다.
func formatTable(headers []string, rows [][]string) {
	if len(headers) == 0 {
		return
	}

	// 각 컬럼의 최대 표시 폭을 계산한다
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = displayWidth(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				w := displayWidth(cell)
				if w > widths[i] {
					widths[i] = w
				}
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
		fmt.Printf(" %s |", padRight(h, widths[i]))
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
			fmt.Printf(" %s |", padRight(cell, widths[i]))
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
