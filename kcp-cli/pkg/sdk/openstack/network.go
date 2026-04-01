package openstack

import (
	"encoding/json"
	"fmt"
)

// NetworkService 는 Neutron(Network) API 클라이언트이다
type NetworkService struct {
	c *Client
}

// NewNetworkService 는 NetworkService를 생성한다
func NewNetworkService(c *Client) *NetworkService {
	return &NetworkService{c: c}
}

// --- 네트워크 ---

// ListNetworks 는 Neutron GET /v2.0/networks 를 호출하여 네트워크 목록을 반환한다
func (s *NetworkService) ListNetworks() ([]json.RawMessage, error) {
	data, statusCode, err := s.c.DoRequest("GET", "network", "/v2.0/networks", nil)
	if err != nil {
		return nil, fmt.Errorf("네트워크 목록 조회 요청 실패: %w", err)
	}
	if err := checkStatusError(data, statusCode, "네트워크 목록 조회"); err != nil {
		return nil, err
	}
	return extractList(data, "networks")
}

// CreateNetwork 는 Neutron POST /v2.0/networks 를 호출하여 네트워크를 생성한다
func (s *NetworkService) CreateNetwork(body map[string]interface{}) (json.RawMessage, error) {
	reqBody := map[string]interface{}{"network": body}
	data, statusCode, err := s.c.DoRequest("POST", "network", "/v2.0/networks", reqBody)
	if err != nil {
		return nil, fmt.Errorf("네트워크 생성 요청 실패: %w", err)
	}
	if err := checkStatusError(data, statusCode, "네트워크 생성"); err != nil {
		return nil, err
	}
	return extractSingle(data, "network")
}

// DeleteNetwork 는 Neutron DELETE /v2.0/networks/{id} 를 호출하여 네트워크를 삭제한다
func (s *NetworkService) DeleteNetwork(id string) error {
	data, statusCode, err := s.c.DoRequest("DELETE", "network", "/v2.0/networks/"+id, nil)
	if err != nil {
		return fmt.Errorf("네트워크 삭제 요청 실패: %w", err)
	}
	return checkStatusError(data, statusCode, "네트워크 삭제")
}

// --- 서브넷 ---

// ListSubnets 는 Neutron GET /v2.0/subnets 를 호출하여 서브넷 목록을 반환한다
func (s *NetworkService) ListSubnets() ([]json.RawMessage, error) {
	data, statusCode, err := s.c.DoRequest("GET", "network", "/v2.0/subnets", nil)
	if err != nil {
		return nil, fmt.Errorf("서브넷 목록 조회 요청 실패: %w", err)
	}
	if err := checkStatusError(data, statusCode, "서브넷 목록 조회"); err != nil {
		return nil, err
	}
	return extractList(data, "subnets")
}

// CreateSubnet 은 Neutron POST /v2.0/subnets 를 호출하여 서브넷을 생성한다
func (s *NetworkService) CreateSubnet(body map[string]interface{}) (json.RawMessage, error) {
	reqBody := map[string]interface{}{"subnet": body}
	data, statusCode, err := s.c.DoRequest("POST", "network", "/v2.0/subnets", reqBody)
	if err != nil {
		return nil, fmt.Errorf("서브넷 생성 요청 실패: %w", err)
	}
	if err := checkStatusError(data, statusCode, "서브넷 생성"); err != nil {
		return nil, err
	}
	return extractSingle(data, "subnet")
}

// DeleteSubnet 은 Neutron DELETE /v2.0/subnets/{id} 를 호출하여 서브넷을 삭제한다
func (s *NetworkService) DeleteSubnet(id string) error {
	data, statusCode, err := s.c.DoRequest("DELETE", "network", "/v2.0/subnets/"+id, nil)
	if err != nil {
		return fmt.Errorf("서브넷 삭제 요청 실패: %w", err)
	}
	return checkStatusError(data, statusCode, "서브넷 삭제")
}

// --- 라우터 ---

// ListRouters 는 Neutron GET /v2.0/routers 를 호출하여 라우터 목록을 반환한다
func (s *NetworkService) ListRouters() ([]json.RawMessage, error) {
	data, statusCode, err := s.c.DoRequest("GET", "network", "/v2.0/routers", nil)
	if err != nil {
		return nil, fmt.Errorf("라우터 목록 조회 요청 실패: %w", err)
	}
	if err := checkStatusError(data, statusCode, "라우터 목록 조회"); err != nil {
		return nil, err
	}
	return extractList(data, "routers")
}

