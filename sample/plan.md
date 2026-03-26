# 코드 생성 계획 — AnonVoice

## 프로젝트 정보
- **프로젝트명**: AnonVoice
- **코드 위치**: `src/`
- **기술 스택**: TypeScript (React 18 + Next.js 14, Node.js + Express, PostgreSQL 16 + Prisma) + Vitest
- **TDD 적용**: Red → Green → Refactor

## 코드 생성 순서

### Phase 1: 프로젝트 구조 및 기반

- [ ] Step 1: 프로젝트 초기화 — 모노레포 디렉토리 구조, TypeScript 설정, Vitest 설정, ESLint/Prettier
- [ ] Step 2: 공통 타입 및 상수 정의 — Role(EMPLOYEE/TEAM_LEADER/HR_ADMIN), Category(CULTURE/PROCESS/LEADERSHIP/OTHER), Sentiment(POSITIVE/NEUTRAL/NEGATIVE) enum, 공통 에러 클래스
- [ ] Step 3: 데이터베이스 스키마 구현 — Prisma schema 정의 (users, teams, feedbacks, conversations, notifications, audit_logs, risk_alerts), 마이그레이션

### Phase 2: 인증 및 사용자 관리 (src/auth/)

- [ ] Step 4: SSO 인증 연동 구현 — NextAuth.js 설정, 사내 SSO provider, JWT 세션 관리
- [ ] Step 5: RBAC 미들웨어 구현 — 역할 기반 접근 제어, 팀 소속 검증 미들웨어, 인증 토큰 검증
- [ ] Step 6: 감사 로그 미들웨어 구현 — 관리자 행위 기록(조회/답변), 일반 직원 피드백 작성은 기록 안 함

### Phase 3: 익명 피드백 핵심 로직 (src/feedback/)

- [ ] Step 7: 피드백 생성 구현 — createFeedback 서비스, 익명 토큰 생성(crypto.randomBytes + SHA-256 해시), userId 미저장으로 익명성 보장
- [ ] Step 8: 피드백 입력 검증 구현 — 카테고리/팀/감정태그 enum 검증, 본문 길이(10~1000자), 대상 팀 존재 확인
- [ ] Step 9: 피드백 조회 구현 — 팀별 피드백 목록(최신순/카테고리 필터), 관리자는 자기 팀만 열람 가능, 페이지네이션
- [ ] Step 10: 피드백 불변성 구현 — 제출 후 수정/삭제 불가 정책, 2년 보관 후 자동 삭제(expiresAt)

### Phase 4: 익명 대화 시스템 (src/conversation/)

- [ ] Step 11: 관리자 답변 구현 — 피드백에 답변 작성, 팀 소속 검증, 감사 로그 기록
- [ ] Step 12: 익명 후속 코멘트 구현 — 익명 토큰 해시 비교로 본인 확인, authorType(ANONYMOUS_AUTHOR/MANAGER) 구분
- [ ] Step 13: 대화 제한 구현 — 최대 5회 왕복(10개 코멘트) 제한, 쓰레드 형태 대화 구조

### Phase 5: 알림 시스템 (src/notification/)

- [ ] Step 14: 인앱 알림 구현 — Notification 모델 CRUD, 읽음 처리, 미읽은 알림 수 조회
- [ ] Step 15: 이메일 알림 구현 — 새 피드백 시 관리자 이메일, 답변 시 작성자 인앱 알림(익명성 유지)
- [ ] Step 16: 위험 키워드 감지 구현 — 키워드 매칭 엔진, 심각도 분류, HR에게 즉시 알림(이메일 + 인앱)

### Phase 6: 통계 및 리포트 (src/stats/)

- [ ] Step 17: 팀별 통계 구현 — 감정 분포 집계, 월별 트렌드 데이터, 카테고리별 분류
- [ ] Step 18: 전사 통계 구현 — HR 전용 전사 대시보드 데이터, 팀 간 비교, 참여율 산출
- [ ] Step 19: 리포트 생성 구현 — 분기별 자동 PDF 리포트, 키워드 워드클라우드 데이터

