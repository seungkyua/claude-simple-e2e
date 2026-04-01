package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadEnvFile 는 임시 .env 파일을 생성하고 로드하여 환경변수가 설정되는지 검증한다
func TestLoadEnvFile(t *testing.T) {
	// 임시 디렉토리에 .env 파일 생성
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")

	content := "TEST_CONFIG_KEY=hello_world\n# 주석은 무시\n\nTEST_CONFIG_KEY2=value2\n"
	if err := os.WriteFile(envPath, []byte(content), 0600); err != nil {
		t.Fatalf(".env 파일 생성 실패: %v", err)
	}

	// 테스트 전 환경변수 정리
	os.Unsetenv("TEST_CONFIG_KEY")
	os.Unsetenv("TEST_CONFIG_KEY2")
	defer os.Unsetenv("TEST_CONFIG_KEY")
	defer os.Unsetenv("TEST_CONFIG_KEY2")

	loadEnvFile(envPath)

	if got := os.Getenv("TEST_CONFIG_KEY"); got != "hello_world" {
		t.Errorf("TEST_CONFIG_KEY = %q, 기대값 %q", got, "hello_world")
	}
	if got := os.Getenv("TEST_CONFIG_KEY2"); got != "value2" {
		t.Errorf("TEST_CONFIG_KEY2 = %q, 기대값 %q", got, "value2")
	}
}

// TestLoadEnvFileDoesNotOverride 는 이미 설정된 환경변수를 .env 파일이 덮어쓰지 않는지 검증한다
func TestLoadEnvFileDoesNotOverride(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")

	content := "TEST_NO_OVERRIDE=from_file\n"
	if err := os.WriteFile(envPath, []byte(content), 0600); err != nil {
		t.Fatalf(".env 파일 생성 실패: %v", err)
	}

	// 시스템 환경변수를 먼저 설정
	os.Setenv("TEST_NO_OVERRIDE", "from_system")
	defer os.Unsetenv("TEST_NO_OVERRIDE")

	loadEnvFile(envPath)

	if got := os.Getenv("TEST_NO_OVERRIDE"); got != "from_system" {
		t.Errorf("환경변수가 덮어써짐: got %q, 기대값 %q", got, "from_system")
	}
}

// TestGetEnv 는 환경변수가 없을 때 기본값을 반환하는지 검증한다
func TestGetEnv(t *testing.T) {
	os.Unsetenv("TEST_GETENV_MISSING")

	// 환경변수가 없으면 기본값 반환
	if got := getEnv("TEST_GETENV_MISSING", "default_val"); got != "default_val" {
		t.Errorf("기본값 반환 실패: got %q, 기대값 %q", got, "default_val")
	}

	// 환경변수가 있으면 해당 값 반환
	os.Setenv("TEST_GETENV_EXISTS", "real_val")
	defer os.Unsetenv("TEST_GETENV_EXISTS")

	if got := getEnv("TEST_GETENV_EXISTS", "default_val"); got != "real_val" {
		t.Errorf("환경변수 값 반환 실패: got %q, 기대값 %q", got, "real_val")
	}
}

// TestLoadMissingRequired 는 필수 환경변수(DATABASE_URL)가 비어있을 때 오류를 반환하는지 검증한다
func TestLoadMissingRequired(t *testing.T) {
	// 모든 필수 환경변수를 초기화
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("OPENSTACK_AUTH_URL")

	_, err := Load()
	if err == nil {
		t.Fatal("필수 환경변수 누락 시 오류가 반환되어야 합니다")
	}

	// DATABASE_URL이 첫 번째 검증이므로 해당 메시지 확인
	expected := "DATABASE_URL 환경변수가 설정되지 않았습니다"
	if err.Error() != expected {
		t.Errorf("오류 메시지 불일치: got %q, 기대값 %q", err.Error(), expected)
	}
}
