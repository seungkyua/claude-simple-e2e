// 인증/인가(프로젝트, 사용자, 역할) 관련 API 핸들러
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kcp-cli/kcp-gateway/config"
)

// IdentityHandler 는 프로젝트, 사용자, 역할 관련 API를 처리한다
type IdentityHandler struct {
	cfg *config.Config
}

// NewIdentityHandler 는 새로운 IdentityHandler를 생성한다
func NewIdentityHandler(cfg *config.Config) *IdentityHandler {
	return &IdentityHandler{cfg: cfg}
}

// notImplemented 는 미구현 상태의 공통 응답을 반환한다
func (h *IdentityHandler) notImplemented(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{"code": "NOT_IMPLEMENTED", "message": "준비 중입니다", "status": 501},
	})
}

// ListProjects 는 프로젝트 목록을 조회한다
func (h *IdentityHandler) ListProjects(c *gin.Context) {
	h.notImplemented(c)
}

// CreateProject 는 새로운 프로젝트를 생성한다
func (h *IdentityHandler) CreateProject(c *gin.Context) {
	h.notImplemented(c)
}

// DeleteProject 는 지정된 프로젝트를 삭제한다
func (h *IdentityHandler) DeleteProject(c *gin.Context) {
	h.notImplemented(c)
}

// ListUsers 는 사용자 목록을 조회한다
func (h *IdentityHandler) ListUsers(c *gin.Context) {
	h.notImplemented(c)
}

// CreateUser 는 새로운 사용자를 생성한다
func (h *IdentityHandler) CreateUser(c *gin.Context) {
	h.notImplemented(c)
}

// DeleteUser 는 지정된 사용자를 삭제한다
func (h *IdentityHandler) DeleteUser(c *gin.Context) {
	h.notImplemented(c)
}

// AssignRole 은 사용자에게 역할을 할당한다
func (h *IdentityHandler) AssignRole(c *gin.Context) {
	h.notImplemented(c)
}

// RevokeRole 은 사용자에게서 역할을 회수한다
func (h *IdentityHandler) RevokeRole(c *gin.Context) {
	h.notImplemented(c)
}
