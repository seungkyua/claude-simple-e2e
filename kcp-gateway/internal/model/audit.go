package model

import "time"

// AuditLog 는 감사 로그 모델이다
type AuditLog struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	Username       string    `json:"username,omitempty"`
	Action         string    `json:"action"`
	ResourceType   string    `json:"resource_type"`
	ResourceID     string    `json:"resource_id,omitempty"`
	Source         string    `json:"source"`
	StatusCode     int       `json:"status_code"`
	RequestSummary string    `json:"request_summary,omitempty"`
	IPAddress      string    `json:"ip_address"`
	CreatedAt      time.Time `json:"created_at"`
}

// AuditLogFilter 는 감사 로그 조회 필터이다
type AuditLogFilter struct {
	UserID       string
	Action       string
	ResourceType string
	From         string
	To           string
	Page         int
	Size         int
}
