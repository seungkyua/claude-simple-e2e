// 스토리지(볼륨, 스냅샷) 관련 API 핸들러 — Cinder v3 API 연동
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kcp-cli/kcp-gateway/internal/openstack"
)

// StorageHandler 는 볼륨 및 스냅샷 관련 API를 처리한다
type StorageHandler struct {
	os *openstack.Client
}

// NewStorageHandler 는 OpenStack 클라이언트를 주입받아 StorageHandler를 생성한다
func NewStorageHandler(osClient *openstack.Client) *StorageHandler {
	return &StorageHandler{os: osClient}
}

// ListVolumes 는 볼륨 목록을 조회한다 (Cinder GET /volumes/detail)
func (h *StorageHandler) ListVolumes(c *gin.Context) {
	data, status, err := h.os.DoRequest("GET", "volumev3", "/volumes/detail", nil)
	forwardOSListResponse(c, data, status, err, "volumes")
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

	// Cinder API 요청 본문 구성
	body := map[string]interface{}{
		"volume": map[string]interface{}{
			"name":        req.Name,
			"size":        req.Size,
			"description": req.Description,
			"volume_type": req.VolumeType,
		},
	}

	data, status, err := h.os.DoRequest("POST", "volumev3", "/volumes", body)
	forwardOSSingleResponse(c, data, status, err, "volume")
}

// DeleteVolume 은 지정된 볼륨을 삭제한다 (Cinder DELETE /volumes/:id)
func (h *StorageHandler) DeleteVolume(c *gin.Context) {
	id := c.Param("id")
	data, status, err := h.os.DoRequest("DELETE", "volumev3", "/volumes/"+id, nil)
	forwardOSResponse(c, data, status, err)
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

	// Cinder os-attach 액션 요청
	body := map[string]interface{}{
		"os-attach": map[string]interface{}{
			"instance_uuid": req.ServerID,
		},
	}

	data, status, err := h.os.DoRequest("POST", "volumev3", "/volumes/"+id+"/action", body)
	forwardOSResponse(c, data, status, err)
}

// DetachVolume 은 서버에서 볼륨을 분리한다 (Cinder POST /volumes/:id/action — os-detach)
func (h *StorageHandler) DetachVolume(c *gin.Context) {
	id := c.Param("id")

	// Cinder os-detach 액션 요청
	body := map[string]interface{}{
		"os-detach": map[string]interface{}{},
	}

	data, status, err := h.os.DoRequest("POST", "volumev3", "/volumes/"+id+"/action", body)
	forwardOSResponse(c, data, status, err)
}

// ListSnapshots 는 스냅샷 목록을 조회한다 (Cinder GET /snapshots/detail)
func (h *StorageHandler) ListSnapshots(c *gin.Context) {
	data, status, err := h.os.DoRequest("GET", "volumev3", "/snapshots/detail", nil)
	forwardOSListResponse(c, data, status, err, "snapshots")
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

	// Cinder 스냅샷 생성 요청 본문
	body := map[string]interface{}{
		"snapshot": map[string]interface{}{
			"name":        req.Name,
			"volume_id":   req.VolumeID,
			"description": req.Description,
		},
	}

	data, status, err := h.os.DoRequest("POST", "volumev3", "/snapshots", body)
	forwardOSSingleResponse(c, data, status, err, "snapshot")
}

// DeleteSnapshot 은 지정된 스냅샷을 삭제한다 (Cinder DELETE /snapshots/:id)
func (h *StorageHandler) DeleteSnapshot(c *gin.Context) {
	id := c.Param("id")
	data, status, err := h.os.DoRequest("DELETE", "volumev3", "/snapshots/"+id, nil)
	forwardOSResponse(c, data, status, err)
}
