---
name: commit-message
description: Generates clear commit messages from git diffs. Use when writing commit messages or reviewing staged changes.
---

# Commit Message 생성

1. `git diff --staged`로 변경사항 확인 (staged 없으면 `git diff`도 확인)
2. `git log --oneline -5`로 최근 커밋 스타일 확인
3. `.claude/rules/common/commit-message-conventions.md`의 규칙에 따라 메시지 작성
4. 사용자 확인 없이 바로 커밋 실행
