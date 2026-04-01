package main

import (
	"flag"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	ossdk "github.com/kcp-cli/kcp-cli/pkg/sdk/openstack"
	"github.com/kcp-cli/kcp-gateway/config"
	"github.com/kcp-cli/kcp-gateway/internal/database"
	"github.com/kcp-cli/kcp-gateway/internal/handler"
	"github.com/kcp-cli/kcp-gateway/internal/middleware"
)

func main() {
	// --config 플래그로 설정 파일 경로 지정 (기본: 현재 디렉토리의 kcp-gateway-config.yaml)
	configPath := flag.String("config", "", "설정 파일 경로 (기본: ./kcp-gateway-config.yaml)")
	flag.Parse()

	// 서버 설정 로드
	cfg, err := config.Load(*configPath)
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

	// OpenStack SDK 클라이언트 초기화
	osClient, err := ossdk.NewClient(&ossdk.OSConfig{
		AuthURL:         cfg.OpenStackAuthURL,
		AuthType:        cfg.OpenStackAuthType,
		Username:        cfg.OpenStackUsername,
		Password:        cfg.OpenStackPassword,
		ProjectID:       cfg.OpenStackProjectID,
		ProjectName:     cfg.OpenStackProjectName,
		ProjectDomainID: cfg.OpenStackProjectDomainID,
		UserDomainID:    cfg.OpenStackUserDomainID,
		DomainName:      cfg.OpenStackDomainName,
		RegionName:      cfg.OpenStackRegionName,
		Insecure:        cfg.OpenStackInsecure,
	})
	if err != nil {
		log.Printf("WARNING: %v", err)
		log.Println("Gateway는 시작되지만, OpenStack API 호출 시 재인증을 시도합니다")
	} else {
		log.Println("OpenStack 인증 성공")
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

	handler.RegisterComputeRoutes(auth, osClient)
	handler.RegisterNetworkRoutes(auth, osClient)
	handler.RegisterStorageRoutes(auth, osClient)
	handler.RegisterIdentityRoutes(auth, osClient)
	handler.RegisterImageRoutes(auth, osClient)
	handler.RegisterAuditRoutes(auth, db)
	handler.RegisterStatsRoutes(auth, osClient, db)

	// 서버 시작
	addr := ":" + cfg.Port
	log.Printf("KCP Gateway 시작: %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("서버 시작 실패: %v", err)
		os.Exit(1)
	}
}
