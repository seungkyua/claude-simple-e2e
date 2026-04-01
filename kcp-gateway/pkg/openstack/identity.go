package openstack

import (
	"encoding/json"
	"fmt"
)

// IdentityService 는 Keystone(Identity) API 클라이언트이다
type IdentityService struct {
	c *Client
}

// NewIdentityService 는 IdentityService를 생성한다
func NewIdentityService(c *Client) *IdentityService {
	return &IdentityService{c: c}
}

// --- 프로젝트 ---

// ListProjects 는 Keystone GET /v3/projects 를 호출하여 프로젝트 목록을 반환한다
func (s *IdentityService) ListProjects() ([]json.RawMessage, error) {
	data, statusCode, err := s.c.DoRequest("GET", "identity", "/v3/projects", nil)
	if err != nil {
		return nil, fmt.Errorf("프로젝트 목록 조회 요청 실패: %w", err)
	}
	if err := checkStatusError(data, statusCode, "프로젝트 목록 조회"); err != nil {
		return nil, err
	}
	return extractList(data, "projects")
}

// CreateProject 는 Keystone POST /v3/projects 를 호출하여 프로젝트를 생성한다
func (s *IdentityService) CreateProject(body map[string]interface{}) (json.RawMessage, error) {
	reqBody := map[string]interface{}{"project": body}
	data, statusCode, err := s.c.DoRequest("POST", "identity", "/v3/projects", reqBody)
	if err != nil {
		return nil, fmt.Errorf("프로젝트 생성 요청 실패: %w", err)
	}
	if err := checkStatusError(data, statusCode, "프로젝트 생성"); err != nil {
		return nil, err
	}
	return extractSingle(data, "project")
}

// DeleteProject 는 Keystone DELETE /v3/projects/{id} 를 호출하여 프로젝트를 삭제한다
func (s *IdentityService) DeleteProject(id string) error {
	data, statusCode, err := s.c.DoRequest("DELETE", "identity", "/v3/projects/"+id, nil)
	if err != nil {
		return fmt.Errorf("프로젝트 삭제 요청 실패: %w", err)
	}
	return checkStatusError(data, statusCode, "프로젝트 삭제")
}

// --- 사용자 ---

// ListUsers 는 Keystone GET /v3/users 를 호출하여 사용자 목록을 반환한다
func (s *IdentityService) ListUsers() ([]json.RawMessage, error) {
	data, statusCode, err := s.c.DoRequest("GET", "identity", "/v3/users", nil)
	if err != nil {
		return nil, fmt.Errorf("사용자 목록 조회 요청 실패: %w", err)
	}
	if err := checkStatusError(data, statusCode, "사용자 목록 조회"); err != nil {
		return nil, err
	}
	return extractList(data, "users")
}

// CreateUser 는 Keystone POST /v3/users 를 호출하여 사용자를 생성한다
func (s *IdentityService) CreateUser(body map[string]interface{}) (json.RawMessage, error) {
	reqBody := map[string]interface{}{"user": body}
	data, statusCode, err := s.c.DoRequest("POST", "identity", "/v3/users", reqBody)
	if err != nil {
		return nil, fmt.Errorf("사용자 생성 요청 실패: %w", err)
	}
	if err := checkStatusError(data, statusCode, "사용자 생성"); err != nil {
		return nil, err
	}
	return extractSingle(data, "user")
}

// DeleteUser 는 Keystone DELETE /v3/users/{id} 를 호출하여 사용자를 삭제한다
func (s *IdentityService) DeleteUser(id string) error {
	data, statusCode, err := s.c.DoRequest("DELETE", "identity", "/v3/users/"+id, nil)
	if err != nil {
		return fmt.Errorf("사용자 삭제 요청 실패: %w", err)
	}
	return checkStatusError(data, statusCode, "사용자 삭제")
}

// --- 역할 할당 ---

// AssignRole 은 Keystone PUT /v3/projects/{pid}/users/{uid}/roles/{rid} 를 호출하여
// 프로젝트의 사용자에게 역할을 할당한다
func (s *IdentityService) AssignRole(projectID, userID, roleID string) error {
	path := fmt.Sprintf("/v3/projects/%s/users/%s/roles/%s", projectID, userID, roleID)
	data, statusCode, err := s.c.DoRequest("PUT", "identity", path, nil)
	if err != nil {
		return fmt.Errorf("역할 할당 요청 실패: %w", err)
	}
	return checkStatusError(data, statusCode, "역할 할당")
}

// RevokeRole 은 Keystone DELETE /v3/projects/{pid}/users/{uid}/roles/{rid} 를 호출하여
// 프로젝트의 사용자에서 역할을 해제한다
func (s *IdentityService) RevokeRole(projectID, userID, roleID string) error {
	path := fmt.Sprintf("/v3/projects/%s/users/%s/roles/%s", projectID, userID, roleID)
	data, statusCode, err := s.c.DoRequest("DELETE", "identity", path, nil)
	if err != nil {
		return fmt.Errorf("역할 해제 요청 실패: %w", err)
	}
	return checkStatusError(data, statusCode, "역할 해제")
}
