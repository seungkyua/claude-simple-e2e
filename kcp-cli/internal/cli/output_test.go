package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

// captureStdout 는 표준 출력을 캡처하는 헬퍼이다
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("파이프 생성 실패: %v", err)
	}
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("출력 읽기 실패: %v", err)
	}
	return buf.String()
}

// TestFormatTable 은 헤더와 행 데이터로 테이블 포맷팅이 올바른지 검증한다
func TestFormatTable(t *testing.T) {
	headers := []string{"ID", "Name", "Status"}
	rows := [][]string{
		{"1", "server-a", "ACTIVE"},
		{"2", "server-b", "SHUTOFF"},
	}

	output := captureStdout(t, func() {
		formatTable(headers, rows)
	})

	// 헤더가 포함되어야 한다
	if !strings.Contains(output, "ID") {
		t.Error("출력에 헤더 'ID'가 포함되어야 한다")
	}
	if !strings.Contains(output, "Name") {
		t.Error("출력에 헤더 'Name'이 포함되어야 한다")
	}
	if !strings.Contains(output, "Status") {
		t.Error("출력에 헤더 'Status'가 포함되어야 한다")
	}

	// 행 데이터가 포함되어야 한다
	if !strings.Contains(output, "server-a") {
		t.Error("출력에 'server-a'가 포함되어야 한다")
	}
	if !strings.Contains(output, "SHUTOFF") {
		t.Error("출력에 'SHUTOFF'가 포함되어야 한다")
	}

	// 구분선이 포함되어야 한다
	if !strings.Contains(output, "+") {
		t.Error("출력에 테이블 구분선 '+'가 포함되어야 한다")
	}
}

// TestFormatJSON 은 JSON 출력이 올바른지 검증한다
func TestFormatJSON(t *testing.T) {
	data := map[string]string{
		"id":   "abc-123",
		"name": "test-server",
	}

	output := captureStdout(t, func() {
		formatJSON(data)
	})

	// JSON 파싱이 가능해야 한다
	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("JSON 파싱 실패: %v", err)
	}

	if result["id"] != "abc-123" {
		t.Errorf("id 값이 'abc-123'이어야 하지만 '%s'이다", result["id"])
	}
	if result["name"] != "test-server" {
		t.Errorf("name 값이 'test-server'이어야 하지만 '%s'이다", result["name"])
	}
}

// TestFormatTableWithEmptyData 는 빈 데이터로 테이블 포맷팅을 검증한다
func TestFormatTableWithEmptyData(t *testing.T) {
	// 빈 헤더: 아무것도 출력하지 않아야 한다
	output := captureStdout(t, func() {
		formatTable([]string{}, [][]string{})
	})

	if output != "" {
		t.Errorf("빈 헤더일 때 출력이 비어야 하지만 '%s'이다", output)
	}
}

// TestFormatTableWithEmptyRows 는 헤더만 있고 행이 없을 때를 검증한다
func TestFormatTableWithEmptyRows(t *testing.T) {
	headers := []string{"ID", "Name"}
	output := captureStdout(t, func() {
		formatTable(headers, [][]string{})
	})

	// 헤더는 출력되어야 한다
	if !strings.Contains(output, "ID") {
		t.Error("헤더만 있을 때도 'ID'가 출력되어야 한다")
	}
}

// TestFormatOutput 은 format 플래그에 따른 출력 분기를 검증한다
func TestFormatOutput(t *testing.T) {
	headers := []string{"ID"}
	rows := [][]string{{"1"}}
	data := map[string]string{"id": "1"}

	// JSON 모드
	jsonOut := captureStdout(t, func() {
		formatOutput("json", headers, rows, data)
	})
	if !strings.Contains(jsonOut, "\"id\"") {
		t.Error("json 모드에서 JSON 출력이어야 한다")
	}

	// 기본 (table) 모드
	tableOut := captureStdout(t, func() {
		formatOutput("table", headers, rows, data)
	})
	if !strings.Contains(tableOut, "ID") {
		t.Error("table 모드에서 테이블 출력이어야 한다")
	}
}
