# 코드 생성 계획 — KCP CLI

## 프로젝트 정보
- **프로젝트명**: KCP CLI
- **코드 위치**: `kcp-cli/`, `kcp-gateway/`, `kcp-webui/`
- **기술 스택**: Go + go test (CLI/SDK/TUI/Gateway), React + Next.js + Vitest (WebUI)
- **TDD 적용**: Red → Green → Refactor

## 코드 생성 순서

---

## 1단계: 프로젝트 환경 초기화 및 구조 세팅

### Phase 1: Go 프로젝트 초기화 (kcp-cli/, kcp-gateway/)

- [ ] Step 1: CLI 프로젝트 초기화 — kcp-cli Go module 생성 (go.mod), 내부 모듈 디렉토리 구조 (cmd/cli/, cmd/tui/, pkg/sdk/)
- [ ] Step 2: Gateway 프로젝트 초기화 — kcp-gateway Go module 생성 (go.mod), Gin 의존성, 디렉토리 구조 (cmd/server/, internal/handler/, internal/service/, internal/middleware/, internal/model/)
- [ ] Step 3: 공통 설정 파일 — Makefile, .gitignore, .editorconfig, golangci-lint 설정

### Phase 2: WebUI 프로젝트 초기화 (kcp-webui/)

- [ ] Step 4: Next.js 프로젝트 생성 — create-next-app, Tailwind CSS 설정, shadcn/ui 초기화, 다크 모드 테마 설정
- [ ] Step 5: WebUI 기본 설정 — ESLint, Prettier, tsconfig, Vitest 설정, 디렉토리 구조 (app/, components/, hooks/, services/, stores/, types/)

### Phase 3: 인프라 및 실행 환경 (프로젝트 루트)

- [ ] Step 6: 컨테이너 환경 구성 — nerdctl용 PostgreSQL 컨테이너 구성 파일, 초기 SQL 스크립트
- [ ] Step 7: 환경변수 및 설정 템플릿 — .env.example (Gateway/WebUI), kcp-gateway 서버 설정 파일 템플릿 (OpenStack 인증 정보 포함)

---

## 2단계: 데이터베이스 연동 및 공통 인터페이스/유틸리티 구현

### Phase 4: 데이터베이스 스키마 및 연동 (kcp-gateway/)

- [ ] Step 8: DB 스키마 정의 — 사용자/세션 테이블, 감사 로그 테이블, 마이그레이션 파일 작성
- [ ] Step 9: DB 연결 모듈 구현 — PostgreSQL 연결 풀, 헬스체크, 마이그레이션 자동 실행

### Phase 5: kcp-sdk 핵심 타입 및 인터페이스 (kcp-cli/pkg/sdk/)

- [ ] Step 10: 공통 타입 정의 — OpenStack 리소스 타입 (Server, Network, Volume, Project, User, Image), 요청/응답 구조체
- [ ] Step 11: SDK 클라이언트 인터페이스 — ComputeClient, NetworkClient, StorageClient, IdentityClient, ImageClient 인터페이스 정의
- [ ] Step 12: HTTP 클라이언트 기반 모듈 — 공통 HTTP 클라이언트 (재시도 3회 + 지수 백오프, HTTPS 선택, 에러 핸들링)

### Phase 6: Gateway 공통 미들웨어 (kcp-gateway/internal/middleware/)

- [ ] Step 13: 인증 미들웨어 — JWT 토큰 검증, 세션 검증, OAuth2 검증, 인증 방식 자동 분기
- [ ] Step 14: 감사 로그 미들웨어 — 요청별 감사 로그 자동 기록 (사용자, 작업, 리소스, IP, 결과)
- [ ] Step 15: 에러 핸들링 미들웨어 — 공통 에러 응답 포맷, OpenStack 에러 변환 (504/502/503)

### Phase 7: CLI 공통 유틸리티 (kcp-cli/)

- [ ] Step 16: 설정 파일 관리 — ~/.kcp/config 읽기/쓰기, --config 플래그, KCP_CONFIG 환경변수, 우선순위 적용
- [ ] Step 17: 인증 모듈 — kcp login/logout 구현, 토큰 저장 (파일 권한 600), 토큰 자동 갱신

---

## 3단계: 핵심 비즈니스 로직 및 백엔드 API 구현

### Phase 8: kcp-sdk OpenStack 클라이언트 구현 (kcp-cli/pkg/sdk/)

- [ ] Step 18: Compute 클라이언트 — Nova API 연동 (VM CRUD, 시작/중지/재부팅, Flavor 관리)
- [ ] Step 19: Network 클라이언트 — Neutron API 연동 (네트워크/서브넷/라우터/보안그룹 CRUD)
- [ ] Step 20: Storage 클라이언트 — Cinder API 연동 (볼륨 CRUD, 연결/분리, 스냅샷)
- [ ] Step 21: Identity 클라이언트 — Keystone API 연동 (프로젝트/사용자 CRUD, 역할 할당)
- [ ] Step 22: Image 클라이언트 — Glance API 연동 (이미지 목록/상세/업로드/삭제)

### Phase 9: Gateway API 핸들러 (kcp-gateway/)