### Phase 7: API 라우트 (src/routes/)

- [ ] Step 20: 피드백 API 구현 — POST /api/feedbacks, GET /api/feedbacks, Request/Response 스키마, Rate Limiting
- [ ] Step 21: 대화 API 구현 — POST /api/feedbacks/:id/comments, GET /api/feedbacks/:id/comments
- [ ] Step 22: 알림 API 구현 — GET /api/notifications, PATCH /api/notifications/:id/read
- [ ] Step 23: 통계/리포트 API 구현 — GET /api/stats/team/:id, GET /api/stats/company, GET /api/reports

### Phase 8: 프론트엔드 — 인증 및 레이아웃 (src/app/)

- [ ] Step 24: 로그인 페이지 구현 — SSO 로그인 버튼, NextAuth.js 클라이언트 연동, 인증 상태 관리
- [ ] Step 25: 공통 레이아웃 구현 — 인증 필수 레이아웃, 역할별 네비게이션, 알림 아이콘(미읽음 배지)

### Phase 9: 프론트엔드 — 피드백 작성/조회 (src/app/feedback/)

- [ ] Step 26: 피드백 작성 폼 구현 — 카테고리/팀 선택, 본문 입력(글자 수 표시), 감정 태그 선택, 제출 후 익명 토큰 localStorage 저장
- [ ] Step 27: 내 피드백 목록 구현 — 익명 토큰 기반 피드백 조회, 답변 확인(쓰레드 UI), 후속 코멘트 작성

### Phase 10: 프론트엔드 — 관리자 대시보드 (src/app/dashboard/)

- [ ] Step 28: 팀 리더 대시보드 구현 — 피드백 목록(필터/정렬), 감정 분포 차트, 월별 트렌드 그래프
- [ ] Step 29: 답변 기능 구현 — 피드백 상세 + 답변 작성 폼, 쓰레드 대화 UI
- [ ] Step 30: HR 대시보드 구현 — 전사 통계 차트, 위험 알림 목록, 리포트 다운로드

### Phase 11: 보안 강화 및 인프라

- [ ] Step 31: 보안 미들웨어 통합 — helmet 보안 헤더, CORS 설정, Rate Limiting(피드백 10회/24h, 인증 5회/15m), CSRF 방지
- [ ] Step 32: 데이터 암호화 구현 — 피드백 본문 AES-256-GCM 암호화, 환경변수 기반 키 관리, 민감 데이터 로그 마스킹
- [ ] Step 33: 데이터 생명주기 구현 — 만료 피드백 자동 삭제 스케줄러, 감사 로그 보관 정책, 알림 90일 자동 삭제

## 스토리 매핑

| 기능 요구사항 | 구현 Step |
|---|---|
| 익명 피드백 작성 (카테고리/팀/본문/감정태그) | Step 7-8, 20, 26 |
| 제출 후 수정/삭제 불가 (익명성 보장) | Step 10 |
| 피드백 대시보드 (목록/차트/트렌드/워드클라우드) | Step 9, 17, 28 |
| 익명 대화 시스템 (답변/후속 코멘트/5회 제한) | Step 11-13, 21, 27, 29 |
| 새 피드백 알림 (관리자 이메일) | Step 14-15, 22 |
| 답변 알림 (작성자 인앱) | Step 14-15, 22 |
| 위험 키워드 감지 알림 (HR 즉시) | Step 16 |
| 통계 및 리포트 (분기 PDF/팀 비교/참여율) | Step 17-19, 23, 30 |
| SSO 인증 | Step 4, 24 |
| RBAC 접근 통제 (관리자 자기 팀만) | Step 5-6, 25 |
| 데이터 보관 2년 후 자동 삭제 | Step 33 |
| 감사 로그 (관리자 조회/답변 기록) | Step 6 |
| 데이터 암호화 및 보안 헤더 | Step 31-32 |
