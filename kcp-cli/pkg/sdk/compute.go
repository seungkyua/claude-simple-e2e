package sdk

import "fmt"

// computeClient 는 Nova (Compute) API 클라이언트 구현체이다
type computeClient struct {
	c *Client
}

// NewComputeClient 는 새로운 Compute 클라이언트를 생성한다
func NewComputeClient(c *Client) ComputeClient {
	return &computeClient{c: c}
}

// ListServers 는 서버 목록을 조회한다
func (cc *computeClient) ListServers(opts *ListOpts) (*ListResponse[Server], error) {
	path := "/compute/servers" + buildQuery(opts)
	var resp ListResponse[Server]
	if err := cc.c.Get(path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetServer 는 특정 서버를 조회한다
func (cc *computeClient) GetServer(id string) (*Server, error) {
	var s Server
	if err := cc.c.Get(fmt.Sprintf("/compute/servers/%s", id), &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// CreateServer 는 새로운 서버를 생성한다
func (cc *computeClient) CreateServer(req *CreateServerRequest) (*Server, error) {
	var s Server
	if err := cc.c.Post("/compute/servers", req, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// DeleteServer 는 서버를 삭제한다
func (cc *computeClient) DeleteServer(id string) error {
	return cc.c.Delete(fmt.Sprintf("/compute/servers/%s", id))
}

// ServerAction 은 서버에 액션(시작, 중지, 재부팅 등)을 수행한다
func (cc *computeClient) ServerAction(id string, action string) error {
	body := map[string]string{"action": action}
	return cc.c.Post(fmt.Sprintf("/compute/servers/%s/action", id), body, nil)
}

// ListFlavors 는 Flavor 목록을 조회한다
func (cc *computeClient) ListFlavors() ([]Flavor, error) {
	var resp ListResponse[Flavor]
	if err := cc.c.Get("/compute/flavors", &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

// CreateFlavor 는 새로운 Flavor 를 생성한다
func (cc *computeClient) CreateFlavor(req *CreateFlavorRequest) (*Flavor, error) {
	var f Flavor
	if err := cc.c.Post("/compute/flavors", req, &f); err != nil {
		return nil, err
	}
	return &f, nil
}

// DeleteFlavor 는 Flavor 를 삭제한다
func (cc *computeClient) DeleteFlavor(id string) error {
	return cc.c.Delete(fmt.Sprintf("/compute/flavors/%s", id))
}
