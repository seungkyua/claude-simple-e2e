package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestResolveStr 은 환경변수 > YAML > 기본값 우선순위를 검증한다
func TestResolveStr(t *testing.T) {
	os.Unsetenv("TEST_RESOLVE")

	// YAML 값이 있고 환경변수가 없으면 YAML 값 사용
	if got := resolveStr("yaml_val", "TEST_RESOLVE", "default"); got != "yaml_val" {
		t.Errorf("got %q, want yaml_val", got)
	}

	// YAML 값이 비어있으면 기본값 사용
	if got := resolveStr("", "TEST_RESOLVE", "default"); got != "default" {
		t.Errorf("got %q, want default", got)
	}

	// 환경변수가 있으면 환경변수 우선
	os.Setenv("TEST_RESOLVE", "env_val")
	defer os.Unsetenv("TEST_RESOLVE")
	if got := resolveStr("yaml_val", "TEST_RESOLVE", "default"); got != "env_val" {
		t.Errorf("got %q, want env_val", got)
	}
}

// TestLoadFromYAML 은 YAML 파일에서 설정을 로드하는지 검증한다
func TestLoadFromYAML(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "kcp-gateway-config.yaml")

	content := `
server:
  port: "9090"
  allowed_origins:
    - "http://localhost:3000"
    - "http://localhost:4000"
database:
  url: "postgresql://test:test@localhost:5432/test"
jwt:
  secret: "test-secret"
  expiry: "30m"
openstack:
  auth_url: "http://keystone:5000/v3"
  username: "admin"
  password: "secret"
  project_id: "proj-123"
  domain_name: "Default"
  region_name: "RegionOne"
tls:
  enabled: false
`
	os.WriteFile(cfgPath, []byte(content), 0600)

	// 환경변수가 설정되어 있지 않은 상태에서 YAML 로드
	for _, key := range []string{"PORT", "DATABASE_URL", "JWT_SECRET", "OS_AUTH_URL", "OS_USERNAME", "OS_PASSWORD"} {
		os.Unsetenv(key)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Port != "9090" {
		t.Errorf("Port = %s, want 9090", cfg.Port)
	}
	if cfg.DatabaseURL != "postgresql://test:test@localhost:5432/test" {
		t.Errorf("DatabaseURL = %s, want test URL", cfg.DatabaseURL)
	}
	if cfg.OpenStackAuthURL != "http://keystone:5000/v3" {
		t.Errorf("OpenStackAuthURL = %s, want http://keystone:5000/v3", cfg.OpenStackAuthURL)
	}
	if len(cfg.AllowedOrigins) != 2 {
		t.Errorf("AllowedOrigins count = %d, want 2", len(cfg.AllowedOrigins))
	}
}

// TestEnvOverridesYAML 은 환경변수가 YAML 값을 덮어쓰는지 검증한다
func TestEnvOverridesYAML(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "kcp-gateway-config.yaml")

	content := `
database:
  url: "postgresql://yaml:yaml@localhost/yaml"
jwt:
  secret: "yaml-secret"
openstack:
  auth_url: "http://yaml-keystone:5000/v3"
  username: "yaml-user"
  password: "yaml-pass"
`
	os.WriteFile(cfgPath, []byte(content), 0600)

	// 환경변수로 덮어쓰기
	os.Setenv("DATABASE_URL", "postgresql://env:env@localhost/env")
	os.Setenv("OS_AUTH_URL", "http://env-keystone:5000/v3")
	defer os.Unsetenv("DATABASE_URL")
	defer os.Unsetenv("OS_AUTH_URL")

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.DatabaseURL != "postgresql://env:env@localhost/env" {
		t.Errorf("DatabaseURL = %s, want env value", cfg.DatabaseURL)
	}
	if cfg.OpenStackAuthURL != "http://env-keystone:5000/v3" {
		t.Errorf("OpenStackAuthURL = %s, want env value", cfg.OpenStackAuthURL)
	}
}

// TestLoadFromEnvOnly 는 YAML 파일 없이 환경변수만으로 설정을 로드하는지 검증한다
func TestLoadFromEnvOnly(t *testing.T) {
	// 존재하지 않는 기본 경로 → 환경변수 fallback
	os.Setenv("DATABASE_URL", "postgresql://env:env@localhost/env")
	os.Setenv("JWT_SECRET", "env-secret")
	os.Setenv("OS_AUTH_URL", "http://env-keystone:5000/v3")
	os.Setenv("OS_USERNAME", "env-user")
	os.Setenv("OS_PASSWORD", "env-pass")
	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("OS_AUTH_URL")
		os.Unsetenv("OS_USERNAME")
		os.Unsetenv("OS_PASSWORD")
	}()

	// 기본 파일이 없는 디렉토리로 이동해서 테스트
	cfg, err := Load("/tmp/nonexistent-dir/kcp-gateway-config.yaml")
	// 명시적 경로 지정 시 파일 없으면 에러
	if err == nil {
		t.Fatal("명시적 경로의 파일이 없으면 에러가 반환되어야 합니다")
	}

	// 빈 문자열 (기본 파일) → 파일 없으면 환경변수 사용
	cfg, err = Load("")
	if err != nil {
		t.Fatalf("환경변수만으로 Load 실패: %v", err)
	}
	if cfg.DatabaseURL != "postgresql://env:env@localhost/env" {
		t.Errorf("DatabaseURL = %s, want env value", cfg.DatabaseURL)
	}
}

// TestLoadMissingRequired 는 필수 값이 없을 때 에러를 반환하는지 검증한다
func TestLoadMissingRequired(t *testing.T) {
	for _, key := range []string{"DATABASE_URL", "JWT_SECRET", "OS_AUTH_URL", "OS_USERNAME", "OS_PASSWORD", "PORT"} {
		os.Unsetenv(key)
	}

	_, err := Load("")
	if err == nil {
		t.Fatal("필수 값 누락 시 에러가 반환되어야 합니다")
	}
}

// TestExplicitConfigNotFound 는 명시적으로 지정한 설정 파일이 없을 때 에러를 반환하는지 검증한다
func TestExplicitConfigNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("존재하지 않는 경로 지정 시 에러가 반환되어야 합니다")
	}
}
