package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestImageListImages 는 이미지 목록 조회 응답 파싱을 검증한다
func TestImageListImages(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/image/images" {
			t.Errorf("요청 경로가 '/image/images'이어야 하지만 '%s'이다", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("요청 메서드가 GET이어야 하지만 '%s'이다", r.Method)
		}

		resp := ListResponse[Image]{
			Items: []Image{
				{
					ID:           "img-001",
					Name:         "Ubuntu 22.04",
					Status:       "active",
					DiskFormat:   "qcow2",
					ContainerFmt: "bare",
					Size:         2147483648,
					Visibility:   "public",
				},
				{
					ID:           "img-002",
					Name:         "CentOS 9",
					Status:       "active",
					DiskFormat:   "raw",
					ContainerFmt: "bare",
					Size:         3221225472,
					Visibility:   "public",
				},
			},
			Pagination: Pagination{Page: 1, Size: 10, Total: 2},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, WithMaxRetries(0))
	ic := NewImageClient(client)

	result, err := ic.ListImages(nil)
	if err != nil {
		t.Fatalf("ListImages 실패: %v", err)
	}

	if len(result.Items) != 2 {
		t.Fatalf("이미지 수가 2개여야 하지만 %d개이다", len(result.Items))
	}
	if result.Items[0].Name != "Ubuntu 22.04" {
		t.Errorf("첫 번째 이미지 이름이 'Ubuntu 22.04'이어야 하지만 '%s'이다", result.Items[0].Name)
	}
	if result.Items[0].DiskFormat != "qcow2" {
		t.Errorf("첫 번째 이미지 디스크 포맷이 'qcow2'이어야 하지만 '%s'이다", result.Items[0].DiskFormat)
	}
	if result.Items[1].Size != 3221225472 {
		t.Errorf("두 번째 이미지 크기가 3221225472이어야 하지만 %d이다", result.Items[1].Size)
	}
}
