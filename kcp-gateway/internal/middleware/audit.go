package middleware

import (
	"database/sql"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuditLog 는 모든 요청의 감사 로그를 자동 기록하는 미들웨어이다
func AuditLog(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 요청 처리 전 정보 수집
		userID, _ := c.Get("userID")
		source := detectSource(c.GetHeader("User-Agent"))

		// 요청 처리
		c.Next()

		// 요청 처리 후 감사 로그 기록
		action := mapMethodToAction(c.Request.Method)
		resourceType, resourceID := extractResource(c.Request.URL.Path)

		// 비동기 로그 기록 (응답 지연 방지)
		go func() {
			_, err := db.Exec(
				`INSERT INTO audit_logs (user_id, action, resource_type, resource_id, source, status_code, request_summary, ip_address)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8::inet)`,
				userID, action, resourceType, resourceID, source, c.Writer.Status(),
				sanitizeRequestSummary(c.Request.Method, c.Request.URL.Path),
				c.ClientIP(),
			)
			if err != nil {
				log.Printf("감사 로그 기록 실패: %v", err)
			}
		}()
	}
}

// detectSource 는 User-Agent 헤더에서 클라이언트 종류를 판별한다
func detectSource(userAgent string) string {
	ua := strings.ToLower(userAgent)
	switch {
	case strings.Contains(ua, "kcp-cli"):
		return "CLI"
	case strings.Contains(ua, "kcp-tui"):
		return "TUI"
	default:
		return "WEBUI"
	}
}

// mapMethodToAction 은 HTTP 메서드를 감사 로그 액션으로 변환한다
func mapMethodToAction(method string) string {
	switch method {
	case "GET":
		return "READ"
	case "POST":
		return "CREATE"
	case "PUT", "PATCH":
		return "UPDATE"
	case "DELETE":
		return "DELETE"
	default:
		return "READ"
	}
}

// extractResource 는 URL 경로에서 리소스 타입과 ID를 추출한다
func extractResource(path string) (string, string) {
	// /api/v1/compute/servers/abc-123 → ("VM", "abc-123")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 3 {
		return "UNKNOWN", ""
	}

	resourceMap := map[string]string{
		"servers":          "VM",
		"flavors":          "FLAVOR",
		"networks":         "NETWORK",
		"subnets":          "SUBNET",
		"routers":          "ROUTER",
		"security-groups":  "SECURITY_GROUP",
		"volumes":          "VOLUME",
		"snapshots":        "SNAPSHOT",
		"projects":         "PROJECT",
		"users":            "USER",
		"images":           "IMAGE",
		"logs":             "AUDIT_LOG",
	}

	// 리소스 타입 탐색
	for _, p := range parts {
		if rt, ok := resourceMap[p]; ok {
			// 다음 파트가 ID인지 확인
			idx := indexOf(parts, p)
			if idx+1 < len(parts) && !isKnownPath(parts[idx+1]) {
				return rt, parts[idx+1]
			}
			return rt, ""
		}
	}

	return "UNKNOWN", ""
}

// sanitizeRequestSummary 는 요청 요약에서 민감 정보를 마스킹한다
func sanitizeRequestSummary(method, path string) string {
	return method + " " + path
}

func indexOf(slice []string, item string) int {
	for i, s := range slice {
		if s == item {
			return i
		}
	}
	return -1
}

func isKnownPath(s string) bool {
	known := []string{"action", "attach", "detach", "add-interface", "rules", "assign", "revoke"}
	for _, k := range known {
		if s == k {
			return true
		}
	}
	return false
}
