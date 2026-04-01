// 이미지 관련 API 핸들러 — OpenStack SDK를 통한 Glance v2 API 연동
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	ossdk "github.com/kcp-cli/kcp-gateway/pkg/openstack"
)

// ImageHandler 는 이미지 관련 API를 처리한다
type ImageHandler struct {
	image *ossdk.ImageService
}

// NewImageHandler 는 OpenStack SDK 클라이언트를 주입받아 ImageHandler를 생성한다
func NewImageHandler(osClient *ossdk.Client) *ImageHandler {
	return &ImageHandler{image: ossdk.NewImageService(osClient)}
}

// ListImages 는 이미지 목록을 조회한다 (Glance GET /v2/images)
func (h *ImageHandler) ListImages(c *gin.Context) {
	items, err := h.image.ListImages()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"code": "OPENSTACK_ERROR", "message": err.Error(), "status": 502},
		})
		return
	}
	c.JSON(http.StatusOK, kcpListResponse{
		Items: items,
		Pagination: kcpPagination{
			Page:  1,
			Size:  len(items),
			Total: len(items),
		},
	})
}

// GetImage 는 특정 이미지의 상세 정보를 조회한다 (Glance GET /v2/images/:id)
func (h *ImageHandler) GetImage(c *gin.Context) {
	id := c.Param("id")
	result, err := h.image.GetImage(id)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"code": "OPENSTACK_ERROR", "message": err.Error(), "status": 502},
		})
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", result)
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
	if err := h.image.DeleteImage(id); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"code": "OPENSTACK_ERROR", "message": err.Error(), "status": 502},
		})
		return
	}
	c.Status(http.StatusNoContent)
}
