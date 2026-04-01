package middleware

import (
	"testing"
)

// TestDetectSource 는 User-Agent 문자열에서 올바른 클라이언트 소스를 판별하는지 검증한다
func TestDetectSource(t *testing.T) {
	tests := []struct {
		name      string
		userAgent string
		want      string
	}{
		{
			name:      "CLI 클라이언트 감지",
			userAgent: "kcp-cli/1.0.0",
			want:      "CLI",
		},
		{
			name:      "TUI 클라이언트 감지",
			userAgent: "kcp-tui/2.0.0",
			want:      "TUI",
		},
		{
			name:      "대소문자 구분 없이 CLI 감지",
			userAgent: "KCP-CLI/1.0",
			want:      "CLI",
		},
		{
			name:      "일반 브라우저는 WEBUI로 판별",
			userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X)",
			want:      "WEBUI",
		},
		{
			name:      "빈 User-Agent는 WEBUI로 판별",
			userAgent: "",
			want:      "WEBUI",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectSource(tt.userAgent)
			if got != tt.want {
				t.Errorf("detectSource(%q) = %q, 기대값 %q", tt.userAgent, got, tt.want)
			}
		})
	}
}

// TestMapMethodToAction 은 HTTP 메서드가 올바른 감사 로그 액션으로 변환되는지 검증한다
func TestMapMethodToAction(t *testing.T) {
	tests := []struct {
		method string
		want   string
	}{
		{"GET", "READ"},
		{"POST", "CREATE"},
		{"PUT", "UPDATE"},
		{"PATCH", "UPDATE"},
		{"DELETE", "DELETE"},
		{"HEAD", "READ"}, // 알 수 없는 메서드는 READ
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			got := mapMethodToAction(tt.method)
			if got != tt.want {
				t.Errorf("mapMethodToAction(%q) = %q, 기대값 %q", tt.method, got, tt.want)
			}
		})
	}
}

// TestExtractResource 는 URL 경로에서 올바른 리소스 타입과 ID를 추출하는지 검증한다
func TestExtractResource(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		wantType     string
		wantID       string
	}{
		{
			name:     "서버 리소스 + ID 추출",
			path:     "/api/v1/compute/servers/abc-123",
			wantType: "VM",
			wantID:   "abc-123",
		},
		{
			name:     "네트워크 리소스 목록 (ID 없음)",
			path:     "/api/v1/networks",
			wantType: "NETWORK",
			wantID:   "",
		},
		{
			name:     "볼륨 리소스 + ID 추출",
			path:     "/api/v1/storage/volumes/vol-456",
			wantType: "VOLUME",
			wantID:   "vol-456",
		},
		{
			name:     "짧은 경로는 UNKNOWN 반환",
			path:     "/api",
			wantType: "UNKNOWN",
			wantID:   "",
		},
		{
			name:     "알 수 없는 리소스는 UNKNOWN 반환",
			path:     "/api/v1/unknown/resource",
			wantType: "UNKNOWN",
			wantID:   "",
		},
		{
			name:     "보안 그룹 리소스 추출",
			path:     "/api/v1/network/security-groups/sg-789",
			wantType: "SECURITY_GROUP",
			wantID:   "sg-789",
		},
		{
			name:     "사용자 리소스 목록",
			path:     "/api/v1/admin/users",
			wantType: "USER",
			wantID:   "",
		},
		{
			name:     "서버 액션 경로 (known path 이후)",
			path:     "/api/v1/compute/servers/abc-123/action",
			wantType: "VM",
			wantID:   "abc-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotID := extractResource(tt.path)
			if gotType != tt.wantType {
				t.Errorf("extractResource(%q) 리소스 타입 = %q, 기대값 %q", tt.path, gotType, tt.wantType)
			}
			if gotID != tt.wantID {
				t.Errorf("extractResource(%q) 리소스 ID = %q, 기대값 %q", tt.path, gotID, tt.wantID)
			}
		})
	}
}
