// Package config 는 Gateway 서버의 설정을 관리한다.
// 설정 로드 우선순위:
//  1. --config 플래그로 지정한 YAML 파일
//  2. 현재 디렉토리의 kcp-gateway-config.yaml
//  3. 환경변수 (YAML 파일이 없는 경우 fallback)
//
// YAML 파일 내의 값보다 환경변수가 우선한다 (이미 설정된 환경변수는 덮어쓰지 않음).
package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// DefaultConfigFile 은 기본 설정 파일명이다
const DefaultConfigFile = "kcp-gateway-config.yaml"

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

	// OpenStack 설정 (openrc 호환)
	OpenStackAuthType        string
	OpenStackAuthURL         string
	OpenStackUsername         string
	OpenStackPassword        string
	OpenStackProjectID       string
	OpenStackProjectName     string
	OpenStackProjectDomainID string
	OpenStackUserDomainID    string
	OpenStackDomainName      string
	OpenStackRegionName      string
	OpenStackInsecure        bool

	// TLS 설정
	TLSEnabled  bool
	TLSCertPath string
	TLSKeyPath  string
}

// yamlConfig 는 YAML 파일의 구조를 나타낸다
type yamlConfig struct {
	Server    yamlServer    `yaml:"server"`
	Database  yamlDatabase  `yaml:"database"`
	JWT       yamlJWT       `yaml:"jwt"`
	OpenStack yamlOpenStack `yaml:"openstack"`
	TLS       yamlTLS       `yaml:"tls"`
}

type yamlServer struct {
	Port           string   `yaml:"port"`
	AllowedOrigins []string `yaml:"allowed_origins"`
}

type yamlDatabase struct {
	URL string `yaml:"url"`
}

type yamlJWT struct {
	Secret string `yaml:"secret"`
	Expiry string `yaml:"expiry"`
}

