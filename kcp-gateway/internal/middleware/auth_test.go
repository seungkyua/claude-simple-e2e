package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// 테스트용 JWT 토큰을 생성하는 헬퍼 함수
func createTestToken(secret string, claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte(secret))
	return tokenStr
}

// TestAuthMissingHeader 는 Authorization 헤더가 없을 때 401을 반환하는지 검증한다
func TestAuthMissingHeader(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	handler := Auth("test-secret", nil)
	handler(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("상태 코드 = %d, 기대값 %d", w.Code, http.StatusUnauthorized)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("응답 JSON 파싱 실패: %v", err)
	}
	errObj := body["error"].(map[string]interface{})
	if errObj["code"] != "UNAUTHORIZED" {
		t.Errorf("에러 코드 = %v, 기대값 UNAUTHORIZED", errObj["code"])
	}
}

// TestAuthInvalidFormat 는 Bearer 접두사가 없는 토큰일 때 401을 반환하는지 검증한다
func TestAuthInvalidFormat(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "InvalidTokenFormat")

	handler := Auth("test-secret", nil)
	handler(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("상태 코드 = %d, 기대값 %d", w.Code, http.StatusUnauthorized)
	}

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	errObj := body["error"].(map[string]interface{})
	if errObj["code"] != "INVALID_TOKEN_FORMAT" {
		t.Errorf("에러 코드 = %v, 기대값 INVALID_TOKEN_FORMAT", errObj["code"])
	}
}

// TestAuthInvalidToken 는 유효하지 않은 JWT 토큰일 때 401을 반환하는지 검증한다
func TestAuthInvalidToken(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer invalid.jwt.token")

	handler := Auth("test-secret", nil)
	handler(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("상태 코드 = %d, 기대값 %d", w.Code, http.StatusUnauthorized)
	}

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	errObj := body["error"].(map[string]interface{})
	if errObj["code"] != "INVALID_TOKEN" {
		t.Errorf("에러 코드 = %v, 기대값 INVALID_TOKEN", errObj["code"])
	}
}

// TestAuthValidToken 는 유효한 JWT 토큰일 때 200을 반환하고 컨텍스트에 사용자 정보가 저장되는지 검증한다
func TestAuthValidToken(t *testing.T) {
	secret := "test-secret-key"
	tokenStr := createTestToken(secret, jwt.MapClaims{
		"sub":      "user-123",
		"username": "testuser",
		"exp":      time.Now().Add(15 * time.Minute).Unix(),
	})

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	var capturedUserID, capturedUsername interface{}

	// 미들웨어 이후 실행될 핸들러 등록
	r.GET("/test", Auth(secret, nil), func(c *gin.Context) {
		capturedUserID, _ = c.Get("userID")
		capturedUsername, _ = c.Get("username")
		c.Status(http.StatusOK)
	})

	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+tokenStr)
	r.ServeHTTP(w, c.Request)

	if w.Code != http.StatusOK {
		t.Errorf("상태 코드 = %d, 기대값 %d", w.Code, http.StatusOK)
	}
	if capturedUserID != "user-123" {
		t.Errorf("userID = %v, 기대값 user-123", capturedUserID)
	}
	if capturedUsername != "testuser" {
		t.Errorf("username = %v, 기대값 testuser", capturedUsername)
	}
}
