// 통계/대시보드 관련 API 핸들러
package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kcp-cli/kcp-gateway/config"
)

// StatsHandler 는 대시보드 통계 관련 API를 처리한다
type StatsHandler struct {
	cfg *config.Config
	db  *sql.DB
}

// NewStatsHandler 는 새로운 StatsHandler를 생성한다
func NewStatsHandler(cfg *config.Config, db *sql.DB) *StatsHandler {
	return &StatsHandler{cfg: cfg, db: db}
}

// Dashboard 는 대시보드 통계 데이터를 조회한다
func (h *StatsHandler) Dashboard(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{"code": "NOT_IMPLEMENTED", "message": "준비 중입니다", "status": 501},
	})
}
