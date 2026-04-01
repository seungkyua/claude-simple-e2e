// 통계/대시보드 관련 API 핸들러 — OpenStack 다중 서비스 병렬 조회
package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kcp-cli/kcp-gateway/internal/openstack"
)

// StatsHandler 는 대시보드 통계 관련 API를 처리한다
type StatsHandler struct {
	os *openstack.Client
	db *sql.DB
}

// NewStatsHandler 는 OpenStack 클라이언트와 DB를 주입받아 StatsHandler를 생성한다
func NewStatsHandler(osClient *openstack.Client, db *sql.DB) *StatsHandler {
	return &StatsHandler{os: osClient, db: db}
}

// serviceCount 는 각 서비스별 리소스 개수를 저장하는 구조체이다
type serviceCount struct {
	name  string
	count int
	err   error
}

// Dashboard 는 대시보드 통계 데이터를 조회한다.
// 여러 OpenStack 서비스 API를 병렬 호출하여 리소스 수를 집계한다.
func (h *StatsHandler) Dashboard(c *gin.Context) {
	// 병렬 조회할 서비스 목록 정의
	type query struct {
		name        string
		serviceType string
		path        string
		countKey    string // JSON 응답에서 배열 키
	}

	queries := []query{
		{name: "servers", serviceType: "compute", path: "/servers", countKey: "servers"},
		{name: "networks", serviceType: "network", path: "/v2.0/networks", countKey: "networks"},
		{name: "subnets", serviceType: "network", path: "/v2.0/subnets", countKey: "subnets"},
		{name: "routers", serviceType: "network", path: "/v2.0/routers", countKey: "routers"},
		{name: "volumes", serviceType: "volumev3", path: "/volumes", countKey: "volumes"},
		{name: "images", serviceType: "image", path: "/v2/images", countKey: "images"},
		{name: "projects", serviceType: "identity", path: "/projects", countKey: "projects"},
		{name: "users", serviceType: "identity", path: "/users", countKey: "users"},
	}

	// 병렬 호출을 위한 WaitGroup 및 채널
	var wg sync.WaitGroup
	results := make(chan serviceCount, len(queries))

	for _, q := range queries {
		wg.Add(1)
		go func(q query) {
			defer wg.Done()
			sc := serviceCount{name: q.name}

			data, statusCode, err := h.os.DoRequest("GET", q.serviceType, q.path, nil)
			if err != nil {
				sc.err = err
				results <- sc
				return
			}

			if statusCode >= 400 {
				// API 오류 시 카운트 0으로 처리
				results <- sc
				return
			}

			// 응답 JSON에서 배열 길이를 추출하여 카운트 계산
			var parsed map[string]json.RawMessage
			if err := json.Unmarshal(data, &parsed); err != nil {
				results <- sc
				return
			}

			if raw, ok := parsed[q.countKey]; ok {
				var items []json.RawMessage
				if err := json.Unmarshal(raw, &items); err == nil {
					sc.count = len(items)
				}
			}

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