// CreateRouter 는 Neutron POST /v2.0/routers 를 호출하여 라우터를 생성한다
func (s *NetworkService) CreateRouter(body map[string]interface{}) (json.RawMessage, error) {
	reqBody := map[string]interface{}{"router": body}
	data, statusCode, err := s.c.DoRequest("POST", "network", "/v2.0/routers", reqBody)
	if err != nil {
		return nil, fmt.Errorf("라우터 생성 요청 실패: %w", err)
	}
	if err := checkStatusError(data, statusCode, "라우터 생성"); err != nil {
		return nil, err
	}
	return extractSingle(data, "router")
}

// DeleteRouter 는 Neutron DELETE /v2.0/routers/{id} 를 호출하여 라우터를 삭제한다
func (s *NetworkService) DeleteRouter(id string) error {
	data, statusCode, err := s.c.DoRequest("DELETE", "network", "/v2.0/routers/"+id, nil)
	if err != nil {
		return fmt.Errorf("라우터 삭제 요청 실패: %w", err)
	}
	return checkStatusError(data, statusCode, "라우터 삭제")
}

// AddRouterInterface 는 Neutron PUT /v2.0/routers/{id}/add_router_interface 를 호출하여
// 라우터에 인터페이스를 추가한다
func (s *NetworkService) AddRouterInterface(routerID string, body map[string]interface{}) (json.RawMessage, error) {
	data, statusCode, err := s.c.DoRequest("PUT", "network", "/v2.0/routers/"+routerID+"/add_router_interface", body)
	if err != nil {
		return nil, fmt.Errorf("라우터 인터페이스 추가 요청 실패: %w", err)
	}
	if err := checkStatusError(data, statusCode, "라우터 인터페이스 추가"); err != nil {
		return nil, err
	}
	return data, nil
}

// --- 보안 그룹 ---

// ListSecurityGroups 는 Neutron GET /v2.0/security-groups 를 호출하여 보안 그룹 목록을 반환한다
func (s *NetworkService) ListSecurityGroups() ([]json.RawMessage, error) {
	data, statusCode, err := s.c.DoRequest("GET", "network", "/v2.0/security-groups", nil)
	if err != nil {
		return nil, fmt.Errorf("보안 그룹 목록 조회 요청 실패: %w", err)
	}
	if err := checkStatusError(data, statusCode, "보안 그룹 목록 조회"); err != nil {
		return nil, err
	}
	return extractList(data, "security_groups")
}

// CreateSecurityGroup 은 Neutron POST /v2.0/security-groups 를 호출하여 보안 그룹을 생성한다
func (s *NetworkService) CreateSecurityGroup(body map[string]interface{}) (json.RawMessage, error) {
	reqBody := map[string]interface{}{"security_group": body}
	data, statusCode, err := s.c.DoRequest("POST", "network", "/v2.0/security-groups", reqBody)
	if err != nil {
		return nil, fmt.Errorf("보안 그룹 생성 요청 실패: %w", err)
	}
	if err := checkStatusError(data, statusCode, "보안 그룹 생성"); err != nil {
		return nil, err
	}
	return extractSingle(data, "security_group")
}

// DeleteSecurityGroup 은 Neutron DELETE /v2.0/security-groups/{id} 를 호출하여 보안 그룹을 삭제한다
func (s *NetworkService) DeleteSecurityGroup(id string) error {
	data, statusCode, err := s.c.DoRequest("DELETE", "network", "/v2.0/security-groups/"+id, nil)
	if err != nil {
		return fmt.Errorf("보안 그룹 삭제 요청 실패: %w", err)
	}
	return checkStatusError(data, statusCode, "보안 그룹 삭제")
}

// AddSecurityGroupRule 은 Neutron POST /v2.0/security-group-rules 를 호출하여 보안 그룹 규칙을 추가한다
func (s *NetworkService) AddSecurityGroupRule(body map[string]interface{}) (json.RawMessage, error) {
	reqBody := map[string]interface{}{"security_group_rule": body}
	data, statusCode, err := s.c.DoRequest("POST", "network", "/v2.0/security-group-rules", reqBody)
	if err != nil {
		return nil, fmt.Errorf("보안 그룹 규칙 추가 요청 실패: %w", err)
	}
	if err := checkStatusError(data, statusCode, "보안 그룹 규칙 추가"); err != nil {
		return nil, err
	}
	return extractSingle(data, "security_group_rule")
}
