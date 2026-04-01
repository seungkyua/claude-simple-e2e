package openstack

import (
	"encoding/json"
	"fmt"
)

// ImageService 는 Glance(Image) API 클라이언트이다
type ImageService struct {
	c *Client
}

// NewImageService 는 ImageService를 생성한다
func NewImageService(c *Client) *ImageService {
	return &ImageService{c: c}
}

// ListImages 는 Glance GET /v2/images 를 호출하여 이미지 목록을 반환한다
func (s *ImageService) ListImages() ([]json.RawMessage, error) {
	data, statusCode, err := s.c.DoRequest("GET", "image", "/v2/images", nil)
	if err != nil {
		return nil, fmt.Errorf("이미지 목록 조회 요청 실패: %w", err)
	}
	if err := checkStatusError(data, statusCode, "이미지 목록 조회"); err != nil {
		return nil, err
	}
	return extractList(data, "images")
}

// GetImage 는 Glance GET /v2/images/{id} 를 호출하여 단일 이미지 정보를 반환한다.
// Glance v2 API는 단일 이미지 조회 시 래퍼 키 없이 직접 객체를 반환한다.
func (s *ImageService) GetImage(id string) (json.RawMessage, error) {
	data, statusCode, err := s.c.DoRequest("GET", "image", "/v2/images/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("이미지 조회 요청 실패: %w", err)
	}
	if err := checkStatusError(data, statusCode, "이미지 조회"); err != nil {
		return nil, err
	}
	// Glance v2는 단일 이미지를 래퍼 없이 직접 반환한다
	return json.RawMessage(data), nil
}

// DeleteImage 는 Glance DELETE /v2/images/{id} 를 호출하여 이미지를 삭제한다
func (s *ImageService) DeleteImage(id string) error {
	data, statusCode, err := s.c.DoRequest("DELETE", "image", "/v2/images/"+id, nil)
	if err != nil {
		return fmt.Errorf("이미지 삭제 요청 실패: %w", err)
	}
	return checkStatusError(data, statusCode, "이미지 삭제")
}
