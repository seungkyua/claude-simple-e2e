// OpenStack API 응답 처리를 위한 공통 헬퍼 함수
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// kcpListResponse 는 KCP 통일 목록 응답 형식이다
// CLI와 WebUI 모두 이 형식을 기대한다
type kcpListResponse struct {
	Items      interface{}   `json:"items"`
	Pagination kcpPagination `json:"pagination"`
}

// kcpPagination 은 페이지네이션 정보이다
type kcpPagination struct {
	Page  int `json:"page"`
	Size  int `json:"size"`
	Total int `json:"total"`
}

// forwardOSListResponse 는 OpenStack 목록 응답을 KCP 통일 포맷으로 변환하여 전달한다.
// osKey: OpenStack 응답의 루트 키 (예: "servers", "projects", "networks")
// 예: {"servers": [...]} → {"items": [...], "pagination": {"page": 1, "size": N, "total": N}}
func forwardOSListResponse(c *gin.Context, data []byte, statusCode int, err error, osKey string) {
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"code":    "OPENSTACK_ERROR",
				"message": err.Error(),
				"status":  502,
			},
		})
		return
	}

	// 에러 상태 코드 처리
	if statusCode >= 400 {
		forwardOSResponse(c, data, statusCode, nil)
		return
	}

	// OpenStack 응답 파싱
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		// 파싱 실패 시 원본 전달
		forwardOSResponse(c, data, statusCode, nil)
		return
	}

	// 지정된 키에서 배열 추출
	itemsRaw, ok := raw[osKey]
	if !ok {
		// 키가 없으면 원본 전달
		forwardOSResponse(c, data, statusCode, nil)
		return
	}

	// 배열 파싱
	var items []json.RawMessage
	if err := json.Unmarshal(itemsRaw, &items); err != nil {
		forwardOSResponse(c, data, statusCode, nil)
		return
	}

	// KCP 통일 포맷으로 변환
	resp := kcpListResponse{
		Items: items,
		Pagination: kcpPagination{
			Page:  1,
			Size:  len(items),
			Total: len(items),
		},
	}

	c.JSON(statusCode, resp)
}

// forwardOSSingleResponse 는 OpenStack 단일 리소스 응답에서 지정된 키의 값을 추출하여 전달한다.
// 예: {"server": {...}} → {...}
func forwardOSSingleResponse(c *gin.Context, data []byte, statusCode int, err error, osKey string) {
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"code":    "OPENSTACK_ERROR",
				"message": err.Error(),
				"status":  502,
			},
		})
		return
	}

	if statusCode >= 400 {
		forwardOSResponse(c, data, statusCode, nil)
		return
	}

	// 단일 리소스 응답 파싱
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		forwardOSResponse(c, data, statusCode, nil)
		return
	}

	if itemRaw, ok := raw[osKey]; ok {
		c.Data(statusCode, "application/json; charset=utf-8", itemRaw)
		return
	}

	// 키가 없으면 원본 전달
	forwardOSResponse(c, data, statusCode, nil)
}

// forwardOSResponse 는 OpenStack API 응답을 JSON 형식으로 클라이언트에 전달한다.
// OpenStack 에러 응답(HTML 포함)을 통일된 JSON 에러 포맷으로 변환한다.
func forwardOSResponse(c *gin.Context, data []byte, statusCode int, err error) {
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"code":    "OPENSTACK_ERROR",
				"message": err.Error(),
				"status":  502,
			},
		})
		return
	}

	// 에러 상태 코드인 경우 JSON 변환 보장
	if statusCode >= 400 {
		var jsonCheck map[string]interface{}
		if json.Unmarshal(data, &jsonCheck) == nil {
			c.Data(statusCode, "application/json; charset=utf-8", data)
			return
		}
		c.JSON(statusCode, gin.H{
			"error": gin.H{
				"code":    httpStatusToCode(statusCode),
				"message": httpStatusToMessage(statusCode),
				"status":  statusCode,
			},
		})
		return
	}

	// 성공 응답
	if len(data) == 0 {
		c.Status(statusCode)
		return
	}

	var jsonCheck json.RawMessage
	if json.Unmarshal(data, &jsonCheck) == nil {
		c.Data(statusCode, "application/json; charset=utf-8", data)
		return
	}

	c.JSON(statusCode, gin.H{
		"data": string(data),
	})
}

func httpStatusToCode(status int) string {
	switch status {
	case 400:
		return "BAD_REQUEST"
	case 401:
		return "UNAUTHORIZED"
	case 403:
		return "FORBIDDEN"
	case 404:
		return "NOT_FOUND"
	case 409:
		return "CONFLICT"
	case 500:
		return "OPENSTACK_INTERNAL_ERROR"
	case 503:
		return "SERVICE_UNAVAILABLE"
	default:
		return "OPENSTACK_ERROR"
	}
}

func httpStatusToMessage(status int) string {
	switch status {
	case 400:
		return "잘못된 요청입니다"
	case 401:
		return "OpenStack 인증이 필요합니다"
	case 403:
		return "접근이 거부되었습니다"
	case 404:
		return "리소스를 찾을 수 없습니다"
	case 409:
		return "리소스 충돌이 발생했습니다"
	case 500:
		return "OpenStack 서버 내부 오류입니다"
	case 503:
		return "OpenStack 서비스를 사용할 수 없습니다"
	default:
		return "OpenStack API 오류가 발생했습니다"
	}
}
