package sdk

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestComputeListServers 는 서버 목록 조회가 올바르게 파싱되는지 검증한다
func TestComputeListServers(t *testing.T) {
	// 모의 서버 설정
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 요청 경로 확인
		if r.URL.Path != "/compute/servers" {
			t.Errorf("요청 경로가 '/compute/servers'이어야 하지만 '%s'이다", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("요청 메서드가 GET이어야 하지만 '%s'이다", r.Method)
		}

		resp := ListResponse[Server]{
			Items: []Server{
				{ID: "srv-001", Name: "web-server", Status: "ACTIVE"},
				{ID: "srv-002", Name: "db-server", Status: "SHUTOFF"},
			},
			Pagination: Pagination{Page: 1, Size: 10, Total: 2},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, WithMaxRetries(0))
	cc := NewComputeClient(client)

	result, err := cc.ListServers(nil)
	if err != nil {
		t.Fatalf("ListServers 실패: %v", err)
	}

	if len(result.Items) != 2 {
		t.Fatalf("서버 수가 2개여야 하지만 %d개이다", len(result.Items))
	}
	if result.Items[0].Name != "web-server" {
		t.Errorf("첫 번째 서버 이름이 'web-server'이어야 하지만 '%s'이다", result.Items[0].Name)
	}
	if result.Items[1].Status != "SHUTOFF" {
		t.Errorf("두 번째 서버 상태가 'SHUTOFF'이어야 하지만 '%s'이다", result.Items[1].Status)
	}
}

// TestComputeCreateServer 는 서버 생성 요청 본문이 올바른지 검증한다
func TestComputeCreateServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/compute/servers" {
			t.Errorf("요청 경로가 '/compute/servers'이어야 하지만 '%s'이다", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("요청 메서드가 POST이어야 하지만 '%s'이다", r.Method)
		}

		// 요청 본문 검증
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("요청 본문 읽기 실패: %v", err)
		}

		var req CreateServerRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("요청 본문 파싱 실패: %v", err)
		}
		if req.Name != "new-server" {
			t.Errorf("서버 이름이 'new-server'이어야 하지만 '%s'이다", req.Name)
		}
		if req.FlavorID != "flavor-1" {
			t.Errorf("FlavorID가 'flavor-1'이어야 하지만 '%s'이다", req.FlavorID)
		}

		// 생성된 서버 응답
		resp := Server{ID: "srv-new", Name: req.Name, Status: "BUILD"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, WithMaxRetries(0))
	cc := NewComputeClient(client)

	server, err := cc.CreateServer(&CreateServerRequest{
		Name:     "new-server",
		FlavorID: "flavor-1",
		ImageID:  "img-1",
	})
	if err != nil {
		t.Fatalf("CreateServer 실패: %v", err)
	}
	if server.ID != "srv-new" {
		t.Errorf("생성된 서버 ID가 'srv-new'이어야 하지만 '%s'이다", server.ID)
	}
}

// TestComputeDeleteServer 는 서버 삭제 시 204 응답에서 에러가 없는지 검증한다
func TestComputeDeleteServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/compute/servers/srv-001" {
			t.Errorf("요청 경로가 '/compute/servers/srv-001'이어야 하지만 '%s'이다", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("요청 메서드가 DELETE이어야 하지만 '%s'이다", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, WithMaxRetries(0))
	cc := NewComputeClient(client)

	err := cc.DeleteServer("srv-001")
	if err != nil {
		t.Fatalf("DeleteServer에서 에러가 발생하면 안 되지만: %v", err)
	}
}