type yamlOpenStack struct {
	AuthType        string `yaml:"auth_type"`
	AuthURL         string `yaml:"auth_url"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	ProjectID       string `yaml:"project_id"`
	ProjectName     string `yaml:"project_name"`
	ProjectDomainID string `yaml:"project_domain_id"`
	UserDomainID    string `yaml:"user_domain_id"`
	DomainName      string `yaml:"domain_name"`
	RegionName      string `yaml:"region_name"`
	Insecure        bool   `yaml:"insecure"`
}

type yamlTLS struct {
	Enabled  bool   `yaml:"enabled"`
	CertPath string `yaml:"cert_path"`
	KeyPath  string `yaml:"key_path"`
}

// Load 는 YAML 설정 파일과 환경변수에서 설정을 로드한다.
// configPath가 빈 문자열이면 현재 디렉토리의 kcp-gateway-config.yaml을 시도한다.
// YAML 파일이 없으면 환경변수만으로 설정을 구성한다.
func Load(configPath string) (*Config, error) {
	var ycfg yamlConfig
	yamlLoaded := false

	// 설정 파일 경로 결정
	if configPath == "" {
		configPath = DefaultConfigFile
	}

	// YAML 파일 로드 시도
	data, err := os.ReadFile(configPath)
	if err == nil {
		if err := yaml.Unmarshal(data, &ycfg); err != nil {
			return nil, fmt.Errorf("설정 파일 파싱 실패 (%s): %w", configPath, err)
		}
		yamlLoaded = true
		log.Printf("설정 파일 로드: %s", configPath)
	} else if configPath != DefaultConfigFile {
		// 명시적으로 지정한 파일이 없으면 에러
		return nil, fmt.Errorf("설정 파일을 찾을 수 없습니다: %s", configPath)
	} else {
		log.Println("설정 파일 없음 — 환경변수에서 설정을 로드합니다")
	}

	// YAML 값을 기본값으로 사용하고, 환경변수가 있으면 덮어쓴다
	cfg := &Config{
		Port:           resolveStr(ycfg.Server.Port, "PORT", "8080"),
		AllowedOrigins: resolveStrSlice(ycfg.Server.AllowedOrigins, "ALLOWED_ORIGINS", "http://localhost:3000"),
		DatabaseURL:    resolveStr(ycfg.Database.URL, "DATABASE_URL", ""),
		JWTSecret:      resolveStr(ycfg.JWT.Secret, "JWT_SECRET", ""),
		JWTExpiry:      resolveStr(ycfg.JWT.Expiry, "JWT_EXPIRY", "15m"),

		OpenStackAuthType:        resolveStr(ycfg.OpenStack.AuthType, "OS_AUTH_TYPE", "password"),
		OpenStackAuthURL:         resolveStr(ycfg.OpenStack.AuthURL, "OS_AUTH_URL", ""),
		OpenStackUsername:        resolveStr(ycfg.OpenStack.Username, "OS_USERNAME", ""),
		OpenStackPassword:       resolveStr(ycfg.OpenStack.Password, "OS_PASSWORD", ""),
		OpenStackProjectID:      resolveStr(ycfg.OpenStack.ProjectID, "OS_PROJECT_ID", ""),
		OpenStackProjectName:    resolveStr(ycfg.OpenStack.ProjectName, "OS_PROJECT_NAME", ""),
		OpenStackProjectDomainID: resolveStr(ycfg.OpenStack.ProjectDomainID, "OS_PROJECT_DOMAIN_ID", "default"),
		OpenStackUserDomainID:   resolveStr(ycfg.OpenStack.UserDomainID, "OS_USER_DOMAIN_ID", "default"),
		OpenStackDomainName:     resolveStr(ycfg.OpenStack.DomainName, "OS_USER_DOMAIN_NAME", "Default"),
		OpenStackRegionName:     resolveStr(ycfg.OpenStack.RegionName, "OS_REGION_NAME", ""),
		OpenStackInsecure:       resolveBool(ycfg.OpenStack.Insecure, "OS_INSECURE", yamlLoaded),

		TLSEnabled:  resolveBool(ycfg.TLS.Enabled, "TLS_ENABLED", yamlLoaded),
		TLSCertPath: resolveStr(ycfg.TLS.CertPath, "TLS_CERT_PATH", ""),
		TLSKeyPath:  resolveStr(ycfg.TLS.KeyPath, "TLS_KEY_PATH", ""),
	}

	// 필수 값 검증
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("database.url (또는 DATABASE_URL 환경변수)이 설정되지 않았습니다")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("jwt.secret (또는 JWT_SECRET 환경변수)이 설정되지 않았습니다")
	}
	if cfg.OpenStackAuthURL == "" {
		return nil, fmt.Errorf("openstack.auth_url (또는 OS_AUTH_URL 환경변수)이 설정되지 않았습니다")
	}
	if cfg.OpenStackUsername == "" {
		return nil, fmt.Errorf("openstack.username (또는 OS_USERNAME 환경변수)이 설정되지 않았습니다")
	}
	if cfg.OpenStackPassword == "" {
		return nil, fmt.Errorf("openstack.password (또는 OS_PASSWORD 환경변수)이 설정되지 않았습니다")
	}

	return cfg, nil
}

// resolveStr 은 환경변수 > YAML 값 > 기본값 순으로 문자열을 결정한다
func resolveStr(yamlVal, envKey, fallback string) string {
	if envVal, ok := os.LookupEnv(envKey); ok {
		return envVal
	}
	if yamlVal != "" {
		return yamlVal
	}
	return fallback
}

// resolveStrSlice 는 환경변수(콤마 구분) > YAML 배열 > 기본값 순으로 결정한다
func resolveStrSlice(yamlVal []string, envKey, fallback string) []string {
	if envVal, ok := os.LookupEnv(envKey); ok {
		return strings.Split(envVal, ",")
	}
	if len(yamlVal) > 0 {
		return yamlVal
	}
	return strings.Split(fallback, ",")
}

// resolveBool 은 환경변수 > YAML 값 순으로 bool을 결정한다
func resolveBool(yamlVal bool, envKey string, yamlLoaded bool) bool {
	if envVal, ok := os.LookupEnv(envKey); ok {
		return envVal == "true"
	}
	if yamlLoaded {
		return yamlVal
	}
	return false
}
