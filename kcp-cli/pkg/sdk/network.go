package sdk

import "fmt"

// networkClient 는 Neutron (Network) API 클라이언트 구현체이다
type networkClient struct {
	c *Client
}

// NewNetworkClient 는 새로운 Network 클라이언트를 생성한다
func NewNetworkClient(c *Client) NetworkClient {
	return &networkClient{c: c}
}

// ListNetworks 는 네트워크 목록을 조회한다
func (nc *networkClient) ListNetworks(opts *ListOpts) (*ListResponse[Network], error) {
	path := "/network/networks" + buildQuery(opts)
	var resp ListResponse[Network]
	if err := nc.c.Get(path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateNetwork 는 새로운 네트워크를 생성한다
func (nc *networkClient) CreateNetwork(req *CreateNetworkRequest) (*Network, error) {
	var n Network
	if err := nc.c.Post("/network/networks", req, &n); err != nil {
		return nil, err
	}
	return &n, nil
}

// DeleteNetwork 는 네트워크를 삭제한다
func (nc *networkClient) DeleteNetwork(id string) error {
	return nc.c.Delete(fmt.Sprintf("/network/networks/%s", id))
}

// ListSubnets 는 서브넷 목록을 조회한다
func (nc *networkClient) ListSubnets(opts *ListOpts) (*ListResponse[Subnet], error) {
	path := "/network/subnets" + buildQuery(opts)
	var resp ListResponse[Subnet]
	if err := nc.c.Get(path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateSubnet 은 새로운 서브넷을 생성한다
func (nc *networkClient) CreateSubnet(req *CreateSubnetRequest) (*Subnet, error) {
	var s Subnet
	if err := nc.c.Post("/network/subnets", req, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// DeleteSubnet 은 서브넷을 삭제한다
func (nc *networkClient) DeleteSubnet(id string) error {
	return nc.c.Delete(fmt.Sprintf("/network/subnets/%s", id))
}

// ListRouters 는 라우터 목록을 조회한다
func (nc *networkClient) ListRouters(opts *ListOpts) (*ListResponse[Router], error) {
	path := "/network/routers" + buildQuery(opts)
	var resp ListResponse[Router]
	if err := nc.c.Get(path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateRouter 는 새로운 라우터를 생성한다
func (nc *networkClient) CreateRouter(req *CreateRouterRequest) (*Router, error) {
	var r Router
	if err := nc.c.Post("/network/routers", req, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// DeleteRouter 는 라우터를 삭제한다
func (nc *networkClient) DeleteRouter(id string) error {
	return nc.c.Delete(fmt.Sprintf("/network/routers/%s", id))
}

// AddRouterInterface 는 라우터에 서브넷 인터페이스를 추가한다
func (nc *networkClient) AddRouterInterface(routerID, subnetID string) error {
	body := map[string]string{"subnetId": subnetID}
	return nc.c.Post(fmt.Sprintf("/network/routers/%s/add-interface", routerID), body, nil)
}

// ListSecurityGroups 는 보안그룹 목록을 조회한다
func (nc *networkClient) ListSecurityGroups(opts *ListOpts) (*ListResponse[SecurityGroup], error) {
	path := "/network/security-groups" + buildQuery(opts)
	var resp ListResponse[SecurityGroup]
	if err := nc.c.Get(path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateSecurityGroup 은 새로운 보안그룹을 생성한다
func (nc *networkClient) CreateSecurityGroup(req *CreateSecGroupRequest) (*SecurityGroup, error) {
	var sg SecurityGroup
	if err := nc.c.Post("/network/security-groups", req, &sg); err != nil {
		return nil, err
	}
	return &sg, nil
}

// DeleteSecurityGroup 은 보안그룹을 삭제한다
func (nc *networkClient) DeleteSecurityGroup(id string) error {
	return nc.c.Delete(fmt.Sprintf("/network/security-groups/%s", id))
}

// AddSecurityGroupRule 은 보안그룹에 규칙을 추가한다
func (nc *networkClient) AddSecurityGroupRule(sgID string, req *CreateSecGroupRuleRequest) (*SecurityGroupRule, error) {
	var rule SecurityGroupRule
	if err := nc.c.Post(fmt.Sprintf("/network/security-groups/%s/rules", sgID), req, &rule); err != nil {
		return nil, err
	}
	return &rule, nil
}
