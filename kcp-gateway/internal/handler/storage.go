// 스토리지(볼륨, 스냅샷) 관련 API 핸들러
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kcp-cli/kcp-gateway/config"
)

// StorageHandler 는 볼륨 및 스냅샷 관련 API를 처리한다
type StorageHandler struct {
	cfg *config.Config
}

// NewStorageHandler 는 새로운 StorageHandler를 생성한다
func NewStorageHandler(cfg *config.Config) *StorageHandler {
	return &StorageHandler{cfg: cfg}
}

// notImplemented 는 미구현 상태의 공통 응답을 반환한다
func (h *StorageHandler) notImplemented(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{"code": "NOT_IMPLEMENTED", "message": "준비 중입니다", "status": 501},
	})
}

// ListVolumes 는 볼륨 목록을 조회한다
func (h *StorageHandler) ListVolumes(c *gin.Context) {
	h.notImplemented(c)
}

// CreateVolume 은 새로운 볼륨을 생성한다
func (h *StorageHandler) CreateVolume(c *gin.Context) {
	h.notImplemented(c)
}

// DeleteVolume 은 지정된 볼륨을 삭제한다
func (h *StorageHandler) DeleteVolume(c *gin.Context) {
	h.notImplemented(c)
}

// AttachVolume 은 볼륨을 서버에 연결한다
func (h *StorageHandler) AttachVolume(c *gin.Context) {
	h.notImplemented(c)
}

// DetachVolume 은 서버에서 볼륨을 분리한다
func (h *StorageHandler) DetachVolume(c *gin.Context) {
	h.notImplemented(c)
}

// ListSnapshots 는 스냅샷 목록을 조회한다
func (h *StorageHandler) ListSnapshots(c *gin.Context) {
	h.notImplemented(c)
}

// CreateSnapshot 은 새로운 스냅샷을 생성한다
func (h *StorageHandler) CreateSnapshot(c *gin.Context) {
	h.notImplemented(c)
}

// DeleteSnapshot 은 지정된 스냅샷을 삭제한다
func (h *StorageHandler) DeleteSnapshot(c *gin.Context) {
	h.notImplemented(c)
}
