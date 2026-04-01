package openstack

import (
	"encoding/json"
	"fmt"
)

// ComputeService 는 Nova(Compute) API 클라이언트이다
type ComputeService struct {
	c *Client
}

// NewComputeService 는 ComputeService를 생성한다
func NewComputeService(c *Client) *ComputeService {
	return &ComputeService{c: c}
}

// ListServers 는 Nova GET /servers/detail 을 호출하여 서버 목록을 반환한다
func (s *ComputeService) ListServers() ([]json.RawMessage, error) {
	data, statusCode, err := s.c.DoRequest("GET", "compute", "/servers/detail", nil)
	if err != nil {
		return nil, fmt.Errorf("서버 목록 조회 요청 실패: %w", err)
	}
	if err := checkStatusError(data, statusCode, "서버 목록 조회"); err != nil {
		return nil, err
	}
	return extractList(data, "servers")
}

// GetServer 는 Nova GET /servers/{id} 를 호출하여 단일 서버 정보를 반환한다
func (s *ComputeService) GetServer(id string) (json.RawMessage, error) {
	data, statusCode, err := s.c.DoRequest("GET", "compute", "/servers/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("서버 조회 요청 실패: %w", err)
	}
	if err := checkStatusError(data, statusCode, "서버 조회"); err != nil {
		return nil, err
	}
	return extractSingle(data, "server")
}

// CreateServer 는 Nova POST /servers 를 호출하여 서버를 생성한다
func (s *ComputeService) CreateServer(body map[string]interface{}) (json.RawMessage, error) {
	reqBody := map[string]interface{}{"server": body}
	data, statusCode, err := s.c.DoRequest("POST", "compute", "/servers", reqBody)
	if err != nil {
		return nil, fmt.Errorf("서버 생성 요청 실패: %w", err)
	}
	if err := checkStatusError(data, statusCode, "서버 생성"); err != nil {
		return nil, err
	}
	return extractSingle(data, "server")
}

// DeleteServer 는 Nova DELETE /servers/{id} 를 호출하여 서버를 삭제한다
func (s *ComputeService) DeleteServer(id string) error {
	data, statusCode, err := s.c.DoRequest("DELETE", "compute", "/servers/"+id, nil)
	if err != nil {
		return fmt.Errorf("서버 삭제 요청 실패: %w", err)
	}
	return checkStatusError(data, statusCode, "서버 삭제")
}

// ServerAction 은 Nova POST /servers/{id}/action 을 호출하여 서버 액션을 수행한다
// 예: 재부팅, 일시정지, 재개 등
func (s *ComputeService) ServerAction(id string, action map[string]interface{}) error {
	data, statusCode, err := s.c.DoRequest("POST", "compute", "/servers/"+id+"/action", action)
	if err != nil {
		return fmt.Errorf("서버 액션 요청 실패: %w", err)
	}
	return checkStatusError(data, statusCode, "서버 액션")
}

// ListFlavors 는 Nova GET /flavors/detail 을 호출하여 Flavor 목록을 반환한다
func (s *ComputeService) ListFlavors() ([]json.RawMessage, error) {
	data, statusCode, err := s.c.DoRequest("GET", "compute", "/flavors/detail", nil)
	if err != nil {
		return nil, fmt.Errorf("Flavor 목록 조회 요청 실패: %w", err)
	}
	if err := checkStatusError(data, statusCode, "Flavor 목록 조회"); err != nil {
		return nil, err
	}
	return extractList(data, "flavors")
}

// CreateFlavor 는 Nova POST /flavors 를 호출하여 Flavor를 생성한다
func (s *ComputeService) CreateFlavor(body map[string]interface{}) (json.RawMessage, error) {
	reqBody := map[string]interface{}{"flavor": body}
	data, statusCode, err := s.c.DoRequest("POST", "compute", "/flavors", reqBody)
	if err != nil {
		return nil, fmt.Errorf("Flavor 생성 요청 실패: %w", err)
	}
	if err := checkStatusError(data, statusCode, "Flavor 생성"); err != nil {
		return nil, err
	}
	return extractSingle(data, "flavor")
}

// DeleteFlavor 는 Nova DELETE /flavors/{id} 를 호출하여 Flavor를 삭제한다
func (s *ComputeService) DeleteFlavor(id string) error {
	data, statusCode, err := s.c.DoRequest("DELETE", "compute", "/flavors/"+id, nil)
	if err != nil {
		return fmt.Errorf("Flavor 삭제 요청 실패: %w", err)
	}
	return checkStatusError(data, statusCode, "Flavor 삭제")
}
