package sdk

import "fmt"

// storageClient 는 Cinder (Storage) API 클라이언트 구현체이다
type storageClient struct {
	c *Client
}

// NewStorageClient 는 새로운 Storage 클라이언트를 생성한다
func NewStorageClient(c *Client) StorageClient {
	return &storageClient{c: c}
}

// ListVolumes 는 볼륨 목록을 조회한다
func (sc *storageClient) ListVolumes(opts *ListOpts) (*ListResponse[Volume], error) {
	path := "/storage/volumes" + buildQuery(opts)
	var resp ListResponse[Volume]
	if err := sc.c.Get(path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateVolume 은 새로운 볼륨을 생성한다
func (sc *storageClient) CreateVolume(req *CreateVolumeRequest) (*Volume, error) {
	var v Volume
	if err := sc.c.Post("/storage/volumes", req, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// DeleteVolume 은 볼륨을 삭제한다
func (sc *storageClient) DeleteVolume(id string) error {
	return sc.c.Delete(fmt.Sprintf("/storage/volumes/%s", id))
}

// AttachVolume 은 볼륨을 서버에 연결한다
func (sc *storageClient) AttachVolume(volumeID, serverID string) error {
	body := map[string]string{"serverId": serverID}
	return sc.c.Post(fmt.Sprintf("/storage/volumes/%s/attach", volumeID), body, nil)
}

// DetachVolume 은 볼륨을 서버에서 분리한다
func (sc *storageClient) DetachVolume(volumeID string) error {
	return sc.c.Post(fmt.Sprintf("/storage/volumes/%s/detach", volumeID), nil, nil)
}

// ListSnapshots 는 스냅샷 목록을 조회한다
func (sc *storageClient) ListSnapshots(opts *ListOpts) (*ListResponse[Snapshot], error) {
	path := "/storage/snapshots" + buildQuery(opts)
	var resp ListResponse[Snapshot]
	if err := sc.c.Get(path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateSnapshot 은 새로운 스냅샷을 생성한다
func (sc *storageClient) CreateSnapshot(req *CreateSnapshotRequest) (*Snapshot, error) {
	var s Snapshot
	if err := sc.c.Post("/storage/snapshots", req, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// DeleteSnapshot 은 스냅샷을 삭제한다
func (sc *storageClient) DeleteSnapshot(id string) error {
	return sc.c.Delete(fmt.Sprintf("/storage/snapshots/%s", id))
}
