package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorHandler 는 패닉을 복구하고 공통 에러 응답을 반환하는 미들웨어이다
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("패닉 복구: %v", r)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"code":    "INTERNAL_ERROR",
						"message": "서버 내부 오류가 발생했습니다",
						"status":  500,
					},
				})
			}
		}()
		c.Next()
	}
}

// CORS 는 Cross-Origin Resource Sharing 미들웨어이다
func CORS(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		for _, o := range allowedOrigins {
			if o == origin || o == "*" {
				c.Header("Access-Control-Allow-Origin", origin)
				break
			}
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, User-Agent")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
