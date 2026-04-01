// 컴퓨트(서버/인스턴스) 관련 API 핸들러
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kcp-cli/kcp-gateway/config"
)

// ComputeHandler 는 컴퓨트(서버, 플레이버) 관련 API를 처리한다
type ComputeHandler struct {
	cfg *config.Config
}

// NewComputeHandler 는 새로운 ComputeHandler를 생성한다
func NewComputeHandler(cfg *config.Config) *ComputeHandler {
	return &ComputeHandler{cfg: cfg}
}

// notImplemented 는 미구현 상태의 공통 응답을 반환한다
func (h *ComputeHandler) notImplemented(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{"code": "NOT_IMPLEMENTED", "message": "준비 중입니다", "status": 501},
	})
}

// ListServers 는 서버 목록을 조회한다
func (h *ComputeHandler) ListServers(c *gin.Context) {
	h.notImplemented(c)
}

// GetServer 는 특정 서버의 상세 정보를 조회한다
func (h *ComputeHandler) GetServer(c *gin.Context) {
	h.notImplemented(c)
}

// CreateServer 는 새로운 서버를 생성한다
func (h *ComputeHandler) CreateServer(c *gin.Context) {
	h.notImplemented(c)
}

// DeleteServer 는 지정된 서버를 삭제한다
func (h *ComputeHandler) DeleteServer(c *gin.Context) {
	h.notImplemented(c)
}

// ServerAction 은 서버에 대한 액션(시작, 중지, 재부팅 등)을 수행한다
func (h *ComputeHandler) ServerAction(c *gin.Context) {
	h.notImplemented(c)
}

// ListFlavors 는 플레이버(인스턴스 사양) 목록을 조회한다
func (h *ComputeHandler) ListFlavors(c *gin.Context) {
	h.notImplemented(c)
}

// CreateFlavor 는 새로운 플레이버를 생성한다
func (h *ComputeHandler) CreateFlavor(c *gin.Context) {
	h.notImplemented(c)
}

// DeleteFlavor 는 지정된 플레이버를 삭제한다
func (h *ComputeHandler) DeleteFlavor(c *gin.Context) {
	h.notImplemented(c)
}
