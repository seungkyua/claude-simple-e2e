package sdk

// ComputeClient 는 Nova (Compute) API 클라이언트 인터페이스이다
type ComputeClient interface {
	ListServers(opts *ListOpts) (*ListResponse[Server], error)
	GetServer(id string) (*Server, error)
	CreateServer(req *CreateServerRequest) (*Server, error)
	DeleteServer(id string) error
	ServerAction(id string, action string) error
	ListFlavors() ([]Flavor, error)
	CreateFlavor(req *CreateFlavorRequest) (*Flavor, error)
	DeleteFlavor(id string) error
}

// NetworkClient 는 Neutron (Network) API 클라이언트 인터페이스이다
type NetworkClient interface {
	ListNetworks(opts *ListOpts) (*ListResponse[Network], error)
	CreateNetwork(req *CreateNetworkRequest) (*Network, error)
	DeleteNetwork(id string) error
	ListSubnets(opts *ListOpts) (*ListResponse[Subnet], error)
	CreateSubnet(req *CreateSubnetRequest) (*Subnet, error)
	DeleteSubnet(id string) error
	ListRouters(opts *ListOpts) (*ListResponse[Router], error)
	CreateRouter(req *CreateRouterRequest) (*Router, error)
	DeleteRouter(id string) error
	AddRouterInterface(routerID, subnetID string) error
	ListSecurityGroups(opts *ListOpts) (*ListResponse[SecurityGroup], error)
	CreateSecurityGroup(req *CreateSecGroupRequest) (*SecurityGroup, error)
	DeleteSecurityGroup(id string) error
	AddSecurityGroupRule(sgID string, req *CreateSecGroupRuleRequest) (*SecurityGroupRule, error)
}

// StorageClient 는 Cinder (Storage) API 클라이언트 인터페이스이다
type StorageClient interface {
	ListVolumes(opts *ListOpts) (*ListResponse[Volume], error)
	CreateVolume(req *CreateVolumeRequest) (*Volume, error)
	DeleteVolume(id string) error
	AttachVolume(volumeID, serverID string) error
	DetachVolume(volumeID string) error
	ListSnapshots(opts *ListOpts) (*ListResponse[Snapshot], error)
	CreateSnapshot(req *CreateSnapshotRequest) (*Snapshot, error)
	DeleteSnapshot(id string) error
}

// IdentityClient 는 Keystone (Identity) API 클라이언트 인터페이스이다
type IdentityClient interface {
	ListProjects(opts *ListOpts) (*ListResponse[Project], error)
	CreateProject(req *CreateProjectRequest) (*Project, error)
	DeleteProject(id string) error
	ListUsers(opts *ListOpts) (*ListResponse[User], error)
	CreateUser(req *CreateUserRequest) (*User, error)
	DeleteUser(id string) error
	AssignRole(userID, projectID, roleID string) error
	RevokeRole(userID, projectID, roleID string) error
}

// ImageClient 는 Glance (Image) API 클라이언트 인터페이스이다
type ImageClient interface {
	ListImages(opts *ListOpts) (*ListResponse[Image], error)
	GetImage(id string) (*Image, error)
	DeleteImage(id string) error
}

// ListOpts 는 목록 조회 공통 옵션이다
type ListOpts struct {
	Page   int               `json:"page,omitempty"`
	Size   int               `json:"size,omitempty"`
	Filter map[string]string `json:"filter,omitempty"`
}

// --- 요청 구조체 ---

// CreateServerRequest 는 VM 생성 요청이다
type CreateServerRequest struct {
	Name             string   `json:"name"`
	FlavorID         string   `json:"flavorId"`
	ImageID          string   `json:"imageId"`
	NetworkIDs       []string `json:"networkIds,omitempty"`
	SecurityGroupIDs []string `json:"securityGroupIds,omitempty"`
	KeyName          string   `json:"keyName,omitempty"`
	UserData         string   `json:"userData,omitempty"`
}

// CreateFlavorRequest 는 Flavor 생성 요청이다
type CreateFlavorRequest struct {
	Name  string `json:"name"`
	VCPUs int    `json:"vcpus"`
	RAM   int    `json:"ram"`
	Disk  int    `json:"disk"`
}

// CreateNetworkRequest 는 네트워크 생성 요청이다
type CreateNetworkRequest struct {
	Name         string `json:"name"`
	AdminStateUp bool   `json:"adminStateUp"`
	Shared       bool   `json:"shared"`
}

// CreateSubnetRequest 는 서브넷 생성 요청이다
type CreateSubnetRequest struct {
	Name       string `json:"name"`
	NetworkID  string `json:"networkId"`
	CIDR       string `json:"cidr"`
	IPVersion  int    `json:"ipVersion"`
	GatewayIP  string `json:"gatewayIp,omitempty"`
	EnableDHCP bool   `json:"enableDhcp"`
}

// CreateRouterRequest 는 라우터 생성 요청이다
type CreateRouterRequest struct {
	Name              string `json:"name"`
	AdminStateUp      bool   `json:"adminStateUp"`
	ExternalGatewayID string `json:"externalGatewayId,omitempty"`
}

// CreateSecGroupRequest 는 보안그룹 생성 요청이다
type CreateSecGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// CreateSecGroupRuleRequest 는 보안그룹 규칙 생성 요청이다
type CreateSecGroupRuleRequest struct {
	Direction      string `json:"direction"`
	Protocol       string `json:"protocol"`
	PortRangeMin   *int   `json:"portRangeMin,omitempty"`
	PortRangeMax   *int   `json:"portRangeMax,omitempty"`
	RemoteIPPrefix string `json:"remoteIpPrefix,omitempty"`
}

// CreateVolumeRequest 는 볼륨 생성 요청이다
type CreateVolumeRequest struct {
	Name        string `json:"name"`
	Size        int    `json:"size"`
	VolumeType  string `json:"volumeType,omitempty"`
	Description string `json:"description,omitempty"`
}

// CreateSnapshotRequest 는 스냅샷 생성 요청이다
type CreateSnapshotRequest struct {
	Name        string `json:"name"`
	VolumeID    string `json:"volumeId"`
	Description string `json:"description,omitempty"`
}

// CreateProjectRequest 는 프로젝트 생성 요청이다
type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	DomainID    string `json:"domainId,omitempty"`
}

// CreateUserRequest 는 사용자 생성 요청이다
type CreateUserRequest struct {
	Name      string `json:"name"`
	Email     string `json:"email,omitempty"`
	Password  string `json:"password"`
	DomainID  string `json:"domainId,omitempty"`
	ProjectID string `json:"projectId,omitempty"`
}
