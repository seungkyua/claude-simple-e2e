package main

import (
	"log"
	"os"

	"github.com/kcp-cli/kcp-gateway/config"
	"github.com/kcp-cli/kcp-gateway/internal/database"
	"github.com/kcp-cli/kcp-gateway/internal/handler"
	"github.com/kcp-cli/kcp-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	// 서버 설정 로드
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("설정 로드 실패: %v", err)
	}

	// DB 연결
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("DB 연결 실패: %v", err)
	}
	defer db.Close()

	// DB 마이그레이션 실행
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("마이그레이션 실패: %v", err)
	}

	// 초기 관리자 계정 생성 (없는 경우에만)
	if err := database.EnsureAdminUser(db); err != nil {
		log.Printf("초기 관리자 계정 확인 실패: %v", err)
	}

	// Gin 라우터 설정
	r := gin.Default()

	// 공통 미들웨어
	r.Use(middleware.CORS(cfg.AllowedOrigins))
	r.Use(middleware.ErrorHandler())

	// API v1 라우트 그룹
	v1 := r.Group("/api/v1")

	// 인증 API (미들웨어 미적용)
	handler.RegisterAuthRoutes(v1, db, cfg)

	// 인증 필요 라우트
	auth := v1.Group("")
	auth.Use(middleware.Auth(cfg.JWTSecret, db))
	auth.Use(middleware.AuditLog(db))

	handler.RegisterComputeRoutes(auth, cfg)
	handler.RegisterNetworkRoutes(auth, cfg)
	handler.RegisterStorageRoutes(auth, cfg)
	handler.RegisterIdentityRoutes(auth, cfg)
	handler.RegisterImageRoutes(auth, cfg)
	handler.RegisterAuditRoutes(auth, db)
	handler.RegisterStatsRoutes(auth, cfg, db)

	// 서버 시작
	addr := ":" + cfg.Port
	log.Printf("KCP Gateway 시작: %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("서버 시작 실패: %v", err)
		os.Exit(1)
	}
}
