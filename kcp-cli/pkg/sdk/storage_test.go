package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestStorageListVolumes 는 볼륨 목록 조회 응답 파싱을 검증한다
func TestStorageListVolumes(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/storage/volumes" {
			t.Errorf("요청 경로가 '/storage/volumes'이어야 하지만 '%s'이다", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("요청 메서드가 GET이어야 하지만 '%s'이다", r.Method)
		}

		resp := ListResponse[Volume]{
			Items: []Volume{
				{ID: "vol-001", Name: "data-disk", Status: "available", Size: 100, VolumeType: "SSD"},
				{ID: "vol-002", Name: "backup-disk", Status: "in-use", Size: 500, VolumeType: "HDD"},
			},
			Pagination: Pagination{Page: 1, Size: 10, Total: 2},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, WithMaxRetries(0))
	sc := NewStorageClient(client)

	result, err := sc.ListVolumes(nil)
	if err != nil {
		t.Fatalf("ListVolumes 실패: %v", err)
	}

	if len(result.Items) != 2 {
		t.Fatalf("볼륨 수가 2개여야 하지만 %d개이다", len(result.Items))
	}
	if result.Items[0].Name != "data-disk" {
		t.Errorf("첫 번째 볼륨 이름이 'data-disk'이어야 하지만 '%s'이다", result.Items[0].Name)
	}
	if result.Items[0].Size != 100 {
		t.Errorf("첫 번째 볼륨 크기가 100이어야 하지만 %d이다", result.Items[0].Size)
	}
	if result.Items[1].Status != "in-use" {
		t.Errorf("두 번째 볼륨 상태가 'in-use'이어야 하지만 '%s'이다", result.Items[1].Status)
	}
}
