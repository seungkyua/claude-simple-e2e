// 감사 로그 관련 API 핸들러
package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuditHandler 는 감사 로그 관련 API를 처리한다
type AuditHandler struct {
	db *sql.DB
}

// NewAuditHandler 는 새로운 AuditHandler를 생성한다
func NewAuditHandler(db *sql.DB) *AuditHandler {
	return &AuditHandler{db: db}
}

// ListLogs 는 감사 로그 목록을 조회한다
func (h *AuditHandler) ListLogs(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{"code": "NOT_IMPLEMENTED", "message": "준비 중입니다", "status": 501},
	})
}
