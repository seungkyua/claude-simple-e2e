// Package sdk 는 OpenStack API 통신을 위한 공유 라이브러리이다.
// CLI, TUI, Gateway 모두에서 사용할 수 있도록 설계되었다.
package sdk

import "time"

// Server 는 OpenStack Nova VM 인스턴스를 나타낸다
type Server struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Status    string            `json:"status"`
	Flavor    Flavor            `json:"flavor"`
	Addresses map[string][]Addr `json:"addresses"`
	Created   time.Time         `json:"created"`
	Updated   time.Time         `json:"updated"`
}

// Flavor 는 VM 사양을 나타낸다
type Flavor struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	VCPUs int    `json:"vcpus"`
	RAM   int    `json:"ram"`
	Disk  int    `json:"disk"`
}

// Addr 는 네트워크 주소를 나타낸다
type Addr struct {
	Version int    `json:"version"`
	Addr    string `json:"addr"`
	Type    string `json:"OS-EXT-IPS:type"`
}

// Network 는 OpenStack Neutron 네트워크를 나타낸다
type Network struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Status       string   `json:"status"`
	Subnets      []string `json:"subnets"`
	AdminStateUp bool     `json:"admin_state_up"`
	Shared       bool     `json:"shared"`
}

// Subnet 은 네트워크 서브넷을 나타낸다
type Subnet struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	NetworkID  string `json:"network_id"`
	CIDR       string `json:"cidr"`
	IPVersion  int    `json:"ip_version"`
	GatewayIP  string `json:"gateway_ip"`
	EnableDHCP bool   `json:"enable_dhcp"`
}

// Router 는 Neutron 라우터를 나타낸다
type Router struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Status            string `json:"status"`
	AdminStateUp      bool   `json:"admin_state_up"`
	ExternalGatewayID string `json:"external_gateway_info,omitempty"`
}

// SecurityGroup 은 보안그룹을 나타낸다
type SecurityGroup struct {
	ID    string              `json:"id"`
	Name  string              `json:"name"`
	Rules []SecurityGroupRule `json:"security_group_rules"`
}

// SecurityGroupRule 은 보안그룹 규칙을 나타낸다
type SecurityGroupRule struct {
	ID             string `json:"id"`
	Direction      string `json:"direction"`
	Protocol       string `json:"protocol"`
	PortRangeMin   *int   `json:"port_range_min"`
	PortRangeMax   *int   `json:"port_range_max"`
	RemoteIPPrefix string `json:"remote_ip_prefix"`
}

// Volume 은 OpenStack Cinder 볼륨을 나타낸다
type Volume struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	Size        int       `json:"size"`
	VolumeType  string    `json:"volume_type"`
	Description string    `json:"description"`
	Attachments []Attach  `json:"attachments"`
	Created     time.Time `json:"created_at"`
}

// Attach 는 볼륨 연결 정보를 나타낸다
type Attach struct {
	ServerID string `json:"server_id"`
	Device   string `json:"device"`
}

// Snapshot 은 볼륨 스냅샷을 나타낸다
type Snapshot struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	VolumeID    string    `json:"volume_id"`
	Size        int       `json:"size"`
	Description string    `json:"description"`
	Created     time.Time `json:"created_at"`
}

// Project 는 OpenStack Keystone 프로젝트를 나타낸다
type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	DomainID    string `json:"domain_id"`
}

// User 는 OpenStack Keystone 사용자를 나타낸다
type User struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Enabled   bool   `json:"enabled"`
	DomainID  string `json:"domain_id"`
	ProjectID string `json:"default_project_id"`
}

// Image 는 OpenStack Glance 이미지를 나타낸다
type Image struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Status        string    `json:"status"`
	DiskFormat    string    `json:"disk_format"`
	ContainerFmt  string    `json:"container_format"`
	Size          int64     `json:"size"`
	MinDisk       int       `json:"min_disk"`
	MinRAM        int       `json:"min_ram"`
	Visibility    string    `json:"visibility"`
	Created       time.Time `json:"created_at"`
}

// Pagination 은 페이지네이션 정보를 나타낸다
type Pagination struct {
	Page  int `json:"page"`
	Size  int `json:"size"`
	Total int `json:"total"`
}

// ListResponse 는 목록 조회 공통 응답 형식이다
type ListResponse[T any] struct {
	Items      []T        `json:"items"`
	Pagination Pagination `json:"pagination"`
}

// ErrorResponse 는 에러 응답 형식이다
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail 은 에러 상세 정보를 나타낸다
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"status"`
	Detail  string `json:"detail,omitempty"`
	TraceID string `json:"traceId,omitempty"`
}
