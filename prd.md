# PRD: KCP CLI

## 1. 프로젝트 개요

| 항목 | 내용 |
|---|---|
| 서비스명 | KCP CLI |
| 한 줄 설명 | OpenStack 인프라를 CLI와 웹콘솔에서 통합 관리하는 관리자 도구 |
| 핵심 가치 | 단일 Gateway를 통한 일관된 OpenStack 운영 경험 |

### 프로젝트 구조

| 프로젝트 | 설명 |
|---|---|
| kcp-cli | CLI 앱 (내부 모듈: kcp-sdk, kcp-cli, kcp-tui) |
| kcp-gateway | Backend API Gateway |
| kcp-webui | 웹 관리 콘솔 |

### CLI 내부 모듈 구조

| 모듈 | 역할 |
|---|---|
| kcp-sdk | OpenStack API 통신 핵심 라이브러리 (재사용 가능) |
| kcp-cli | 명령줄 인터페이스 (kcp-sdk 사용) |
| kcp-tui | 터미널 UI (kcp-sdk 사용) |

---

## 2. 타겟 유저 및 사용자 시나리오

### 타겟 유저

- **인프라 운영 관리자** (단일 역할)
  - OpenStack 기반 인프라의 전체 리소스를 운영/관리하는 담당자

### 사용자 시나리오

1. 관리자는 CLI에서 `kcp vm list` 명령으로 전체 VM 현황을 즉시 확인한다
2. 관리자는 TUI에서 인터랙티브하게 VM을 선택하고 상세 정보를 조회한다
3. 관리자는 웹콘솔 대시보드에서 전체 리소스 현황을 한눈에 파악한다
4. 관리자는 CLI와 웹콘솔에서 동일한 Gateway API를 사용하여 일관된 결과를 얻는다
5. 관리자는 감사 로그에서 누가 어떤 작업을 했는지 추적한다

---

## 3. 핵심 기능 명세

### 필수 기능 (MVP)

#### 1. Compute (Nova) 관리
- VM 인스턴스 목록 조회 / 상세 조회
- VM 생성 / 삭제 / 시작 / 중지 / 재부팅
- Flavor 목록 조회 / 생성 / 삭제

#### 2. Network (Neutron) 관리
- 네트워크 목록 조회 / 생성 / 삭제
- 서브넷 목록 조회 / 생성 / 삭제
- 라우터 목록 조회 / 생성 / 삭제 / 인터페이스 연결
- 보안그룹 목록 조회 / 생성 / 삭제 / 규칙 관리

#### 3. Storage (Cinder) 관리
- 볼륨 목록 조회 / 생성 / 삭제
- 볼륨 연결 / 분리 (VM에)
- 스냅샷 생성 / 삭제 / 목록 조회

#### 4. Identity (Keystone) 관리
- 프로젝트 목록 조회 / 생성 / 삭제
- 사용자 목록 조회 / 생성 / 삭제
- 역할 할당 / 해제

#### 5. Image (Glance) 관리
- OS 이미지 목록 조회 / 상세 조회
- 이미지 업로드 / 삭제

#### 6. 인증 시스템
- CLI 로그인 / 로그아웃
- 웹콘솔 로그인 / 로그아웃
- JWT 토큰, 세션 기반, OAuth2 모두 지원
- CLI 설정 파일: `~/.kcp/config` (기본)
- 실행 인자(`--config`) 및 환경변수(`KCP_CONFIG`)로 설정 파일 오버라이드

#### 7. Gateway API
- CLI와 WebUI가 동일한 API를 사용
- 기본은 OpenStack API 단순 프록시
- 필요 시 여러 OpenStack API를 조합하는 비즈니스 로직 지원
- OpenStack API 호출 실패 시 3회 재시도 + 지수 백오프

#### 8. 감사 로그
- 관리자의 모든 작업 이력 기록
- 감사 로그 조회 (WebUI + CLI)

### 선택 기능 (추후 확장)

1. **다중 OpenStack 클러스터 관리** — 여러 OpenStack 환경을 하나의 Gateway에서 관리
2. **작업 자동화** — CLI 스크립트/배치 작업 지원
3. **알림 시스템** — 리소스 이상 감지 시 알림

---

## 4. 기술 스택

| 구분 | 기술 | 비고 |
|---|---|---|
| CLI / SDK / TUI | Go | kcp-sdk를 핵심 라이브러리로 공유 |
| Gateway | Go + Gin | OpenStack API 프록시 + 비즈니스 로직 |
| WebUI | React + Next.js + Tailwind CSS + shadcn/ui | 다크 모드 기본 |
| 데이터베이스 | PostgreSQL | nerdctl 컨테이너로 실행 |
| OpenStack 연동 | kcp-sdk (Go) | Gateway에서 kcp-sdk를 사용하여 OpenStack API 호출 |

