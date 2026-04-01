// 네트워크 관련 API 핸들러
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kcp-cli/kcp-gateway/config"
)

// NetworkHandler 는 네트워크, 서브넷, 라우터, 보안그룹 관련 API를 처리한다
type NetworkHandler struct {
	cfg *config.Config
}

// NewNetworkHandler 는 새로운 NetworkHandler를 생성한다
func NewNetworkHandler(cfg *config.Config) *NetworkHandler {
	return &NetworkHandler{cfg: cfg}
}

// notImplemented 는 미구현 상태의 공통 응답을 반환한다
func (h *NetworkHandler) notImplemented(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{"code": "NOT_IMPLEMENTED", "message": "준비 중입니다", "status": 501},
	})
}

// ListNetworks 는 네트워크 목록을 조회한다
func (h *NetworkHandler) ListNetworks(c *gin.Context) {
	h.notImplemented(c)
}

// CreateNetwork 는 새로운 네트워크를 생성한다
func (h *NetworkHandler) CreateNetwork(c *gin.Context) {
	h.notImplemented(c)
}

// DeleteNetwork 는 지정된 네트워크를 삭제한다
func (h *NetworkHandler) DeleteNetwork(c *gin.Context) {
	h.notImplemented(c)
}

// ListSubnets 는 서브넷 목록을 조회한다
func (h *NetworkHandler) ListSubnets(c *gin.Context) {
	h.notImplemented(c)
}

// CreateSubnet 은 새로운 서브넷을 생성한다
func (h *NetworkHandler) CreateSubnet(c *gin.Context) {
	h.notImplemented(c)
}

// DeleteSubnet 은 지정된 서브넷을 삭제한다
func (h *NetworkHandler) DeleteSubnet(c *gin.Context) {
	h.notImplemented(c)
}

// ListRouters 는 라우터 목록을 조회한다
func (h *NetworkHandler) ListRouters(c *gin.Context) {
	h.notImplemented(c)
}

// CreateRouter 는 새로운 라우터를 생성한다
func (h *NetworkHandler) CreateRouter(c *gin.Context) {
	h.notImplemented(c)
}

// DeleteRouter 는 지정된 라우터를 삭제한다
func (h *NetworkHandler) DeleteRouter(c *gin.Context) {
	h.notImplemented(c)
}

// AddRouterInterface 는 라우터에 인터페이스를 추가한다
func (h *NetworkHandler) AddRouterInterface(c *gin.Context) {
	h.notImplemented(c)
}

// ListSecurityGroups 는 보안그룹 목록을 조회한다
func (h *NetworkHandler) ListSecurityGroups(c *gin.Context) {
	h.notImplemented(c)
}

// CreateSecurityGroup 은 새로운 보안그룹을 생성한다
func (h *NetworkHandler) CreateSecurityGroup(c *gin.Context) {
	h.notImplemented(c)
}

// DeleteSecurityGroup 은 지정된 보안그룹을 삭제한다
func (h *NetworkHandler) DeleteSecurityGroup(c *gin.Context) {
	h.notImplemented(c)
}

// AddSecurityGroupRule 은 보안그룹에 규칙을 추가한다
func (h *NetworkHandler) AddSecurityGroupRule(c *gin.Context) {
	h.notImplemented(c)
}
