// Package database 는 PostgreSQL 연결 및 마이그레이션을 관리한다.
package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// Connect 는 PostgreSQL에 연결하고 연결 풀을 반환한다
func Connect(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("DB 연결 열기 실패: %w", err)
	}

	// 연결 풀 설정 (동시 접속 최대 20명 규모)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)

	// 연결 확인
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("DB 연결 확인 실패: %w", err)
	}

	log.Println("PostgreSQL 연결 성공")
	return db, nil
}

// HealthCheck 는 DB 연결 상태를 확인한다
func HealthCheck(db *sql.DB) error {
	return db.Ping()
}

// RunMigrations 는 마이그레이션 SQL을 실행한다
func RunMigrations(db *sql.DB) error {
	migrations := []string{
		migrationCreateUsers,
		migrationCreateSessions,
		migrationCreateAuditLogs,
	}

	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("마이그레이션 실패: %w", err)
		}
	}

	log.Println("DB 마이그레이션 완료")
	return nil
}

// EnsureAdminUser 는 초기 관리자 계정이 없으면 자동으로 생성한다.
// 환경변수 KCP_ADMIN_USER, KCP_ADMIN_PASSWORD 로 변경 가능.
// 기본값: admin / admin123
func EnsureAdminUser(db *sql.DB) error {
	adminUser := os.Getenv("KCP_ADMIN_USER")
	if adminUser == "" {
		adminUser = "admin"
	}
	adminPass := os.Getenv("KCP_ADMIN_PASSWORD")
	if adminPass == "" {
		adminPass = "admin123"
	}

	// 이미 존재하는지 확인
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", adminUser).Scan(&exists)
	if err != nil {
		return fmt.Errorf("관리자 계정 확인 실패: %w", err)
	}
	if exists {
		log.Printf("관리자 계정 '%s' 이미 존재합니다", adminUser)
		return nil
	}

	// bcrypt 해싱 (비용 인자 12)
	hash, err := bcrypt.GenerateFromPassword([]byte(adminPass), 12)
	if err != nil {
		return fmt.Errorf("비밀번호 해싱 실패: %w", err)
	}

	_, err = db.Exec(
		"INSERT INTO users (username, password_hash, email, role) VALUES ($1, $2, $3, $4)",
		adminUser, string(hash), adminUser+"@kcp.local", "ADMIN",
	)
	if err != nil {
		return fmt.Errorf("관리자 계정 생성 실패: %w", err)
	}

	log.Printf("초기 관리자 계정 생성 완료: %s (비밀번호를 반드시 변경하세요)", adminUser)
	return nil
}

const migrationCreateUsers = `
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE,
    role VARCHAR(20) NOT NULL DEFAULT 'ADMIN',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_user_username ON users(username);
`

const migrationCreateSessions = `
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    auth_type VARCHAR(20) NOT NULL,
    ip_address INET NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_session_user ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_session_expires ON sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_session_token ON sessions(token);
`

const migrationCreateAuditLogs = `
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    action VARCHAR(20) NOT NULL,
    resource_type VARCHAR(30) NOT NULL,
    resource_id VARCHAR(255),
    source VARCHAR(10) NOT NULL,
    status_code INT NOT NULL,
    request_summary TEXT,
    ip_address INET NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_audit_user_date ON audit_logs(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_logs(created_at);
`