---

## 5. 주요 화면 구성 및 User Flow

### 화면 목록

1. **로그인 페이지** — 인증 (JWT / 세션 / OAuth2)
2. **대시보드** — 리소스 현황 요약 (VM 수, 네트워크 수, 볼륨 수 등)
3. **Compute 관리** — VM 목록 / 상세 / 생성 / 삭제
4. **Network 관리** — 네트워크 / 서브넷 / 라우터 / 보안그룹
5. **Storage 관리** — 볼륨 / 스냅샷
6. **Identity 관리** — 프로젝트 / 사용자
7. **Image 관리** — OS 이미지 목록 / 업로드
8. **감사 로그** — 관리자 작업 이력 조회

### User Flow

```
로그인 → 대시보드 → 각 리소스 관리 화면 → 작업 수행 → 결과 확인
                                                    ↓
                                              감사 로그에 기록
```

### CLI Flow

```
kcp login → kcp <resource> <action> → Gateway API 호출 → 결과 출력
                                          ↓
                                    감사 로그에 기록
```

---

## 6. 상세 기능 명세 및 예외 처리

### 인증 흐름

- **CLI 로그인**: `kcp login` → 자격 증명 입력 → Gateway 인증 API 호출 → 토큰을 `~/.kcp/config`에 저장
- **설정 파일 우선순위**: 실행 인자(`--config`) > 환경변수(`KCP_CONFIG`) > 기본값(`~/.kcp/config`)
- **WebUI 로그인**: 로그인 페이지 → Gateway 인증 API 호출 → JWT/세션 저장
- **토큰 만료**: 자동 갱신 시도 → 실패 시 재로그인 요구

### Gateway → OpenStack 연동

- **OpenStack 인증 정보**: Gateway 서버 설정 파일에 사전 구성 (Keystone endpoint, admin credential)
- **기본 동작**: OpenStack API 단순 프록시 (요청 변환 → 전달 → 응답 반환)
- **조합 로직**: 필요 시 여러 OpenStack API를 순차/병렬 호출하여 결과 조합
- **재시도 정책**: API 호출 실패 시 3회 재시도 + 지수 백오프
- **예외 처리**:
  - OpenStack API 타임아웃 → 클라이언트에 504 Gateway Timeout 반환
  - OpenStack 인증 실패 → 클라이언트에 502 Bad Gateway + 상세 에러 메시지 반환
  - OpenStack 서비스 다운 → 클라이언트에 503 Service Unavailable 반환

### 감사 로그

- 기록 항목: 사용자, 작업 종류, 대상 리소스, 시각, 결과(성공/실패), 요청 IP
- 보관 기간: 1년 (이후 자동 삭제)
- 조회: WebUI 화면 + CLI `kcp audit list` 명령

---

## 7. UI/UX 디자인 가이드

- **테마**: 다크 모드
- **톤앤매너**: 관리 도구 — 깔끔하고 정보 밀도가 높은 스타일
- **UI 라이브러리**: shadcn/ui + Tailwind CSS
- **반응형**: 데스크탑 전용
- **주요 컬러**: 다크 배경 기반, 상태 표시에 시맨틱 컬러 사용 (성공=녹색, 경고=노란색, 오류=빨간색)
- **데이터 표시**: 테이블 중심 레이아웃, 페이지네이션, 검색/필터 기능

---

## 8. 제약사항 및 보안 요구사항

### 제약사항

- **OpenStack 버전**: Epoxy 2025.1
- **동시 접속자**: 최대 20명
- **PostgreSQL**: nerdctl 컨테이너로 실행
- **CLI 설정 파일**: `~/.kcp/config` 기본 경로

### 보안 요구사항

- **HTTPS**: Gateway ↔ OpenStack API 통신 — 선택 가능 (설정으로 on/off)
- **HTTPS**: CLI/WebUI ↔ Gateway 통신 — 선택 가능 (설정으로 on/off)
- **인증**: JWT 토큰 / 세션 기반 / OAuth2 모두 지원
- **OpenStack 인증 정보**: Gateway 서버 설정에 사전 구성 (소스코드 하드코딩 금지)
- **감사 로그**: 모든 관리자 작업 이력 기록, 1년 보관
- **토큰 관리**: CLI 설정 파일의 토큰은 파일 권한 600으로 보호
