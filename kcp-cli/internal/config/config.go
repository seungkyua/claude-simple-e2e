// Package config 는 KCP CLI의 설정 파일을 관리한다.
// 설정 파일 형식: YAML
// 우선순위: --config 플래그 > KCP_CONFIG 환경변수 > ~/.kcp/kcp-config.yaml
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// CLIConfig 는 CLI 설정 파일의 내용을 나타낸다
// username은 설정에 저장하지 않고, 로그인 시 항상 입력받는다.
type CLIConfig struct {
	ServerURL string `yaml:"server_url"` // Gateway API 서버 URL
	Token     string `yaml:"token"`      // 인증 토큰
	AuthType  string `yaml:"auth_type"`  // 인증 방식 (JWT, SESSION, OAUTH2)
}

// DefaultConfigPath 는 기본 설정 파일 경로를 반환한다
func DefaultConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".kcp", "kcp-config.yaml")
}

// ResolvePath 는 설정 파일 경로를 우선순위에 따라 결정한다
// 우선순위: flagPath > KCP_CONFIG 환경변수 > 기본 경로
func ResolvePath(flagPath string) string {
	if flagPath != "" {
		return flagPath
	}
	if envPath := os.Getenv("KCP_CONFIG"); envPath != "" {
		return envPath
	}
	return DefaultConfigPath()
}

// Load 는 설정 파일을 읽어 CLIConfig를 반환한다.
// 파일이 없으면 기본값으로 생성한다.
func Load(path string) (*CLIConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// 설정 파일이 없으면 기본값 반환
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("설정 파일 읽기 실패 (%s): %w", path, err)
	}

	var cfg CLIConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("설정 파일 파싱 실패: %w", err)
	}

	return &cfg, nil
}

// Save 는 CLIConfig를 YAML 설정 파일에 저장한다 (파일 권한 600)
func Save(path string, cfg *CLIConfig) error {
	// 디렉토리 생성
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("설정 디렉토리 생성 실패: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("설정 직렬화 실패: %w", err)
	}

	// 파일 권한 600으로 저장 (소유자만 읽기/쓰기)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("설정 파일 저장 실패: %w", err)
	}

	return nil
}

// DefaultConfig 는 기본 설정값을 반환한다
func DefaultConfig() *CLIConfig {
	return &CLIConfig{
		ServerURL: "http://localhost:8080/api/v1",
		AuthType:  "JWT",
	}
}

// InitConfig 는 기본 설정 파일이 없으면 생성한다
func InitConfig(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil // 이미 존재
	}

	cfg := DefaultConfig()
	if err := Save(path, cfg); err != nil {
		return err
	}

	fmt.Printf("설정 파일 생성: %s\n", path)
	return nil
}
