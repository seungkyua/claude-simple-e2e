// Package service 는 Gateway의 비즈니스 로직을 제공한다.
package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/kcp-cli/kcp-gateway/internal/model"
	"github.com/kcp-cli/kcp-gateway/internal/repository"
)

// AuthService 는 인증 비즈니스 로직을 처리한다
type AuthService struct {
	userRepo    *repository.UserRepository
	sessionRepo *repository.SessionRepository
	jwtSecret   string
	jwtExpiry   time.Duration
}

// NewAuthService 는 새로운 AuthService를 생성한다
func NewAuthService(userRepo *repository.UserRepository, sessionRepo *repository.SessionRepository, jwtSecret string, expiry time.Duration) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtSecret:   jwtSecret,
		jwtExpiry:   expiry,
	}
}

// LoginResult 는 로그인 결과이다
type LoginResult struct {
	Token     string
	ExpiresAt time.Time
	User      *model.User
}

// Login 은 사용자 인증을 수행하고 JWT 토큰을 발급한다
func (s *AuthService) Login(username, password, authType, ip string) (*LoginResult, error) {
	// 사용자 조회
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("인증 실패: 사용자를 찾을 수 없습니다")
	}

	if !user.IsActive {
		return nil, fmt.Errorf("인증 실패: 비활성화된 계정입니다")
	}

	// 비밀번호 검증 (bcrypt)
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("인증 실패: 자격 증명이 올바르지 않습니다")
	}

	// JWT 토큰 생성
	expiresAt := time.Now().Add(s.jwtExpiry)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      expiresAt.Unix(),
		"iat":      time.Now().Unix(),
		"jti":      uuid.New().String(),
	})

	tokenStr, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("토큰 생성 실패: %w", err)
	}

	// 세션 저장
	session := &model.Session{
		UserID:    user.ID,
		Token:     tokenStr,
		AuthType:  authType,
		IPAddress: ip,
		ExpiresAt: expiresAt,
	}
	_ = s.sessionRepo.Create(session)

	return &LoginResult{
		Token:     tokenStr,
		ExpiresAt: expiresAt,
		User:      user,
	}, nil
}

// Logout 은 사용자의 모든 세션을 무효화한다
func (s *AuthService) Logout(userID string) error {
	return s.sessionRepo.DeleteByUserID(userID)
}

// HashPassword 는 비밀번호를 bcrypt로 해싱한다 (비용 인자 12)
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
