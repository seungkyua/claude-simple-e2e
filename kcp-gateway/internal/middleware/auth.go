// Package middleware 는 Gateway의 Gin 미들웨어를 제공한다.
package middleware

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Auth 는 JWT/세션/OAuth2 인증을 검증하는 미들웨어이다
func Auth(jwtSecret string, db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "인증 토큰이 필요합니다",
					"status":  401,
				},
			})
			return
		}

		// Bearer 토큰 추출
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "INVALID_TOKEN_FORMAT",
					"message": "Bearer 토큰 형식이 올바르지 않습니다",
					"status":  401,
				},
			})
			return
		}

		// JWT 토큰 검증
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "INVALID_TOKEN",
					"message": "토큰이 유효하지 않거나 만료되었습니다",
					"status":  401,
				},
			})
			return
		}

		// 클레임에서 사용자 정보 추출
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "INVALID_CLAIMS",
					"message": "토큰 클레임을 읽을 수 없습니다",
					"status":  401,
				},
			})
			return
		}

		// 컨텍스트에 사용자 정보 저장
		c.Set("userID", claims["sub"])
		c.Set("username", claims["username"])
		c.Next()
	}
}
