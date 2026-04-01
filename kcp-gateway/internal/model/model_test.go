package model

import (
	"testing"
	"time"
)

// TestUserStructInstantiation 는 User 구조체의 필드가 올바르게 설정되는지 검증한다
func TestUserStructInstantiation(t *testing.T) {
	now := time.Now()
	user := User{
		ID:           "user-001",
		Username:     "admin",
		PasswordHash: "$2a$12$hashedvalue",
		Email:        "admin@example.com",
		Role:         "ADMIN",
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if user.ID != "user-001" {
		t.Errorf("ID = %q, 기대값 %q", user.ID, "user-001")
	}
	if user.Username != "admin" {
		t.Errorf("Username = %q, 기대값 %q", user.Username, "admin")
	}
	if user.Role != "ADMIN" {
		t.Errorf("Role = %q, 기대값 %q", user.Role, "ADMIN")
	}
	if !user.IsActive {
		t.Error("IsActive = false, 기대값 true")
	}
}

// TestSessionStructInstantiation 는 Session 구조체의 필드가 올바르게 설정되는지 검증한다
func TestSessionStructInstantiation(t *testing.T) {
	expires := time.Now().Add(15 * time.Minute)
	session := Session{
		ID:        "sess-001",
		UserID:    "user-001",
		Token:     "jwt-token-string",
		AuthType:  "JWT",
		IPAddress: "192.168.1.1",
		ExpiresAt: expires,
	}

	if session.UserID != "user-001" {
		t.Errorf("UserID = %q, 기대값 %q", session.UserID, "user-001")
	}
	if session.AuthType != "JWT" {
		t.Errorf("AuthType = %q, 기대값 %q", session.AuthType, "JWT")
	}
	if session.ExpiresAt != expires {
		t.Errorf("ExpiresAt 불일치")
	}
}

// TestAuditLogStructInstantiation 는 AuditLog 구조체의 필드가 올바르게 설정되는지 검증한다
func TestAuditLogStructInstantiation(t *testing.T) {
	log := AuditLog{
		ID:             "log-001",
		UserID:         "user-001",
		Username:       "admin",
		Action:         "CREATE",
		ResourceType:   "VM",
		ResourceID:     "vm-123",
		Source:         "CLI",
		StatusCode:     200,
		RequestSummary: "POST /api/v1/compute/servers",
		IPAddress:      "10.0.0.1",
		CreatedAt:      time.Now(),
	}

	if log.Action != "CREATE" {
		t.Errorf("Action = %q, 기대값 %q", log.Action, "CREATE")
	}
	if log.ResourceType != "VM" {
		t.Errorf("ResourceType = %q, 기대값 %q", log.ResourceType, "VM")
	}
	if log.StatusCode != 200 {
		t.Errorf("StatusCode = %d, 기대값 %d", log.StatusCode, 200)
	}
}

// TestAuditLogFilterStructInstantiation 는 AuditLogFilter 구조체의 필드가 올바르게 설정되는지 검증한다
func TestAuditLogFilterStructInstantiation(t *testing.T) {
	filter := AuditLogFilter{
		UserID:       "user-001",
		Action:       "READ",
		ResourceType: "VM",
		From:         "2026-01-01",
		To:           "2026-12-31",
		Page:         1,
		Size:         20,
	}

	if filter.Page != 1 {
		t.Errorf("Page = %d, 기대값 %d", filter.Page, 1)
	}
	if filter.Size != 20 {
		t.Errorf("Size = %d, 기대값 %d", filter.Size, 20)
	}
}
