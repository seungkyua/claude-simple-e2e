package service

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// TestHashPassword 는 비밀번호를 해싱하고 bcrypt 비교가 성공하는지 검증한다
func TestHashPassword(t *testing.T) {
	password := "secureP@ssw0rd!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword 실패: %v", err)
	}

	// 해시가 비어있지 않은지 확인
	if hash == "" {
		t.Fatal("해시가 빈 문자열입니다")
	}

	// 원본 비밀번호와 해시 비교
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		t.Errorf("해시와 원본 비밀번호 비교 실패: %v", err)
	}

	// 잘못된 비밀번호로 비교 시 실패 확인
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("wrongpassword")); err == nil {
		t.Error("잘못된 비밀번호인데 비교가 성공함")
	}
}

// TestHashPasswordProducesDifferentHashes 는 동일한 입력에 대해 다른 해시를 생성하는지 검증한다 (솔트 적용)
func TestHashPasswordProducesDifferentHashes(t *testing.T) {
	password := "samePassword123!"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("첫 번째 HashPassword 실패: %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("두 번째 HashPassword 실패: %v", err)
	}

	if hash1 == hash2 {
		t.Error("동일한 비밀번호에 대해 같은 해시가 생성됨 — 솔트가 적용되지 않았습니다")
	}
}