- [ ] Step 23: 인증 API — 로그인/로그아웃/토큰 갱신 엔드포인트, JWT/세션/OAuth2 지원
- [ ] Step 24: Compute API 핸들러 — /api/v1/compute/* 라우트, kcp-sdk ComputeClient 호출, 프록시 + 조합 로직
- [ ] Step 25: Network API 핸들러 — /api/v1/network/* 라우트, kcp-sdk NetworkClient 호출
- [ ] Step 26: Storage API 핸들러 — /api/v1/storage/* 라우트, kcp-sdk StorageClient 호출
- [ ] Step 27: Identity API 핸들러 — /api/v1/identity/* 라우트, kcp-sdk IdentityClient 호출
- [ ] Step 28: Image API 핸들러 — /api/v1/image/* 라우트, kcp-sdk ImageClient 호출
- [ ] Step 29: 감사 로그 API — /api/v1/audit/* 라우트, 감사 로그 조회/필터, 1년 보관 정책
- [ ] Step 30: 통계 API — /api/v1/stats/* 라우트, 대시보드용 리소스 현황 요약 데이터

### Phase 10: CLI 명령어 구현 (kcp-cli/cmd/cli/)

- [ ] Step 31: CLI 프레임워크 세팅 — cobra 기반 루트 커맨드, 글로벌 플래그 (--config, --output format)
- [ ] Step 32: Compute 명령어 — kcp vm list/show/create/delete/start/stop/reboot, kcp flavor list/create/delete
- [ ] Step 33: Network 명령어 — kcp network/subnet/router/secgroup CRUD 명령어
- [ ] Step 34: Storage 명령어 — kcp volume list/create/delete/attach/detach, kcp snapshot CRUD
- [ ] Step 35: Identity 명령어 — kcp project/user CRUD, kcp role assign/revoke
- [ ] Step 36: Image 명령어 — kcp image list/show/upload/delete
- [ ] Step 37: 감사 로그 명령어 — kcp audit list, 필터 옵션 (날짜, 사용자, 작업 유형)

### Phase 11: TUI 구현 (kcp-cli/cmd/tui/)

- [ ] Step 38: TUI 프레임워크 세팅 — bubbletea 기반 메인 모델, 네비게이션 구조
- [ ] Step 39: TUI 리소스 뷰 — VM/Network/Volume/Project/Image 목록 뷰, 상세 뷰, 인터랙티브 선택

---

## 4단계: 프론트엔드 주요 화면(UI) 컴포넌트 및 API 연동

### Phase 12: WebUI 공통 레이어 (kcp-webui/)

- [ ] Step 40: API 서비스 레이어 — Gateway API 통신 모듈 (axios), 인증 토큰 인터셉터, 에러 핸들링
- [ ] Step 41: 인증 스토어 및 페이지 — Zustand 인증 스토어, 로그인 페이지, 인증 가드
- [ ] Step 42: 공통 레이아웃 — Header, Sidebar, 다크 모드 레이아웃, 네비게이션

### Phase 13: WebUI 대시보드 (kcp-webui/)

- [ ] Step 43: 대시보드 페이지 — 리소스 현황 요약 카드 (VM 수, 네트워크 수, 볼륨 수), 통계 차트

### Phase 14: WebUI 리소스 관리 화면 (kcp-webui/)

- [ ] Step 44: Compute 관리 화면 — VM 목록 테이블 (검색/필터/페이지네이션), VM 상세/생성/삭제 다이얼로그
- [ ] Step 45: Network 관리 화면 — 네트워크/서브넷/라우터/보안그룹 목록, CRUD UI
- [ ] Step 46: Storage 관리 화면 — 볼륨/스냅샷 목록, 생성/삭제/연결 UI
- [ ] Step 47: Identity 관리 화면 — 프로젝트/사용자 목록, CRUD UI, 역할 할당
- [ ] Step 48: Image 관리 화면 — 이미지 목록, 업로드/삭제 UI

### Phase 15: WebUI 감사 로그 (kcp-webui/)

- [ ] Step 49: 감사 로그 화면 — 작업 이력 테이블, 날짜/사용자/작업 유형 필터, 페이지네이션

---

## 스토리 매핑

| 기능 요구사항 | 구현 Step |
|---|---|
| Compute (Nova) 관리 | Step 10-11, 18, 24, 32, 39, 44 |
| Network (Neutron) 관리 | Step 10-11, 19, 25, 33, 39, 45 |
| Storage (Cinder) 관리 | Step 10-11, 20, 26, 34, 39, 46 |
| Identity (Keystone) 관리 | Step 10-11, 21, 27, 35, 39, 47 |
| Image (Glance) 관리 | Step 10-11, 22, 28, 36, 39, 48 |
| 인증 시스템 | Step 13, 16-17, 23, 41 |
| Gateway API (프록시 + 비즈니스 로직) | Step 12, 15, 24-30 |
| 감사 로그 | Step 8-9, 14, 29, 37, 49 |
| 대시보드 | Step 30, 43 |
| CLI 명령어 | Step 31-37 |
| TUI | Step 38-39 |
| WebUI | Step 40-49 |
