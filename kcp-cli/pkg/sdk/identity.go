package sdk

import "fmt"

// identityClient 는 Keystone (Identity) API 클라이언트 구현체이다
type identityClient struct {
	c *Client
}

// NewIdentityClient 는 새로운 Identity 클라이언트를 생성한다
func NewIdentityClient(c *Client) IdentityClient {
	return &identityClient{c: c}
}

// ListProjects 는 프로젝트 목록을 조회한다
func (ic *identityClient) ListProjects(opts *ListOpts) (*ListResponse[Project], error) {
	path := "/identity/projects" + buildQuery(opts)
	var resp ListResponse[Project]
	if err := ic.c.Get(path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateProject 는 새로운 프로젝트를 생성한다
func (ic *identityClient) CreateProject(req *CreateProjectRequest) (*Project, error) {
	var p Project
	if err := ic.c.Post("/identity/projects", req, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// DeleteProject 는 프로젝트를 삭제한다
func (ic *identityClient) DeleteProject(id string) error {
	return ic.c.Delete(fmt.Sprintf("/identity/projects/%s", id))
}

// ListUsers 는 사용자 목록을 조회한다
func (ic *identityClient) ListUsers(opts *ListOpts) (*ListResponse[User], error) {
	path := "/identity/users" + buildQuery(opts)
	var resp ListResponse[User]
	if err := ic.c.Get(path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateUser 는 새로운 사용자를 생성한다
func (ic *identityClient) CreateUser(req *CreateUserRequest) (*User, error) {
	var u User
	if err := ic.c.Post("/identity/users", req, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

// DeleteUser 는 사용자를 삭제한다
func (ic *identityClient) DeleteUser(id string) error {
	return ic.c.Delete(fmt.Sprintf("/identity/users/%s", id))
}

// AssignRole 은 사용자에게 프로젝트 내 역할을 부여한다
func (ic *identityClient) AssignRole(userID, projectID, roleID string) error {
	body := map[string]string{
		"userId":    userID,
		"projectId": projectID,
		"roleId":    roleID,
	}
	return ic.c.Post("/identity/roles/assign", body, nil)
}

// RevokeRole 은 사용자에게서 프로젝트 내 역할을 회수한다
func (ic *identityClient) RevokeRole(userID, projectID, roleID string) error {
	body := map[string]string{
		"userId":    userID,
		"projectId": projectID,
		"roleId":    roleID,
	}
	return ic.c.Post("/identity/roles/revoke", body, nil)
}
