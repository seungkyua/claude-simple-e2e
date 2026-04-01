package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestErrorHandlerRecoversPanic 는 핸들러에서 패닉이 발생했을 때 500 응답을 반환하는지 검증한다
func TestErrorHandlerRecoversPanic(t *testing.T) {
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(ErrorHandler())
	r.GET("/panic", func(c *gin.Context) {
		panic("테스트 패닉")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("상태 코드 = %d, 기대값 %d", w.Code, http.StatusInternalServerError)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("응답 JSON 파싱 실패: %v", err)
	}
	errObj := body["error"].(map[string]interface{})
	if errObj["code"] != "INTERNAL_ERROR" {
		t.Errorf("에러 코드 = %v, 기대값 INTERNAL_ERROR", errObj["code"])
	}
}

// TestCORSAllowedOrigin 는 허용된 Origin에 대해 올바른 CORS 헤더가 설정되는지 검증한다
func TestCORSAllowedOrigin(t *testing.T) {
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(CORS([]string{"http://localhost:3000", "https://example.com"}))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	r.ServeHTTP(w, req)

	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin != "http://localhost:3000" {
		t.Errorf("Allow-Origin = %q, 기대값 %q", origin, "http://localhost:3000")
	}

	methods := w.Header().Get("Access-Control-Allow-Methods")
	if methods == "" {
		t.Error("Access-Control-Allow-Methods 헤더가 비어있습니다")
	}
}

// TestCORSDisallowedOrigin 는 허용되지 않은 Origin에 대해 CORS 헤더가 없는지 검증한다
func TestCORSDisallowedOrigin(t *testing.T) {
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(CORS([]string{"http://localhost:3000"}))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://evil.com")
	r.ServeHTTP(w, req)

	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin != "" {
		t.Errorf("허용되지 않은 Origin에 Allow-Origin 헤더가 설정됨: %q", origin)
	}
}

// TestCORSPreflight 는 OPTIONS 요청에 대해 204를 반환하는지 검증한다
func TestCORSPreflight(t *testing.T) {
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(CORS([]string{"http://localhost:3000"}))
	// OPTIONS 라우트가 등록되지 않아도 미들웨어에서 처리
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("상태 코드 = %d, 기대값 %d", w.Code, http.StatusNoContent)
	}

	maxAge := w.Header().Get("Access-Control-Max-Age")
	if maxAge != "86400" {
		t.Errorf("Max-Age = %q, 기대값 %q", maxAge, "86400")
	}
}
