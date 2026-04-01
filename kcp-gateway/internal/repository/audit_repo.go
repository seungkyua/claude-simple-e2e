package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/kcp-cli/kcp-gateway/internal/model"
)

// AuditRepository 는 감사 로그 데이터 접근을 담당한다
type AuditRepository struct {
	db *sql.DB
}

// NewAuditRepository 는 새로운 AuditRepository를 생성한다
func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// List 는 필터 조건에 따라 감사 로그를 조회한다
func (r *AuditRepository) List(f *model.AuditLogFilter) ([]model.AuditLog, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if f.UserID != "" {
		conditions = append(conditions, fmt.Sprintf("a.user_id = $%d", argIdx))
		args = append(args, f.UserID)
		argIdx++
	}
	if f.Action != "" {
		conditions = append(conditions, fmt.Sprintf("a.action = $%d", argIdx))
		args = append(args, f.Action)
		argIdx++
	}
	if f.ResourceType != "" {
		conditions = append(conditions, fmt.Sprintf("a.resource_type = $%d", argIdx))
		args = append(args, f.ResourceType)
		argIdx++
	}
	if f.From != "" {
		conditions = append(conditions, fmt.Sprintf("a.created_at >= $%d", argIdx))
		args = append(args, f.From)
		argIdx++
	}
	if f.To != "" {
		conditions = append(conditions, fmt.Sprintf("a.created_at <= $%d", argIdx))
		args = append(args, f.To)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 총 개수 조회
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_logs a %s", where)
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("감사 로그 개수 조회 실패: %w", err)
	}

	// 페이지네이션
	page := f.Page
	size := f.Size
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	offset := (page - 1) * size

	query := fmt.Sprintf(`
		SELECT a.id, a.user_id, u.username, a.action, a.resource_type, a.resource_id,
		       a.source, a.status_code, a.request_summary, a.ip_address, a.created_at
		FROM audit_logs a
		LEFT JOIN users u ON a.user_id = u.id
		%s
		ORDER BY a.created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)
	args = append(args, size, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("감사 로그 조회 실패: %w", err)
	}
	defer rows.Close()

	var logs []model.AuditLog
	for rows.Next() {
		var l model.AuditLog
		var resourceID, requestSummary sql.NullString
		err := rows.Scan(&l.ID, &l.UserID, &l.Username, &l.Action, &l.ResourceType, &resourceID,
			&l.Source, &l.StatusCode, &requestSummary, &l.IPAddress, &l.CreatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("감사 로그 행 스캔 실패: %w", err)
		}
		if resourceID.Valid {
			l.ResourceID = resourceID.String
		}
		if requestSummary.Valid {
			l.RequestSummary = requestSummary.String
		}
		logs = append(logs, l)
	}

	return logs, total, nil
}
