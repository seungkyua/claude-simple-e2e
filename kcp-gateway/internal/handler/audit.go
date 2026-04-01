// 감사 로그 관련 API 핸들러 — 로컬 DB(PostgreSQL) 연동
package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kcp-cli/kcp-gateway/internal/model"
	"github.com/kcp-cli/kcp-gateway/internal/repository"
)

// AuditHandler 는 감사 로그 관련 API를 처리한다
type AuditHandler struct {
	db   *sql.DB
	repo *repository.AuditRepository
}

// NewAuditHandler 는 새로운 AuditHandler를 생성한다
func NewAuditHandler(db *sql.DB) *AuditHandler {
	return &AuditHandler{
		db:   db,
		repo: repository.NewAuditRepository(db),
	}
}

// ListLogs 는 감사 로그 목록을 조회한다.
// 쿼리 파라미터: user_id, action, resource_type, from, to, page, size
func (h *AuditHandler) ListLogs(c *gin.Context) {
	// 쿼리 파라미터에서 필터 조건 추출
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	filter := &model.AuditLogFilter{
		UserID:       c.Query("user_id"),
		Action:       c.Query("action"),
		ResourceType: c.Query("resource_type"),
		From:         c.Query("from"),
		To:           c.Query("to"),
		Page:         page,
		Size:         size,
	}

	// 리포지토리를 통해 감사 로그 조회
	logs, total, err := h.repo.List(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DB_ERROR",
				"message": "감사 로그 조회 중 오류가 발생했습니다",
				"status":  500,
			},
		})
		return
	}

	// 페이지네이션 메타데이터 포함 응답
	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"total": total,
		"page":  page,
		"size":  size,
	})
}
