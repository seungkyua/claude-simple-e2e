// Package config 는 Gateway 서버의 설정을 관리한다.
// 모든 민감 정보는 환경변수에서 로드하며, 소스코드에 하드코딩하지 않는다.
package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Config 는 Gateway 서버의 전체 설정을 나타낸다
type Config struct {
	// 서버 설정
	Port           string
	AllowedOrigins []string

	// DB 설정
	DatabaseURL string

	// JWT 설정
	JWTSecret string
	JWTExpiry string

	// OpenStack 설정
	OpenStackAuthURL   string
	OpenStackUsername   string
	OpenStackPassword  string
	OpenStackProjectID string
	OpenStackDomainID  string

	// TLS 설정
	TLSEnabled  bool
	TLSCertPath string
	TLSKeyPath  string
}

// loadEnvFile 은 .env 파일이 있으면 읽어서 환경변수로 설정한다.
// 이미 설정된 환경변수는 덮어쓰지 않는다.
func loadEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return // 파일이 없으면 무시
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 빈 줄, 주석 무시
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		// 이미 설정된 환경변수는 유지 (시스템 환경변수 우선)
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, val)
		}
	}
}

// Load 는 .env 파일과 환경변수에서 설정을 로드한다
func Load() (*Config, error) {
	// .env 파일 자동 로드 (존재하는 경우만)
	loadEnvFile(".env")
	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		AllowedOrigins: strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:3000"), ","),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		JWTExpiry:      getEnv("JWT_EXPIRY", "15m"),

		OpenStackAuthURL:   getEnv("OPENSTACK_AUTH_URL", ""),
		OpenStackUsername:   getEnv("OPENSTACK_USERNAME", ""),
		OpenStackPassword:  getEnv("OPENSTACK_PASSWORD", ""),
		OpenStackProjectID: getEnv("OPENSTACK_PROJECT_ID", ""),
		OpenStackDomainID:  getEnv("OPENSTACK_DOMAIN_ID", "default"),

		TLSEnabled:  getEnv("TLS_ENABLED", "false") == "true",
		TLSCertPath: getEnv("TLS_CERT_PATH", ""),
		TLSKeyPath:  getEnv("TLS_KEY_PATH", ""),
	}

	// 필수 환경변수 검증
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL 환경변수가 설정되지 않았습니다")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET 환경변수가 설정되지 않았습니다")
	}
	if cfg.OpenStackAuthURL == "" {
		return nil, fmt.Errorf("OPENSTACK_AUTH_URL 환경변수가 설정되지 않았습니다")
	}

	return cfg, nil
}

// getEnv 는 환경변수를 읽고, 없으면 기본값을 반환한다
func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
