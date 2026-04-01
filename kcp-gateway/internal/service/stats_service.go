package service

// StatsService 는 대시보드 통계 비즈니스 로직을 처리한다
type StatsService struct{}

// NewStatsService 는 새로운 StatsService를 생성한다
func NewStatsService() *StatsService {
	return &StatsService{}
}

// DashboardStats 는 대시보드 통계 데이터이다
type DashboardStats struct {
	Compute  ComputeStats  `json:"compute"`
	Network  NetworkStats  `json:"network"`
	Storage  StorageStats  `json:"storage"`
	Identity IdentityStats `json:"identity"`
	Image    ImageStats    `json:"image"`
}

// ComputeStats 는 Compute 리소스 통계이다
type ComputeStats struct {
	Total   int `json:"total"`
	Active  int `json:"active"`
	Shutoff int `json:"shutoff"`
	Error   int `json:"error"`
}

// NetworkStats 는 Network 리소스 통계이다
type NetworkStats struct {
	Networks int `json:"networks"`
	Subnets  int `json:"subnets"`
	Routers  int `json:"routers"`
}

// StorageStats 는 Storage 리소스 통계이다
type StorageStats struct {
	Volumes     int `json:"volumes"`
	TotalSizeGB int `json:"totalSizeGB"`
	Snapshots   int `json:"snapshots"`
}

// IdentityStats 는 Identity 리소스 통계이다
type IdentityStats struct {
	Projects int `json:"projects"`
	Users    int `json:"users"`
}

// ImageStats 는 Image 리소스 통계이다
type ImageStats struct {
	Total int `json:"total"`
}
