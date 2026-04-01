// 통계/대시보드 관련 API 핸들러 — OpenStack SDK 다중 서비스 병렬 조회
package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	ossdk "github.com/kcp-cli/kcp-cli/pkg/sdk/openstack"
)

// StatsHandler 는 대시보드 통계 관련 API를 처리한다
type StatsHandler struct {
	compute  *ossdk.ComputeService
	net      *ossdk.NetworkService
	storage  *ossdk.StorageService
	image    *ossdk.ImageService
	identity *ossdk.IdentityService
	db       *sql.DB
}

// NewStatsHandler 는 OpenStack SDK 클라이언트와 DB를 주입받아 StatsHandler를 생성한다
func NewStatsHandler(osClient *ossdk.Client, db *sql.DB) *StatsHandler {
	return &StatsHandler{
		compute:  ossdk.NewComputeService(osClient),
		net:      ossdk.NewNetworkService(osClient),
		storage:  ossdk.NewStorageService(osClient),
		image:    ossdk.NewImageService(osClient),
		identity: ossdk.NewIdentityService(osClient),
		db:       db,
	}
}

// serviceCount 는 각 서비스별 리소스 개수를 저장하는 구조체이다
type serviceCount struct {
	name  string
	count int
	err   error
}

// Dashboard 는 대시보드 통계 데이터를 조회한다.
// 여러 OpenStack SDK 서비스를 병렬 호출하여 리소스 수를 집계한다.
func (h *StatsHandler) Dashboard(c *gin.Context) {
	// 병렬 조회할 서비스별 함수 정의
	type queryFunc struct {
		name string
		fn   func() ([]json.RawMessage, error)
	}

	queries := []queryFunc{
		{name: "servers", fn: h.compute.ListServers},
		{name: "flavors", fn: h.compute.ListFlavors},
		{name: "networks", fn: h.net.ListNetworks},
		{name: "subnets", fn: h.net.ListSubnets},
		{name: "routers", fn: h.net.ListRouters},
		{name: "security_groups", fn: h.net.ListSecurityGroups},
		{name: "volumes", fn: h.storage.ListVolumes},
		{name: "snapshots", fn: h.storage.ListSnapshots},
		{name: "images", fn: h.image.ListImages},
		{name: "projects", fn: h.identity.ListProjects},
		{name: "users", fn: h.identity.ListUsers},
	}

	// 병렬 호출을 위한 WaitGroup 및 채널
	var wg sync.WaitGroup
	results := make(chan serviceCount, len(queries))

	for _, q := range queries {
		wg.Add(1)
		go func(q queryFunc) {
			defer wg.Done()
			sc := serviceCount{name: q.name}

			items, err := q.fn()
			if err != nil {
				sc.err = err
				results <- sc
				return
			}

			sc.count = len(items)
			results <- sc
		}(q)
	}

	// 모든 고루틴 완료 대기 후 채널 닫기
	go func() {
		wg.Wait()
		close(results)
	}()

	// 결과 집계
	stats := make(map[string]interface{})
	var errors []string

	for sc := range results {
		if sc.err != nil {
			errors = append(errors, sc.name+": "+sc.err.Error())
			stats[sc.name] = 0
		} else {
			stats[sc.name] = sc.count
		}
	}

	response := gin.H{
		"stats": stats,
	}

	// 일부 서비스에서 오류가 발생해도 부분 결과 반환
	if len(errors) > 0 {
		response["warnings"] = errors
	}

	c.JSON(http.StatusOK, response)
}
