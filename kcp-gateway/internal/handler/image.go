// 이미지 관련 API 핸들러
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kcp-cli/kcp-gateway/config"
)

// ImageHandler 는 이미지 관련 API를 처리한다
type ImageHandler struct {
	cfg *config.Config
}

// NewImageHandler 는 새로운 ImageHandler를 생성한다
func NewImageHandler(cfg *config.Config) *ImageHandler {
	return &ImageHandler{cfg: cfg}
}

// notImplemented 는 미구현 상태의 공통 응답을 반환한다
func (h *ImageHandler) notImplemented(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{"code": "NOT_IMPLEMENTED", "message": "준비 중입니다", "status": 501},
	})
}

// ListImages 는 이미지 목록을 조회한다
func (h *ImageHandler) ListImages(c *gin.Context) {
	h.notImplemented(c)
}

// GetImage 는 특정 이미지의 상세 정보를 조회한다
func (h *ImageHandler) GetImage(c *gin.Context) {
	h.notImplemented(c)
}

// UploadImage 는 새로운 이미지를 업로드한다
func (h *ImageHandler) UploadImage(c *gin.Context) {
	h.notImplemented(c)
}

// DeleteImage 는 지정된 이미지를 삭제한다
func (h *ImageHandler) DeleteImage(c *gin.Context) {
	h.notImplemented(c)
}
