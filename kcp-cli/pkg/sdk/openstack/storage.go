package openstack

import (
	"encoding/json"
	"fmt"
)

// StorageService 는 Cinder(Block Storage) API 클라이언트이다
type StorageService struct {
	c *Client
}

// NewStorageService 는 StorageService를 생성한다
func NewStorageService(c *Client) *StorageService {
	return &StorageService{c: c}
}

// --- 볼륨 ---

// ListVolumes 는 Cinder GET /volumes/detail 을 호출하여 볼륨 목록을 반환한다
func (s *StorageService) ListVolumes() ([]json.RawMessage, error) {
	data, statusCode, err := s.c.DoRequest("GET", "volumev3", "/volumes/detail", nil)
	if err != nil {
		return nil, fmt.Errorf("볼륨 목록 조회 요청 실패: %w", err)
	}
	if err := checkStatusError(data, statusCode, "볼륨 목록 조회"); err != nil {
		return nil, err
	}
	return extractList(data, "volumes")
}

// CreateVolume 은 Cinder POST /volumes 를 호출하여 볼륨을 생성한다
func (s *StorageService) CreateVolume(body map[string]interface{}) (json.RawMessage, error) {
	reqBody := map[string]interface{}{"volume": body}
	data, statusCode, err := s.c.DoRequest("POST", "volumev3", "/volumes", reqBody)
	if err != nil {
		return nil, fmt.Errorf("볼륨 생성 요청 실패: %w", err)
	}
	if err := checkStatusError(data, statusCode, "볼륨 생성"); err != nil {
		return nil, err
	}
	return extractSingle(data, "volume")
}

// DeleteVolume 은 Cinder DELETE /volumes/{id} 를 호출하여 볼륨을 삭제한다
func (s *StorageService) DeleteVolume(id string) error {
	data, statusCode, err := s.c.DoRequest("DELETE", "volumev3", "/volumes/"+id, nil)
	if err != nil {
		return fmt.Errorf("볼륨 삭제 요청 실패: %w", err)
	}
	return checkStatusError(data, statusCode, "볼륨 삭제")
}

// AttachVolume 은 Cinder POST /volumes/{id}/action 을 호출하여 볼륨을 서버에 연결한다
func (s *StorageService) AttachVolume(volumeID string, serverID string, device string) error {
	action := map[string]interface{}{
		"os-attach": map[string]interface{}{
			"instance_uuid": serverID,
			"mountpoint":    device,
		},
	}
	data, statusCode, err := s.c.DoRequest("POST", "volumev3", "/volumes/"+volumeID+"/action", action)
	if err != nil {
		return fmt.Errorf("볼륨 연결 요청 실패: %w", err)
	}
	return checkStatusError(data, statusCode, "볼륨 연결")
}

// DetachVolume 은 Cinder POST /volumes/{id}/action 을 호출하여 볼륨을 서버에서 분리한다
func (s *StorageService) DetachVolume(volumeID string) error {
	action := map[string]interface{}{
		"os-detach": map[string]interface{}{},
	}
	data, statusCode, err := s.c.DoRequest("POST", "volumev3", "/volumes/"+volumeID+"/action", action)
	if err != nil {
		return fmt.Errorf("볼륨 분리 요청 실패: %w", err)
	}
	return checkStatusError(data, statusCode, "볼륨 분리")
}

// --- 스냅샷 ---

// ListSnapshots 는 Cinder GET /snapshots/detail 을 호출하여 스냅샷 목록을 반환한다
func (s *StorageService) ListSnapshots() ([]json.RawMessage, error) {
	data, statusCode, err := s.c.DoRequest("GET", "volumev3", "/snapshots/detail", nil)
	if err != nil {
		return nil, fmt.Errorf("스냅샷 목록 조회 요청 실패: %w", err)
	}
	if err := checkStatusError(data, statusCode, "스냅샷 목록 조회"); err != nil {
		return nil, err
	}
	return extractList(data, "snapshots")
}

// CreateSnapshot 은 Cinder POST /snapshots 를 호출하여 스냅샷을 생성한다
func (s *StorageService) CreateSnapshot(body map[string]interface{}) (json.RawMessage, error) {
	reqBody := map[string]interface{}{"snapshot": body}
	data, statusCode, err := s.c.DoRequest("POST", "volumev3", "/snapshots", reqBody)
	if err != nil {
		return nil, fmt.Errorf("스냅샷 생성 요청 실패: %w", err)
	}
	if err := checkStatusError(data, statusCode, "스냅샷 생성"); err != nil {
		return nil, err
	}
	return extractSingle(data, "snapshot")
}

// DeleteSnapshot 은 Cinder DELETE /snapshots/{id} 를 호출하여 스냅샷을 삭제한다
func (s *StorageService) DeleteSnapshot(id string) error {
	data, statusCode, err := s.c.DoRequest("DELETE", "volumev3", "/snapshots/"+id, nil)
	if err != nil {
		return fmt.Errorf("스냅샷 삭제 요청 실패: %w", err)
	}
	return checkStatusError(data, statusCode, "스냅샷 삭제")
}
