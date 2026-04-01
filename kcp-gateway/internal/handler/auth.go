package handler

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kcp-cli/kcp-gateway/config"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler 는 인증 관련 API를 처리한다
type AuthHandler struct {
	db  *sql.DB
	cfg *config.Config
}

// NewAuthHandler 는 새로운 AuthHandler를 생성한다
func NewAuthHandler(db *sql.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{db: db, cfg: cfg}
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	AuthType string `json:"authType"`
}

// Login 은 사용자 인증 후 JWT 토큰을 발급한다
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_INPUT", "message": "필수 필드가 누락되었습니다", "status": 400},
		})
		return
	}

	if req.AuthType == "" {
		req.AuthType = "JWT"
	}

	// 사용자 조회
	var userID, username, passwordHash, role string
	err := h.db.QueryRow(
		"SELECT id, username, password_hash, role FROM users WHERE username = $1 AND is_active = true",
		req.Username,
	).Scan(&userID, &username, &passwordHash, &role)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"code": "AUTH_FAILED", "message": "자격 증명이 올바르지 않습니다", "status": 401},
		})
		return
	}

	// 비밀번호 검증 (bcrypt)
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"code": "AUTH_FAILED", "message": "자격 증명이 올바르지 않습니다", "status": 401},
		})
		return
	}

	// JWT 토큰 생성 (15분 만료)
	expiresAt := time.Now().Add(15 * time.Minute)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      userID,
		"username": username,
		"role":     role,
		"exp":      expiresAt.Unix(),
		"iat":      time.Now().Unix(),
		"jti":      uuid.New().String(),
	})

	tokenStr, err := token.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "TOKEN_ERROR", "message": "토큰 생성에 실패했습니다", "status": 500},
		})
		return
	}

	// 세션 저장
	_, _ = h.db.Exec(
		"INSERT INTO sessions (user_id, token, auth_type, ip_address, expires_at) VALUES ($1, $2, $3, $4::inet, $5)",
		userID, tokenStr, req.AuthType, c.ClientIP(), expiresAt,
	)

	c.JSON(http.StatusOK, gin.H{
		"token":     tokenStr,
		"authType":  req.AuthType,
		"expiresAt": expiresAt.Format(time.RFC3339),
		"user": gin.H{
			"id":       userID,
			"username": username,
			"role":     role,
		},
	})
}

// Logout 은 현재 세션을 무효화한다
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, _ := c.Get("userID")
	_, _ = h.db.Exec("DELETE FROM sessions WHERE user_id = $1", userID)
	c.Status(http.StatusNoContent)
}

// Refresh 는 만료된 토큰을 갱신한다
func (h *AuthHandler) Refresh(c *gin.Context) {
	// TODO: 토큰 갱신 로직 구현
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{"code": "NOT_IMPLEMENTED", "message": "토큰 갱신 기능은 준비 중입니다", "status": 501},
	})
}
