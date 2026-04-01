// 이미지 관련 API 핸들러 — Glance v2 API 연동
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kcp-cli/kcp-gateway/internal/openstack"
)

// ImageHandler 는 이미지 관련 API를 처리한다
type ImageHandler struct {
	os *openstack.Client
}

// NewImageHandler 는 OpenStack 클라이언트를 주입받아 ImageHandler를 생성한다
func NewImageHandler(osClient *openstack.Client) *ImageHandler {
	return &ImageHandler{os: osClient}
}

// ListImages 는 이미지 목록을 조회한다 (Glance GET /v2/images)
func (h *ImageHandler) ListImages(c *gin.Context) {
	data, status, err := h.os.DoRequest("GET", "image", "/v2/images", nil)
	forwardOSListResponse(c, data, status, err, "images")
}

// GetImage 는 특정 이미지의 상세 정보를 조회한다 (Glance GET /v2/images/:id)
func (h *ImageHandler) GetImage(c *gin.Context) {
	id := c.Param("id")
	data, status, err := h.os.DoRequest("GET", "image", "/v2/images/"+id, nil)
	forwardOSResponse(c, data, status, err)
}

// UploadImage 는 이미지 업로드 — 멀티파트 업로드가 필요하므로 별도 구현 예정
func (h *ImageHandler) UploadImage(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "이미지 업로드는 별도 구현이 필요합니다",
			"status":  501,
		},
	})
}

// DeleteImage 는 지정된 이미지를 삭제한다 (Glance DELETE /v2/images/:id)
func (h *ImageHandler) DeleteImage(c *gin.Context) {
	id := c.Param("id")
	data, status, err := h.os.DoRequest("DELETE", "image", "/v2/images/"+id, nil)
	forwardOSResponse(c, data, status, err)
}
