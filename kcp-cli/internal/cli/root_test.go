package cli

import (
	"testing"
)

// TestRootCommand 는 루트 커맨드가 올바르게 정의되어 있는지 검증한다
func TestRootCommand(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd가 nil이다")
	}

	// Use 필드가 "kcp"인지 확인한다
	if rootCmd.Use != "kcp" {
		t.Errorf("rootCmd.Use가 'kcp'이어야 하지만 '%s'이다", rootCmd.Use)
	}

	// Short 설명이 비어있지 않은지 확인한다
	if rootCmd.Short == "" {
		t.Error("rootCmd.Short가 비어있으면 안 된다")
	}
}

// TestGlobalFlags 는 글로벌 플래그가 올바르게 등록되어 있는지 검증한다
func TestGlobalFlags(t *testing.T) {
	// --config 플래그 존재 확인
	configFlag := rootCmd.PersistentFlags().Lookup("config")
	if configFlag == nil {
		t.Fatal("--config 플래그가 등록되어 있지 않다")
	}
	if configFlag.DefValue != "" {
		t.Errorf("--config 기본값이 비어있어야 하지만 '%s'이다", configFlag.DefValue)
	}

	// --output 플래그 존재 확인
	outputFlag := rootCmd.PersistentFlags().Lookup("output")
	if outputFlag == nil {
		t.Fatal("--output 플래그가 등록되어 있지 않다")
	}
	if outputFlag.DefValue != "table" {
		t.Errorf("--output 기본값이 'table'이어야 하지만 '%s'이다", outputFlag.DefValue)
	}

	// --output 단축키가 'o'인지 확인한다
	if outputFlag.Shorthand != "o" {
		t.Errorf("--output 단축키가 'o'이어야 하지만 '%s'이다", outputFlag.Shorthand)
	}
}
