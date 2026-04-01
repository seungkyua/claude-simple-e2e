// 스토리지(볼륨, 스냅샷) 관련 API 핸들러 — OpenStack SDK를 통한 Cinder v3 API 연동
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	ossdk "github.com/kcp-cli/kcp-cli/pkg/sdk/openstack"
)

// StorageHandler 는 볼륨 및 스냅샷 관련 API를 처리한다
type StorageHandler struct {
	storage *ossdk.StorageService
}

// NewStorageHandler 는 OpenStack SDK 클라이언트를 주입받아 StorageHandler를 생성한다
func NewStorageHandler(osClient *ossdk.Client) *StorageHandler {
	return &StorageHandler{storage: ossdk.NewStorageService(osClient)}
}

// ListVolumes 는 볼륨 목록을 조회한다 (Cinder GET /volumes/detail)
func (h *StorageHandler) ListVolumes(c *gin.Context) {
	items, err := h.storage.ListVolumes()
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

// createVolumeRequest 는 볼륨 생성 요청 본문이다
type createVolumeRequest struct {
	Name        string `json:"name" binding:"required"`
	Size        int    `json:"size" binding:"required,min=1"`
	Description string `json:"description,omitempty"`
	VolumeType  string `json:"volume_type,omitempty"`
}

// CreateVolume 은 새로운 볼륨을 생성한다 (Cinder POST /volumes)
func (h *StorageHandler) CreateVolume(c *gin.Context) {
	var req createVolumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error(), "status": 400},
		})
		return
	}

	// SDK에 전달할 볼륨 본문 구성
	body := map[string]interface{}{
		"name":        req.Name,
		"size":        req.Size,
		"description": req.Description,
		"volume_type": req.VolumeType,
	}

	result, err := h.storage.CreateVolume(body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"code": "OPENSTACK_ERROR", "message": err.Error(), "status": 502},
		})
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", result)
}

// DeleteVolume 은 지정된 볼륨을 삭제한다 (Cinder DELETE /volumes/:id)
func (h *StorageHandler) DeleteVolume(c *gin.Context) {
	id := c.Param("id")
	if err := h.storage.DeleteVolume(id); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"code": "OPENSTACK_ERROR", "message": err.Error(), "status": 502},
		})
		return
	}
	c.Status(http.StatusNoContent)
}

// attachVolumeRequest 는 볼륨 연결 요청 본문이다
type attachVolumeRequest struct {
	ServerID string `json:"server_id" binding:"required"`
}

// AttachVolume 은 볼륨을 서버에 연결한다 (Cinder POST /volumes/:id/action — os-attach)
func (h *StorageHandler) AttachVolume(c *gin.Context) {
	id := c.Param("id")

	var req attachVolumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error(), "status": 400},
		})
		return
	}

	// SDK의 AttachVolume 호출 (device는 자동 할당)
	if err := h.storage.AttachVolume(id, req.ServerID, ""); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"code": "OPENSTACK_ERROR", "message": err.Error(), "status": 502},
		})
		return
	}
	c.Status(http.StatusAccepted)
}

// DetachVolume 은 서버에서 볼륨을 분리한다 (Cinder POST /volumes/:id/action — os-detach)
func (h *StorageHandler) DetachVolume(c *gin.Context) {
	id := c.Param("id")

	if err := h.storage.DetachVolume(id); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"code": "OPENSTACK_ERROR", "message": err.Error(), "status": 502},
		})
		return
	}
	c.Status(http.StatusAccepted)
}

// ListSnapshots 는 스냅샷 목록을 조회한다 (Cinder GET /snapshots/detail)
func (h *StorageHandler) ListSnapshots(c *gin.Context) {
	items, err := h.storage.ListSnapshots()
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

// createSnapshotRequest 는 스냅샷 생성 요청 본문이다
type createSnapshotRequest struct {
	Name        string `json:"name" binding:"required"`
	VolumeID    string `json:"volume_id" binding:"required"`
	Description string `json:"description,omitempty"`
}

// CreateSnapshot 은 새로운 스냅샷을 생성한다 (Cinder POST /snapshots)
func (h *StorageHandler) CreateSnapshot(c *gin.Context) {
	var req createSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error(), "status": 400},
		})
		return
	}

	// SDK에 전달할 스냅샷 본문 구성
	body := map[string]interface{}{
		"name":        req.Name,
		"volume_id":   req.VolumeID,
		"description": req.Description,
	}

	result, err := h.storage.CreateSnapshot(body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"code": "OPENSTACK_ERROR", "message": err.Error(), "status": 502},
		})
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", result)
}

// DeleteSnapshot 은 지정된 스냅샷을 삭제한다 (Cinder DELETE /snapshots/:id)
func (h *StorageHandler) DeleteSnapshot(c *gin.Context) {
	id := c.Param("id")
	if err := h.storage.DeleteSnapshot(id); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"code": "OPENSTACK_ERROR", "message": err.Error(), "status": 502},
		})
		return
	}
	c.Status(http.StatusNoContent)
}
