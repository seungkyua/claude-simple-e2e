package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestIdentityListProjects 는 프로젝트 목록 조회 응답 파싱을 검증한다
func TestIdentityListProjects(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/identity/projects" {
			t.Errorf("요청 경로가 '/identity/projects'이어야 하지만 '%s'이다", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("요청 메서드가 GET이어야 하지만 '%s'이다", r.Method)
		}

		resp := ListResponse[Project]{
			Items: []Project{
				{ID: "proj-001", Name: "production", Description: "운영 환경", Enabled: true, DomainID: "default"},
				{ID: "proj-002", Name: "staging", Description: "스테이징 환경", Enabled: true, DomainID: "default"},
			},
			Pagination: Pagination{Page: 1, Size: 10, Total: 2},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, WithMaxRetries(0))
	ic := NewIdentityClient(client)

	result, err := ic.ListProjects(nil)
	if err != nil {
		t.Fatalf("ListProjects 실패: %v", err)
	}

	if len(result.Items) != 2 {
		t.Fatalf("프로젝트 수가 2개여야 하지만 %d개이다", len(result.Items))
	}
	if result.Items[0].Name != "production" {
		t.Errorf("첫 번째 프로젝트 이름이 'production'이어야 하지만 '%s'이다", result.Items[0].Name)
	}
	if !result.Items[0].Enabled {
		t.Error("첫 번째 프로젝트가 활성 상태여야 한다")
	}
	if result.Items[1].Description != "스테이징 환경" {
		t.Errorf("두 번째 프로젝트 설명이 '스테이징 환경'이어야 하지만 '%s'이다", result.Items[1].Description)
	}
}
