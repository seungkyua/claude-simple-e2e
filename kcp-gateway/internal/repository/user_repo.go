// Package repository 는 데이터베이스 접근 레이어를 제공한다.
package repository

import (
	"database/sql"
	"fmt"

	"github.com/kcp-cli/kcp-gateway/internal/model"
)

// UserRepository 는 사용자 데이터 접근을 담당한다
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository 는 새로운 UserRepository를 생성한다
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByUsername 은 사용자명으로 사용자를 조회한다
func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var u model.User
	err := r.db.QueryRow(
		"SELECT id, username, password_hash, email, role, is_active, created_at, updated_at FROM users WHERE username = $1",
		username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Email, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("사용자 조회 실패: %w", err)
	}
	return &u, nil
}

// Create 는 새로운 사용자를 생성한다
func (r *UserRepository) Create(u *model.User) error {
	_, err := r.db.Exec(
		"INSERT INTO users (username, password_hash, email, role) VALUES ($1, $2, $3, $4)",
		u.Username, u.PasswordHash, u.Email, u.Role,
	)
	return err
}
