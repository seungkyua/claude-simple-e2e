package service

import (
	"github.com/kcp-cli/kcp-gateway/internal/model"
	"github.com/kcp-cli/kcp-gateway/internal/repository"
)

// AuditService 는 감사 로그 비즈니스 로직을 처리한다
type AuditService struct {
	repo *repository.AuditRepository
}

// NewAuditService 는 새로운 AuditService를 생성한다
func NewAuditService(repo *repository.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

// ListLogs 는 감사 로그를 조회한다
func (s *AuditService) ListLogs(filter *model.AuditLogFilter) ([]model.AuditLog, int, error) {
	return s.repo.List(filter)
}
