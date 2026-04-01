package openstack

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestNormalizeAuthURL 은 다양한 형태의 Auth URL을 올바르게 정규화하는지 검증한다
func TestNormalizeAuthURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "/v3가 없는 경우 자동 추가",
			input:    "http://host/identity",
			expected: "http://host/identity/v3",
		},
		{
			name:     "이미 /v3가 있으면 변경 없음",
			input:    "http://host:5000/v3",
			expected: "http://host:5000/v3",
		},
		{
			name:     "trailing slash 제거 후 /v3 추가",
			input:    "http://host/identity/",
			expected: "http://host/identity/v3",
		},
		{
			name:     "/v3/ trailing slash 제거",
			input:    "http://host/identity/v3/",
			expected: "http://host/identity/v3",
		},
		{
			name:     "포트 포함 URL",
			input:    "http://211.34.229.207:5000",
			expected: "http://211.34.229.207:5000/v3",
		},
		{
			name:     "실제 openrc URL (포트 없음)",
			input:    "http://211.34.229.207/identity",
			expected: "http://211.34.229.207/identity/v3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeAuthURL(tt.input)
			if got != tt.expected {
				t.Errorf("normalizeAuthURL(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestBuildAuthRequest 는 OSConfig에서 올바른 Keystone 인증 요청을 생성하는지 검증한다
func TestBuildAuthRequest(t *testing.T) {
	cfg := &OSConfig{
		AuthURL:         "http://host/identity",
		Username:        "admin",
		Password:        "secret",
		ProjectName:     "admin",
		ProjectDomainID: "default",
		UserDomainID:    "default",
		RegionName:      "RegionOne",
	}

	req := buildAuthRequest(cfg)

	// identity 검증
	if req.Auth.Identity.Methods[0] != "password" {
		t.Errorf("method = %s, want password", req.Auth.Identity.Methods[0])
	}
	if req.Auth.Identity.Password.User.Name != "admin" {
		t.Errorf("username = %s, want admin", req.Auth.Identity.Password.User.Name)
	}
	if req.Auth.Identity.Password.User.Password != "secret" {
		t.Errorf("password = %s, want secret", req.Auth.Identity.Password.User.Password)
	}

	// user domain 검증
	if req.Auth.Identity.Password.User.Domain == nil {
		t.Fatal("user domain is nil")
	}
	if req.Auth.Identity.Password.User.Domain.ID != "default" {
		t.Errorf("user_domain_id = %s, want default", req.Auth.Identity.Password.User.Domain.ID)
	}

	// project scope 검증
	if req.Auth.Scope == nil || req.Auth.Scope.Project == nil {
		t.Fatal("project scope is nil")
	}
	if req.Auth.Scope.Project.Name != "admin" {
		t.Errorf("project_name = %s, want admin", req.Auth.Scope.Project.Name)
	}
	if req.Auth.Scope.Project.Domain == nil {
		t.Fatal("project domain is nil")
	}
	if req.Auth.Scope.Project.Domain.ID != "default" {
		t.Errorf("project_domain_id = %s, want default", req.Auth.Scope.Project.Domain.ID)
	}
}

// TestBuildAuthRequestWithProjectID 는 ProjectID가 있을 때 올바른 스코프를 생성하는지 검증한다
func TestBuildAuthRequestWithProjectID(t *testing.T) {
	cfg := &OSConfig{
		Username:        "admin",
		Password:        "secret",
		ProjectID:       "proj-123",
		ProjectDomainID: "domain-456",
		UserDomainID:    "default",
	}

	req := buildAuthRequest(cfg)

	if req.Auth.Scope.Project.ID != "proj-123" {
		t.Errorf("project_id = %s, want proj-123", req.Auth.Scope.Project.ID)
	}
	if req.Auth.Scope.Project.Domain.ID != "domain-456" {
		t.Errorf("project_domain_id = %s, want domain-456", req.Auth.Scope.Project.Domain.ID)
	}
}

// TestAuthenticate 는 mock Keystone 서버로 인증 흐름을 검증한다
func TestAuthenticate(t *testing.T) {
	// mock Keystone 서버
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 경로 확인: /v3/auth/tokens
		if r.URL.Path != "/v3/auth/tokens" {
			t.Errorf("요청 경로 = %s, want /v3/auth/tokens", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Method != "POST" {
			t.Errorf("method = %s, want POST", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// 요청 본문 검증
		var authReq keystoneAuthRequest
		json.NewDecoder(r.Body).Decode(&authReq)
		if authReq.Auth.Identity.Password.User.Name != "admin" {
			t.Errorf("username = %s, want admin", authReq.Auth.Identity.Password.User.Name)
		}

		// 응답: 토큰 + 카탈로그
		w.Header().Set("X-Subject-Token", "test-token-abc123")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"token": map[string]interface{}{
				"expires_at": time.Now().Add(1 * time.Hour).Format(time.RFC3339),
				"catalog": []map[string]interface{}{
					{
						"type": "compute",
						"name": "nova",
						"endpoints": []map[string]interface{}{
							{"url": "http://nova:8774/v2.1", "interface": "public", "region_id": "RegionOne"},
						},
					},
					{
						"type": "network",
						"name": "neutron",
						"endpoints": []map[string]interface{}{
							{"url": "http://neutron:9696", "interface": "public", "region_id": "RegionOne"},
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	// 클라이언트 생성 (mock 서버 URL 사용)
	client, err := NewClient(&OSConfig{
		AuthURL:         server.URL, // /v3가 없으므로 자동 추가
		Username:        "admin",
		Password:        "secret",
		ProjectName:     "admin",
		ProjectDomainID: "default",
		UserDomainID:    "default",
		RegionName:      "RegionOne",
	})
	if err != nil {
		t.Fatalf("NewClient 실패: %v", err)
	}

	// 토큰 검증
	token, err := client.GetToken()
	if err != nil {
		t.Fatalf("GetToken 실패: %v", err)
	}
	if token != "test-token-abc123" {
		t.Errorf("token = %s, want test-token-abc123", token)
	}

	// 엔드포인트 검증
	computeURL, err := client.GetEndpoint("compute")
	if err != nil {
		t.Fatalf("GetEndpoint(compute) 실패: %v", err)
	}
	if computeURL != "http://nova:8774/v2.1" {
		t.Errorf("compute endpoint = %s, want http://nova:8774/v2.1", computeURL)
	}

	networkURL, err := client.GetEndpoint("network")
	if err != nil {
		t.Fatalf("GetEndpoint(network) 실패: %v", err)
	}
	if networkURL != "http://neutron:9696" {
		t.Errorf("network endpoint = %s, want http://neutron:9696", networkURL)
	}

	// 존재하지 않는 서비스
	_, err = client.GetEndpoint("nonexistent")
	if err == nil {
		t.Error("존재하지 않는 서비스에 대해 에러가 반환되어야 합니다")
	}
}

// TestDoRequest 는 인증된 API 요청이 올바르게 전송되는지 검증한다
func TestDoRequest(t *testing.T) {
	// mock 서버: Keystone + Nova
	mux := http.NewServeMux()

	// Keystone 인증
	mux.HandleFunc("/v3/auth/tokens", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Subject-Token", "test-token")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"token": map[string]interface{}{
				"expires_at": time.Now().Add(1 * time.Hour).Format(time.RFC3339),
				"catalog": []map[string]interface{}{
					{
						"type": "compute",
						"name": "nova",
						"endpoints": []map[string]interface{}{
							{"url": "", "interface": "public", "region_id": "RegionOne"},
						},
					},
				},
			},
		})
	})

	// Nova 서버 목록
	mux.HandleFunc("/v2.1/servers/detail", func(w http.ResponseWriter, r *http.Request) {
		// X-Auth-Token 헤더 확인
		if r.Header.Get("X-Auth-Token") != "test-token" {
			t.Errorf("X-Auth-Token = %s, want test-token", r.Header.Get("X-Auth-Token"))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"servers": []map[string]string{
				{"id": "server-1", "name": "test-vm"},
			},
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	// Nova 엔드포인트를 mock 서버 URL로 설정하기 위해 카탈로그 URL을 동적으로 변경
	mux2 := http.NewServeMux()
	mux2.HandleFunc("/v3/auth/tokens", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Subject-Token", "test-token")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"token": map[string]interface{}{
				"expires_at": time.Now().Add(1 * time.Hour).Format(time.RFC3339),
				"catalog": []map[string]interface{}{
					{
						"type": "compute",
						"name": "nova",
						"endpoints": []map[string]interface{}{
							{"url": server.URL + "/v2.1", "interface": "public", "region_id": "RegionOne"},
						},
					},
				},
			},
		})
	})
	keystoneServer := httptest.NewServer(mux2)
	defer keystoneServer.Close()

	client, err := NewClient(&OSConfig{
		AuthURL:         keystoneServer.URL,
		Username:        "admin",
		Password:        "secret",
		ProjectName:     "admin",
		ProjectDomainID: "default",
		UserDomainID:    "default",
		RegionName:      "RegionOne",
	})
	if err != nil {
		t.Fatalf("NewClient 실패: %v", err)
	}

	// DoRequest로 서버 목록 조회
	data, statusCode, err := client.DoRequest("GET", "compute", "/servers/detail", nil)
	if err != nil {
		t.Fatalf("DoRequest 실패: %v", err)
	}
	if statusCode != 200 {
		t.Errorf("status = %d, want 200", statusCode)
	}

	var result map[string]interface{}
	json.Unmarshal(data, &result)
	servers := result["servers"].([]interface{})
	if len(servers) != 1 {
		t.Errorf("servers count = %d, want 1", len(servers))
	}
}

// TestExtractErrorMessage 는 에러 응답에서 의미 있는 메시지를 추출하는지 검증한다
func TestExtractErrorMessage(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		statusCode   int
		wantContains string
	}{
		{
			name:         "JSON 에러 응답",
			body:         `{"error":{"message":"The request you have made requires authentication.","code":401}}`,
			statusCode:   401,
			wantContains: "The request you have made requires authentication",
		},
		{
			name:         "HTML 500 에러",
			body:         "<html><title>500 Internal Server Error</title></html>",
			statusCode:   500,
			wantContains: "OpenStack 서버 내부 오류",
		},
		{
			name:         "HTML 404 에러",
			body:         "<html><title>404 Not Found</title></html>",
			statusCode:   404,
			wantContains: "Keystone 엔드포인트를 찾을 수 없습니다",
		},
		{
			name:         "HTML 401 에러",
			body:         "<html>Unauthorized</html>",
			statusCode:   401,
			wantContains: "인증 정보가 올바르지 않습니다",
		},
		{
			name:         "HTML 503 에러",
			body:         "<html>Service Unavailable</html>",
			statusCode:   503,
			wantContains: "사용할 수 없습니다",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractErrorMessage([]byte(tt.body), tt.statusCode)
			if !strings.Contains(got, tt.wantContains) {
				t.Errorf("extractErrorMessage() = %q, want contains %q", got, tt.wantContains)
			}
		})
	}
}

// TestNewClientLazyAuth 는 인증 실패 시에도 클라이언트가 반환되는지 검증한다
func TestNewClientLazyAuth(t *testing.T) {
	// 인증 실패하는 mock 서버
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":{"message":"Internal Server Error","code":500}}`))
	}))
	defer server.Close()

	client, err := NewClient(&OSConfig{
		AuthURL:      server.URL,
		Username:     "admin",
		Password:     "wrong",
		ProjectName:  "admin",
		UserDomainID: "default",
	})

	// 에러는 반환되지만 클라이언트도 반환되어야 한다
	if err == nil {
		t.Error("인증 실패 시 에러가 반환되어야 합니다")
	}
	if client == nil {
		t.Fatal("인증 실패 시에도 클라이언트가 반환되어야 합니다 (지연 인증)")
	}
}

// TestExtractList 는 JSON 응답에서 배열을 올바르게 추출하는지 검증한다
func TestExtractList(t *testing.T) {
	data := []byte(`{"servers":[{"id":"s1"},{"id":"s2"}]}`)
	items, err := extractList(data, "servers")
	if err != nil {
		t.Fatalf("extractList 실패: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("items count = %d, want 2", len(items))
	}

	// 존재하지 않는 키
	_, err = extractList(data, "nonexistent")
	if err == nil {
		t.Error("존재하지 않는 키에 대해 에러가 반환되어야 합니다")
	}
}

// TestExtractSingle 은 JSON 응답에서 단일 객체를 올바르게 추출하는지 검증한다
func TestExtractSingle(t *testing.T) {
	data := []byte(`{"server":{"id":"s1","name":"test"}}`)
	item, err := extractSingle(data, "server")
	if err != nil {
		t.Fatalf("extractSingle 실패: %v", err)
	}

	var server map[string]string
	json.Unmarshal(item, &server)
	if server["id"] != "s1" {
		t.Errorf("server id = %s, want s1", server["id"])
	}

	// 존재하지 않는 키
	_, err = extractSingle(data, "nonexistent")
	if err == nil {
		t.Error("존재하지 않는 키에 대해 에러가 반환되어야 합니다")
	}
}

// TestCheckStatusError 는 HTTP 상태 코드에 따른 에러 반환을 검증한다
func TestCheckStatusError(t *testing.T) {
	// 정상 상태 코드
	err := checkStatusError([]byte("{}"), 200, "테스트")
	if err != nil {
		t.Errorf("200 상태에서 에러가 반환되면 안 됩니다: %v", err)
	}

	err = checkStatusError([]byte("{}"), 204, "테스트")
	if err != nil {
		t.Errorf("204 상태에서 에러가 반환되면 안 됩니다: %v", err)
	}

	// 에러 상태 코드
	err = checkStatusError([]byte(`{"error":{"message":"Not Found","code":404}}`), 404, "리소스 조회")
	if err == nil {
		t.Error("404 상태에서 에러가 반환되어야 합니다")
	}
	if !strings.Contains(err.Error(), "리소스 조회 실패") {
		t.Errorf("에러 메시지에 작업명이 포함되어야 합니다: %v", err)
	}
}
