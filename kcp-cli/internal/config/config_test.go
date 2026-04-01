package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolvePath(t *testing.T) {
	// 플래그 우선순위 테스트
	result := ResolvePath("/custom/path")
	if result != "/custom/path" {
		t.Errorf("ResolvePath with flag = %s, want /custom/path", result)
	}

	// 환경변수 우선순위 테스트
	os.Setenv("KCP_CONFIG", "/env/path")
	defer os.Unsetenv("KCP_CONFIG")
	result = ResolvePath("")
	if result != "/env/path" {
		t.Errorf("ResolvePath with env = %s, want /env/path", result)
	}

	// 기본값 테스트
	os.Unsetenv("KCP_CONFIG")
	result = ResolvePath("")
	expected := DefaultConfigPath()
	if result != expected {
		t.Errorf("ResolvePath default = %s, want %s", result, expected)
	}
}

func TestDefaultConfigPath(t *testing.T) {
	path := DefaultConfigPath()
	if !strings.HasSuffix(path, filepath.Join(".kcp", "kcp-config.yaml")) {
		t.Errorf("DefaultConfigPath = %s, want suffix .kcp/kcp-config.yaml", path)
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".kcp", "kcp-config.yaml")

	cfg := &CLIConfig{
		ServerURL: "http://localhost:8080/api/v1",
		Token:     "test-token",
		AuthType:  "JWT",
	}

	// 저장
	err := Save(path, cfg)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// 파일 권한 확인 (600)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("file permission = %o, want 600", info.Mode().Perm())
	}

	// YAML 형식 확인
	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.Contains(content, "server_url:") {
		t.Errorf("파일이 YAML 형식이 아닙니다: %s", content)
	}

	// 로드
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.ServerURL != cfg.ServerURL {
		t.Errorf("ServerURL = %s, want %s", loaded.ServerURL, cfg.ServerURL)
	}
	if loaded.Token != cfg.Token {
		t.Errorf("Token = %s, want %s", loaded.Token, cfg.Token)
	}
	if loaded.AuthType != cfg.AuthType {
		t.Errorf("AuthType = %s, want %s", loaded.AuthType, cfg.AuthType)
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	// 존재하지 않는 파일 → 기본값 반환
	cfg, err := Load("/tmp/nonexistent-kcp-config.yaml")
	if err != nil {
		t.Fatalf("Load should not fail for missing file: %v", err)
	}
	if cfg.ServerURL != "http://localhost:8080/api/v1" {
		t.Errorf("ServerURL = %s, want default", cfg.ServerURL)
	}
	if cfg.AuthType != "JWT" {
		t.Errorf("AuthType = %s, want JWT", cfg.AuthType)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.ServerURL != "http://localhost:8080/api/v1" {
		t.Errorf("ServerURL = %s, want http://localhost:8080/api/v1", cfg.ServerURL)
	}
	if cfg.AuthType != "JWT" {
		t.Errorf("AuthType = %s, want JWT", cfg.AuthType)
	}
}

func TestInitConfig(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".kcp", "kcp-config.yaml")

	// 파일이 없을 때 생성
	err := InitConfig(path)
	if err != nil {
		t.Fatalf("InitConfig failed: %v", err)
	}

	// 파일 존재 확인
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("InitConfig did not create file")
	}

	// 다시 호출해도 에러 없음 (이미 존재)
	err = InitConfig(path)
	if err != nil {
		t.Fatalf("InitConfig should not fail for existing file: %v", err)
	}
}
