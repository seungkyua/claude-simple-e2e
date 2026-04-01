// 인증/인가(프로젝트, 사용자, 역할) 관련 API 핸들러 — Keystone v3 API 연동
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kcp-cli/kcp-gateway/internal/openstack"
)

// IdentityHandler 는 프로젝트, 사용자, 역할 관련 API를 처리한다
type IdentityHandler struct {
	os *openstack.Client
}

// NewIdentityHandler 는 OpenStack 클라이언트를 주입받아 IdentityHandler를 생성한다
func NewIdentityHandler(osClient *openstack.Client) *IdentityHandler {
	return &IdentityHandler{os: osClient}
}

// ListProjects 는 프로젝트 목록을 조회한다 (Keystone GET /projects)
func (h *IdentityHandler) ListProjects(c *gin.Context) {
	data, status, err := h.os.DoRequest("GET", "identity", "/v3/projects", nil)
	forwardOSListResponse(c, data, status, err, "projects")
}

// createProjectRequest 는 프로젝트 생성 요청 본문이다
type createProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description,omitempty"`
	DomainID    string `json:"domain_id,omitempty"`
	Enabled     *bool  `json:"enabled,omitempty"`
}

// CreateProject 는 새로운 프로젝트를 생성한다 (Keystone POST /projects)
func (h *IdentityHandler) CreateProject(c *gin.Context) {
	var req createProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error(), "status": 400},
		})
		return
	}

	// Keystone 프로젝트 생성 요청 본문
	project := map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
	}
	if req.DomainID != "" {
		project["domain_id"] = req.DomainID
	}
	if req.Enabled != nil {
		project["enabled"] = *req.Enabled
	}

	body := map[string]interface{}{"project": project}
	data, status, err := h.os.DoRequest("POST", "identity", "/v3/projects", body)
	forwardOSSingleResponse(c, data, status, err, "project")
}

// DeleteProject 는 지정된 프로젝트를 삭제한다 (Keystone DELETE /projects/:id)
func (h *IdentityHandler) DeleteProject(c *gin.Context) {
	id := c.Param("id")
	data, status, err := h.os.DoRequest("DELETE", "identity", "/projects/"+id, nil)
	forwardOSResponse(c, data, status, err)
}

// ListUsers 는 사용자 목록을 조회한다 (Keystone GET /users)
func (h *IdentityHandler) ListUsers(c *gin.Context) {
	data, status, err := h.os.DoRequest("GET", "identity", "/v3/users", nil)
	forwardOSListResponse(c, data, status, err, "users")
}

// createUserRequest 는 사용자 생성 요청 본문이다
type createUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
	DomainID string `json:"domain_id,omitempty"`
	Email    string `json:"email,omitempty"`
	Enabled  *bool  `json:"enabled,omitempty"`
}

// CreateUser 는 새로운 사용자를 생성한다 (Keystone POST /users)
func (h *IdentityHandler) CreateUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error(), "status": 400},
		})
		return
	}

	// Keystone 사용자 생성 요청 본문
	user := map[string]interface{}{
		"name":     req.Name,
		"password": req.Password,
	}
	if req.DomainID != "" {
		user["domain_id"] = req.DomainID
	}
	if req.Email != "" {
		user["email"] = req.Email
	}
	if req.Enabled != nil {
		user["enabled"] = *req.Enabled
	}

	body := map[string]interface{}{"user": user}
	data, status, err := h.os.DoRequest("POST", "identity", "/v3/users", body)
	forwardOSSingleResponse(c, data, status, err, "user")
}

// DeleteUser 는 지정된 사용자를 삭제한다 (Keystone DELETE /users/:id)
func (h *IdentityHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	data, status, err := h.os.DoRequest("DELETE", "identity", "/users/"+id, nil)
	forwardOSResponse(c, data, status, err)
}

// assignRoleRequest 는 역할 할당/회수 요청 본문이다
type assignRoleRequest struct {
	UserID    string `json:"user_id" binding:"required"`
	ProjectID string `json:"project_id" binding:"required"`
	RoleID    string `json:"role_id" binding:"required"`
}

// AssignRole 은 사용자에게 역할을 할당한다
// Keystone PUT /projects/{projectId}/users/{userId}/roles/{roleId}
func (h *IdentityHandler) AssignRole(c *gin.Context) {
	var req assignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error(), "status": 400},
		})
		return
	}

	path := "/v3/projects/" + req.ProjectID + "/users/" + req.UserID + "/roles/" + req.RoleID
	data, status, err := h.os.DoRequest("PUT", "identity", path, nil)
	forwardOSResponse(c, data, status, err)
}

// RevokeRole 은 사용자에게서 역할을 회수한다
// Keystone DELETE /projects/{projectId}/users/{userId}/roles/{roleId}
func (h *IdentityHandler) RevokeRole(c *gin.Context) {
	var req assignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error(), "status": 400},
		})
		return
	}

	path := "/v3/projects/" + req.ProjectID + "/users/" + req.UserID + "/roles/" + req.RoleID
	data, status, err := h.os.DoRequest("DELETE", "identity", path, nil)
	forwardOSResponse(c, data, status, err)
}
