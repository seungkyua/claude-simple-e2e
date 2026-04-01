// Package model 은 Gateway의 데이터 모델을 정의한다.
package model

import "time"

// User 는 KCP 사용자 모델이다
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Session 은 사용자 세션 모델이다
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	AuthType  string    `json:"auth_type"`
	IPAddress string    `json:"ip_address"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
