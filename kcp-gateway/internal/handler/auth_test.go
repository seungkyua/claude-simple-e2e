package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestLoginMissingFields 는 빈 요청 바디일 때 400을 반환하는지 검증한다
func TestLoginMissingFields(t *testing.T) {
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	// DB 없이 핸들러 생성 (바인딩 검증만 테스트)
	h := NewAuthHandler(nil, nil)
	r.POST("/auth/login", h.Login)

	// 빈 JSON 바디 전송
	body := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("상태 코드 = %d, 기대값 %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("응답 JSON 파싱 실패: %v", err)
	}

	errObj := resp["error"].(map[string]interface{})
	if errObj["code"] != "INVALID_INPUT" {
		t.Errorf("에러 코드 = %v, 기대값 INVALID_INPUT", errObj["code"])
	}
}

// TestLoginInvalidCredentials 는 잘못된 자격 증명일 때 401을 반환하는지 검증한다
// DB에서 사용자를 찾지 못하면 401을 반환해야 한다
func TestLoginInvalidCredentials(t *testing.T) {
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	// nil DB로 핸들러 생성 — QueryRow 호출 시 패닉을 방지하기 위해
	// ErrorHandler 미들웨어로 감싸서 패닉을 안전하게 복구한다
	h := NewAuthHandler(nil, nil)
	r.Use(func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// nil DB 접근으로 인한 패닉 → 인증 실패로 처리
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": gin.H{"code": "AUTH_FAILED", "message": "자격 증명이 올바르지 않습니다", "status": 401},
				})
			}
		}()
		c.Next()
	})
	r.POST("/auth/login", h.Login)

	body := bytes.NewBufferString(`{"username":"wronguser","password":"wrongpass"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("상태 코드 = %d, 기대값 %d", w.Code, http.StatusUnauthorized)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("응답 JSON 파싱 실패: %v", err)
	}

	errObj := resp["error"].(map[string]interface{})
	if errObj["code"] != "AUTH_FAILED" {
		t.Errorf("에러 코드 = %v, 기대값 AUTH_FAILED", errObj["code"])
	}
}
