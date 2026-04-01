package repository

import (
	"database/sql"
	"fmt"

	"github.com/kcp-cli/kcp-gateway/internal/model"
)

// SessionRepository 는 세션 데이터 접근을 담당한다
type SessionRepository struct {
	db *sql.DB
}

// NewSessionRepository 는 새로운 SessionRepository를 생성한다
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create 는 새로운 세션을 저장한다
func (r *SessionRepository) Create(s *model.Session) error {
	_, err := r.db.Exec(
		"INSERT INTO sessions (user_id, token, auth_type, ip_address, expires_at) VALUES ($1, $2, $3, $4::inet, $5)",
		s.UserID, s.Token, s.AuthType, s.IPAddress, s.ExpiresAt,
	)
	return err
}

// DeleteByUserID 는 사용자의 모든 세션을 삭제한다
func (r *SessionRepository) DeleteByUserID(userID string) error {
	_, err := r.db.Exec("DELETE FROM sessions WHERE user_id = $1", userID)
	return err
}

// FindByToken 은 토큰으로 세션을 조회한다
func (r *SessionRepository) FindByToken(token string) (*model.Session, error) {
	var s model.Session
	err := r.db.QueryRow(
		"SELECT id, user_id, token, auth_type, ip_address, expires_at, created_at FROM sessions WHERE token = $1 AND expires_at > now()",
		token,
	).Scan(&s.ID, &s.UserID, &s.Token, &s.AuthType, &s.IPAddress, &s.ExpiresAt, &s.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("세션 조회 실패: %w", err)
	}
	return &s, nil
}

// CleanExpired 는 만료된 세션을 삭제한다
func (r *SessionRepository) CleanExpired() (int64, error) {
	result, err := r.db.Exec("DELETE FROM sessions WHERE expires_at < now()")
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
