# 시스템 설계서: KCP CLI

## 1. 프로젝트 개요

| 항목 | 내용 |
|---|---|
| 프로젝트명 | KCP CLI |
| 한 줄 설명 | OpenStack 인프라를 CLI와 웹콘솔에서 통합 관리하는 관리자 도구 |
| 기술 스택 | Go + Gin (CLI/SDK/TUI/Gateway), React + Next.js + Tailwind CSS + shadcn/ui (WebUI), PostgreSQL (nerdctl) |
| 작성일 | 2026-03-30 |
| 기반 문서 | prd.md |
| 대상 OpenStack | Epoxy 2025.1 |

---

## 2. 전체 시스템 아키텍처

```mermaid
flowchart TB
    subgraph Client["클라이언트"]
        CLI["kcp-cli\n(Go, cobra)"]
        TUI["kcp-tui\n(Go, bubbletea)"]
        WebUI["kcp-webui\n(Next.js)"]
    end

    subgraph SDK["kcp-sdk (Go 공유 라이브러리)"]
        HTTPClient["HTTP Client\n(재시도/백오프)"]
        ComputeSDK["ComputeClient"]
        NetworkSDK["NetworkClient"]
        StorageSDK["StorageClient"]
        IdentitySDK["IdentityClient"]
        ImageSDK["ImageClient"]
    end

    subgraph Gateway["kcp-gateway (Go + Gin)"]
        AuthMW["인증 미들웨어\n(JWT/Session/OAuth2)"]
        AuditMW["감사 로그 미들웨어"]
        ErrorMW["에러 핸들링 미들웨어"]
        Handlers["API 핸들러"]
        BizLogic["비즈니스 로직\n(다중 API 조합)"]
    end

    subgraph Data["데이터 레이어"]
        PostgreSQL[(PostgreSQL\nnerdctl 컨테이너)]
    end

    subgraph OpenStack["OpenStack Epoxy 2025.1"]
        Nova["Nova\n(Compute)"]
        Neutron["Neutron\n(Network)"]
        Cinder["Cinder\n(Storage)"]
        Keystone["Keystone\n(Identity)"]
        Glance["Glance\n(Image)"]
    end

    CLI --> SDK
    TUI --> SDK
    SDK --> Gateway
    WebUI -->|HTTPS 선택| Gateway

    Gateway --> AuthMW --> AuditMW --> ErrorMW --> Handlers
    Handlers --> BizLogic
    BizLogic --> SDK
    SDK --> HTTPClient

    HTTPClient -->|HTTPS 선택| Nova
    HTTPClient -->|HTTPS 선택| Neutron
    HTTPClient -->|HTTPS 선택| Cinder
    HTTPClient -->|HTTPS 선택| Keystone
    HTTPClient -->|HTTPS 선택| Glance

    Gateway -->|사용자/세션/감사 로그| PostgreSQL
```

**핵심 설계 원칙:**
- **kcp-sdk가 모든 OpenStack 통신의 단일 진입점** — CLI, TUI, Gateway 모두 동일 SDK 사용
- Gateway는 기본적으로 프록시이며, 필요 시 다중 API 조합 비즈니스 로직 수행
- HTTPS는 모든 구간에서 선택적 활성화 (설정 기반 on/off)
- 동시 접속 최대 20명 규모

---

## 3. 데이터베이스 모델링

### 3.1 ERD

```mermaid
erDiagram
    User ||--o{ Session : "보유"
    User ||--o{ AuditLog : "행위 기록"

    User {
        uuid id PK
        string username "로그인 ID"
        string password_hash "bcrypt 해시"
        string email
        string role "ADMIN"
        boolean is_active
        timestamp created_at
        timestamp updated_at
    }

    Session {
        uuid id PK
        uuid user_id FK
        string token "JWT 또는 세션 토큰"
        string auth_type "JWT | SESSION | OAUTH2"
        inet ip_address "접속 IP"
        timestamp expires_at
        timestamp created_at
    }

    AuditLog {
        uuid id PK
        uuid user_id FK
        string action "CREATE | READ | UPDATE | DELETE | LOGIN | LOGOUT"
        string resource_type "VM | NETWORK | VOLUME | PROJECT | USER | IMAGE"
        string resource_id "OpenStack 리소스 ID"
        string source "CLI | TUI | WEBUI"
        int status_code "HTTP 응답 코드"
        text request_summary "요청 요약 (민감 정보 마스킹)"
        inet ip_address
        timestamp created_at
    }
```

### 3.2 테이블 상세 명세

#### User 테이블

| 컬럼명 | 타입 | 제약조건 | 설명 |
|---|---|---|---|
| id | UUID | PK, DEFAULT gen_random_uuid() | 사용자 고유 ID |
| username | VARCHAR(100) | UNIQUE, NOT NULL | 로그인 ID |
| password_hash | VARCHAR(255) | NOT NULL | bcrypt 해시 (비용 인자 12 이상) |
| email | VARCHAR(255) | UNIQUE | 이메일 |
| role | VARCHAR(20) | NOT NULL, DEFAULT 'ADMIN' | 역할 |
| is_active | BOOLEAN | DEFAULT true | 활성 상태 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 생성 시각 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 수정 시각 |

**인덱스:** `idx_user_username` — username (로그인 조회)

#### Session 테이블

| 컬럼명 | 타입 | 제약조건 | 설명 |
|---|---|---|---|
| id | UUID | PK | 세션 고유 ID |
| user_id | UUID | FK(User.id), NOT NULL | 소유 사용자 |
| token | TEXT | NOT NULL | JWT 또는 세션 토큰 |
| auth_type | VARCHAR(20) | NOT NULL | JWT, SESSION, OAUTH2 |
| ip_address | INET | NOT NULL | 접속 IP |
| expires_at | TIMESTAMPTZ | NOT NULL | 만료 시각 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 생성 시각 |

**인덱스:**
- `idx_session_user` — user_id (사용자별 세션 조회)
- `idx_session_expires` — expires_at (만료 세션 정리)
- `idx_session_token` — token (토큰 검증)

#### AuditLog 테이블

| 컬럼명 | 타입 | 제약조건 | 설명 |
|---|---|---|---|
| id | UUID | PK | 로그 고유 ID |
| user_id | UUID | FK(User.id), NOT NULL | 행위자 |
| action | VARCHAR(20) | NOT NULL | CREATE, READ, UPDATE, DELETE, LOGIN, LOGOUT |
| resource_type | VARCHAR(30) | NOT NULL | VM, NETWORK, VOLUME, PROJECT, USER, IMAGE |
| resource_id | VARCHAR(255) | | OpenStack 리소스 ID |
| source | VARCHAR(10) | NOT NULL | CLI, TUI, WEBUI |
| status_code | INT | NOT NULL | HTTP 응답 코드 |
| request_summary | TEXT | | 요청 요약 (민감 정보 마스킹) |
| ip_address | INET | NOT NULL | 접속 IP |
| created_at | TIMESTAMPTZ | DEFAULT now() | 기록 시각 |

**인덱스:**
- `idx_audit_user_date` — user_id, created_at DESC (사용자별 이력)
- `idx_audit_resource` — resource_type, resource_id (리소스별 이력)
- `idx_audit_created` — created_at (1년 보관 정책 삭제 배치)

**파티셔닝:** created_at 기준 월별 파티셔닝 (1년 보관 후 파티션 단위 삭제)

---

## 4. 핵심 API 인터페이스 명세

모든 API는 `/api/v1` 접두사. 인증이 필요한 API는 `Authorization: Bearer <token>` 또는 세션 쿠키 필수.

### 4.1 인증 API

#### POST /api/v1/auth/login
**설명**: 로그인 및 토큰 발급
**인증**: 불필요

**Request Body**:
```json
{
  "username": "string (required)",
  "password": "string (required)",
  "authType": "string (enum: JWT, SESSION, OAUTH2, default: JWT)"
}
```

**Response 200**:
```json
{
  "token": "string",
  "authType": "JWT",
  "expiresAt": "ISO 8601",
  "user": {
    "id": "uuid",
    "username": "string",
    "role": "ADMIN"
  }
}
```

**에러 응답**:

| 코드 | type | 설명 |
|---|---|---|
| 401 | AUTH_FAILED | 자격 증명 불일치 |
| 429 | RATE_LIMITED | 로그인 시도 초과 |

#### POST /api/v1/auth/logout
**설명**: 로그아웃 및 토큰 무효화
**인증**: 필수

**Response 204**: No Content

#### POST /api/v1/auth/refresh
**설명**: 토큰 갱신
**인증**: 만료된 토큰 허용

**Response 200**:
```json
{
  "token": "string",
  "expiresAt": "ISO 8601"
}
```

### 4.2 Compute API

#### GET /api/v1/compute/servers
**설명**: VM 인스턴스 목록 조회
**인증**: 필수

**Query Params**:

| 파라미터 | 타입 | 필수 | 설명 |
|---|---|---|---|
| project_id | string | N | OpenStack 프로젝트 ID 필터 |
| status | string | N | ACTIVE, SHUTOFF, ERROR 등 |
| name | string | N | 이름 검색 (부분 일치) |
| page | int | N (기본 1) | 페이지 번호 |
| size | int | N (기본 20, 최대 100) | 페이지 크기 |

**Response 200**:
```json
{
  "servers": [
    {
      "id": "string",
      "name": "string",
      "status": "ACTIVE",
      "flavor": { "id": "string", "name": "string", "vcpus": 2, "ram": 4096, "disk": 40 },
      "addresses": {},
      "created": "ISO 8601",
      "updated": "ISO 8601"
    }
  ],
  "pagination": { "page": 1, "size": 20, "total": 45 }
}
```

#### POST /api/v1/compute/servers
**설명**: VM 인스턴스 생성
**인증**: 필수

**Request Body**:
```json
{
  "name": "string (required)",
  "flavorId": "string (required)",
  "imageId": "string (required)",
  "networkIds": ["string"],
  "securityGroupIds": ["string"],
  "keyName": "string",
  "userData": "string (base64)"
}
```

**Response 201**:
```json
{
  "id": "string",
  "name": "string",
  "status": "BUILD"
}
```

#### POST /api/v1/compute/servers/{serverId}/action
**설명**: VM 액션 수행 (시작/중지/재부팅)
**인증**: 필수

**Request Body**:
```json
{
  "action": "string (enum: start, stop, reboot, hard-reboot)"
}
```

**Response 202**: Accepted

#### DELETE /api/v1/compute/servers/{serverId}
**설명**: VM 삭제
**인증**: 필수
**Response 204**: No Content

#### GET /api/v1/compute/flavors
**설명**: Flavor 목록 조회
**인증**: 필수
**Response 200**: Flavor 배열

### 4.3 Network API

#### GET /api/v1/network/networks
**설명**: 네트워크 목록 조회
**인증**: 필수
**Response 200**: Network 배열 + pagination

#### POST /api/v1/network/networks
**설명**: 네트워크 생성
**인증**: 필수

**Request Body**:
```json
{
  "name": "string (required)",
  "adminStateUp": "boolean (default: true)",
  "shared": "boolean (default: false)"
}
```

#### GET /api/v1/network/subnets
**설명**: 서브넷 목록 조회
**Response 200**: Subnet 배열

#### POST /api/v1/network/subnets
**설명**: 서브넷 생성

**Request Body**:
```json
{
  "name": "string (required)",
  "networkId": "string (required)",
  "cidr": "string (required, CIDR 형식)",
  "ipVersion": "int (4 or 6)",
  "gatewayIp": "string",
  "enableDhcp": "boolean (default: true)"
}
```

#### GET/POST/DELETE /api/v1/network/routers
**설명**: 라우터 CRUD

#### POST /api/v1/network/routers/{routerId}/add-interface
**설명**: 라우터에 서브넷 인터페이스 연결

**Request Body**:
```json
{
  "subnetId": "string (required)"
}
```

#### GET/POST/DELETE /api/v1/network/security-groups
**설명**: 보안그룹 CRUD

#### POST /api/v1/network/security-groups/{sgId}/rules
**설명**: 보안그룹 규칙 추가

**Request Body**:
```json
{
  "direction": "string (ingress | egress)",
  "protocol": "string (tcp | udp | icmp)",
  "portRangeMin": "int",
  "portRangeMax": "int",
  "remoteIpPrefix": "string (CIDR)"
}
```

### 4.4 Storage API

#### GET /api/v1/storage/volumes
**설명**: 볼륨 목록 조회
**Response 200**: Volume 배열 + pagination

#### POST /api/v1/storage/volumes
**설명**: 볼륨 생성

**Request Body**:
```json
{
  "name": "string (required)",
  "size": "int (required, GB)",
  "volumeType": "string",
  "description": "string"
}
```

#### POST /api/v1/storage/volumes/{volumeId}/attach
**설명**: 볼륨을 VM에 연결

**Request Body**:
```json
{
  "serverId": "string (required)"
}
```

#### POST /api/v1/storage/volumes/{volumeId}/detach
**설명**: 볼륨을 VM에서 분리
**Response 202**: Accepted

#### GET/POST/DELETE /api/v1/storage/snapshots
**설명**: 스냅샷 CRUD

### 4.5 Identity API

#### GET/POST/DELETE /api/v1/identity/projects
**설명**: 프로젝트 CRUD

#### GET/POST/DELETE /api/v1/identity/users
**설명**: 사용자 CRUD

#### POST /api/v1/identity/roles/assign
**설명**: 역할 할당

**Request Body**:
```json
{
  "userId": "string (required)",
  "projectId": "string (required)",
  "roleId": "string (required)"
}
```

#### DELETE /api/v1/identity/roles/revoke
**설명**: 역할 해제

### 4.6 Image API

#### GET /api/v1/image/images
**설명**: 이미지 목록 조회
**Response 200**: Image 배열 + pagination

#### POST /api/v1/image/images
**설명**: 이미지 업로드
**Content-Type**: multipart/form-data

#### DELETE /api/v1/image/images/{imageId}
**설명**: 이미지 삭제
**Response 204**: No Content

### 4.7 감사 로그 API

#### GET /api/v1/audit/logs
**설명**: 감사 로그 조회
**인증**: 필수

**Query Params**:

| 파라미터 | 타입 | 필수 | 설명 |
|---|---|---|---|
| userId | UUID | N | 사용자 필터 |
| action | string | N | 작업 유형 필터 |
| resourceType | string | N | 리소스 유형 필터 |
| from | ISO 8601 | N | 시작 일시 |
| to | ISO 8601 | N | 종료 일시 |
| page | int | N | 페이지 번호 |
| size | int | N | 페이지 크기 |

**Response 200**:
```json
{
  "logs": [
    {
      "id": "uuid",
      "user": { "id": "uuid", "username": "string" },
      "action": "CREATE",
      "resourceType": "VM",
      "resourceId": "string",
      "source": "CLI",
      "statusCode": 201,
      "requestSummary": "VM 생성: web-server-01",
      "ipAddress": "192.168.x.x",
      "createdAt": "ISO 8601"
    }
  ],
  "pagination": { "page": 1, "size": 20, "total": 150 }
}
```

### 4.8 통계 API

#### GET /api/v1/stats/dashboard
**설명**: 대시보드 리소스 현황 요약
**인증**: 필수

**Response 200**:
```json
{
  "compute": { "total": 25, "active": 20, "shutoff": 3, "error": 2 },
  "network": { "networks": 5, "subnets": 12, "routers": 3 },
  "storage": { "volumes": 30, "totalSizeGB": 1500, "snapshots": 10 },
  "identity": { "projects": 4, "users": 15 },
  "image": { "total": 8 }
}
```

---

## 5. 폴더 구조 및 컴포넌트 분리 전략

```
kcp/
├── kcp-cli/                         # CLI, SDK, TUI (Go)
│   ├── cmd/
│   │   ├── cli/                     # CLI 진입점
│   │   │   └── main.go
│   │   └── tui/                     # TUI 진입점
│   │       └── main.go
│   ├── pkg/
│   │   └── sdk/                     # 공유 SDK 라이브러리
│   │       ├── client.go            # 공통 HTTP 클라이언트 (재시도/백오프)
│   │       ├── types.go             # OpenStack 리소스 타입 정의
│   │       ├── compute.go           # ComputeClient 구현
│   │       ├── network.go           # NetworkClient 구현
│   │       ├── storage.go           # StorageClient 구현
│   │       ├── identity.go          # IdentityClient 구현
│   │       └── image.go             # ImageClient 구현
│   ├── internal/
│   │   ├── cli/                     # CLI 명령어 (cobra)
│   │   │   ├── root.go             # 루트 커맨드, 글로벌 플래그
│   │   │   ├── auth.go             # login/logout
│   │   │   ├── compute.go          # vm, flavor 명령어
│   │   │   ├── network.go          # network, subnet, router, secgroup
│   │   │   ├── storage.go          # volume, snapshot
│   │   │   ├── identity.go         # project, user, role
│   │   │   ├── image.go            # image
│   │   │   ├── audit.go            # audit
│   │   │   └── output.go           # 출력 포맷터 (table/json/yaml)
│   │   ├── tui/                     # TUI 화면 (bubbletea)
│   │   │   ├── app.go              # 메인 모델, 네비게이션
│   │   │   ├── views/              # 리소스별 뷰
│   │   │   └── styles/             # 스타일 정의
│   │   └── config/                  # 설정 파일 관리
│   │       └── config.go           # ~/.kcp/config 읽기/쓰기, 우선순위
│   ├── go.mod
│   ├── go.sum
│   └── Makefile
│
├── kcp-gateway/                     # Backend API Gateway (Go + Gin)
│   ├── cmd/
│   │   └── server/
│   │       └── main.go             # 서버 진입점
│   ├── internal/
│   │   ├── handler/                 # API 핸들러 (라우트별)
│   │   │   ├── auth.go
│   │   │   ├── compute.go
│   │   │   ├── network.go
│   │   │   ├── storage.go
│   │   │   ├── identity.go
│   │   │   ├── image.go
│   │   │   ├── audit.go
│   │   │   └── stats.go
│   │   ├── service/                 # 비즈니스 로직
│   │   │   ├── auth_service.go      # 인증 로직 (JWT/Session/OAuth2)
│   │   │   ├── audit_service.go     # 감사 로그 CRUD
│   │   │   └── stats_service.go     # 통계 집계
│   │   ├── middleware/              # Gin 미들웨어
│   │   │   ├── auth.go             # 인증 검증 (JWT/Session/OAuth2 분기)
│   │   │   ├── audit.go            # 감사 로그 자동 기록
│   │   │   ├── error.go            # 에러 핸들링
│   │   │   ├── cors.go             # CORS 설정
│   │   │   └── ratelimit.go        # Rate Limiting
│   │   ├── model/                   # DB 모델
│   │   │   ├── user.go
│   │   │   ├── session.go
│   │   │   └── audit_log.go
│   │   ├── repository/             # DB 접근 레이어
│   │   │   ├── user_repo.go
│   │   │   ├── session_repo.go
│   │   │   └── audit_repo.go
│   │   └── database/               # DB 연결 및 마이그레이션
│   │       ├── postgres.go         # 연결 풀, 헬스체크
│   │       └── migrations/         # SQL 마이그레이션 파일
│   ├── config/
│   │   └── config.go               # 서버 설정 (OpenStack 인증 정보 포함)
│   ├── .env.example
│   ├── go.mod
│   ├── go.sum
│   └── Makefile
│
├── kcp-webui/                       # 웹 관리 콘솔 (Next.js)
│   ├── src/
│   │   ├── app/                     # App Router 페이지
│   │   │   ├── layout.tsx           # 루트 레이아웃 (다크 모드)
│   │   │   ├── page.tsx             # 리다이렉트
│   │   │   ├── login/
│   │   │   │   └── page.tsx         # 로그인
│   │   │   ├── dashboard/
│   │   │   │   └── page.tsx         # 대시보드
│   │   │   ├── compute/
│   │   │   │   ├── page.tsx         # VM 목록
│   │   │   │   └── [id]/page.tsx    # VM 상세
│   │   │   ├── network/
│   │   │   │   └── page.tsx         # 네트워크 관리
│   │   │   ├── storage/
│   │   │   │   └── page.tsx         # 볼륨/스냅샷
│   │   │   ├── identity/
│   │   │   │   └── page.tsx         # 프로젝트/사용자
│   │   │   ├── image/
│   │   │   │   └── page.tsx         # 이미지 관리
│   │   │   └── audit/
│   │   │       └── page.tsx         # 감사 로그
│   │   ├── components/              # UI 컴포넌트
│   │   │   ├── ui/                  # shadcn/ui 원자 컴포넌트
│   │   │   ├── layout/
│   │   │   │   ├── Header.tsx
│   │   │   │   ├── Sidebar.tsx
│   │   │   │   └── AuthGuard.tsx
│   │   │   ├── resource/           # 리소스 관리 공통 컴포넌트
│   │   │   │   ├── DataTable.tsx    # 공통 테이블 (검색/필터/페이지네이션)
│   │   │   │   ├── ResourceForm.tsx # 공통 생성/수정 폼
│   │   │   │   └── ConfirmDialog.tsx
│   │   │   └── dashboard/
│   │   │       ├── StatCard.tsx
│   │   │       └── StatusChart.tsx
│   │   ├── hooks/
│   │   │   ├── useAuth.ts
│   │   │   ├── useResource.ts       # 범용 리소스 CRUD 훅
│   │   │   └── useDashboard.ts
│   │   ├── services/                # API 통신
│   │   │   ├── api.ts               # axios 인스턴스 + 인터셉터
│   │   │   ├── computeService.ts
│   │   │   ├── networkService.ts
│   │   │   ├── storageService.ts
│   │   │   ├── identityService.ts
│   │   │   ├── imageService.ts
│   │   │   └── auditService.ts
│   │   ├── stores/                  # Zustand
│   │   │   ├── authStore.ts
│   │   │   └── sidebarStore.ts
│   │   ├── types/                   # TypeScript 타입
│   │   │   ├── compute.ts
│   │   │   ├── network.ts
│   │   │   ├── storage.ts
│   │   │   ├── identity.ts
│   │   │   ├── image.ts
│   │   │   ├── audit.ts
│   │   │   └── api.ts
│   │   └── utils/
│   │       └── formatters.ts
│   ├── .env.example
│   ├── next.config.js
│   ├── tailwind.config.ts
│   └── package.json
│
├── db/                              # DB 관련
│   ├── init.sql                     # 초기 스키마
│   └── nerdctl-postgres.sh          # nerdctl PostgreSQL 실행 스크립트
├── .gitignore
├── .editorconfig
└── Makefile                         # 루트 Makefile (전체 빌드/실행)
```

**컴포넌트 분리 기준:**
- **kcp-sdk**: 외부 공개 패키지 (`pkg/`), CLI/TUI/Gateway 모두에서 import 가능
- **kcp-gateway**: Handler → Service → Repository 3단 레이어
- **kcp-webui**: 리소스별 공통 패턴 (DataTable, ResourceForm)으로 중복 최소화

---

## 6. 상태 관리 및 데이터 흐름도

### 6.1 CLI/TUI 데이터 흐름

```mermaid
flowchart LR
    Config["~/.kcp/config\n(토큰, 서버 URL)"]
    CLI["CLI 명령어\n(cobra)"]
    SDK["kcp-sdk"]
    GW["Gateway API"]
    OS["OpenStack"]
    Output["출력\n(table/json/yaml)"]

    Config -->|읽기| CLI
    CLI -->|요청 구성| SDK
    SDK -->|HTTP| GW
    GW -->|프록시/조합| OS
    OS -->|응답| GW
    GW -->|응답| SDK
    SDK -->|파싱| CLI
    CLI -->|포맷팅| Output
```

### 6.2 WebUI 상태 관리

| 상태 유형 | 관리 방식 | 예시 |
|---|---|---|
| 서버 상태 | React Query (커스텀 훅) | VM 목록, 네트워크 목록 |
| 전역 클라이언트 상태 | Zustand | 인증 정보, 사이드바 상태 |
| 로컬 UI 상태 | React useState | 필터값, 모달 열림/닫힘, 폼 입력 |

### 6.3 인증 상태 흐름

```mermaid
stateDiagram-v2
    [*] --> 미인증

    미인증 --> 로그인중: 로그인 요청
    로그인중 --> 인증완료: 토큰 수신
    로그인중 --> 미인증: 인증 실패

    인증완료 --> API요청: 리소스 접근
    API요청 --> 인증완료: 정상 응답
    API요청 --> 토큰갱신중: 401 응답
    토큰갱신중 --> 인증완료: 갱신 성공
    토큰갱신중 --> 미인증: 갱신 실패 (재로그인 요구)

    인증완료 --> 미인증: 로그아웃
```

---

## 7. 보안 설계

### 7.1 OWASP Top 10 대응

| 위협 | 대응 방안 | 구현 위치 |
|---|---|---|
| A01: Broken Access Control | 인증 미들웨어 필수 적용, API별 권한 검증 | `middleware/auth.go` |
| A02: Cryptographic Failures | bcrypt 비밀번호 해싱, HTTPS 선택 지원, 토큰 파일 권한 600 | `service/auth_service.go`, `config/config.go` |
| A03: Injection | Go의 database/sql Parameterized Query 강제, 입력 검증 | `repository/*.go`, `handler/*.go` |
| A04: Insecure Design | SDK 단일 진입점, Gateway 프록시 패턴 | 아키텍처 전체 |
| A05: Security Misconfiguration | CORS 화이트리스트, 불필요한 헤더 제거, 디버그 모드 환경 분리 | `middleware/cors.go` |
| A06: Vulnerable Components | Go mod tidy, npm audit, MIT/Apache 2.0 라이선스만 허용 | CI |
| A07: Auth Failures | JWT 만료 15분, Refresh Token, Rate Limiting (로그인 5회/15분) | `middleware/auth.go`, `middleware/ratelimit.go` |
| A08: Data Integrity | 감사 로그 불변 (UPDATE/DELETE 불가), request_summary 기록 | `audit_repo.go` |
| A09: Logging Failures | 감사 로그 자동 기록, 민감 정보 마스킹 (비밀번호, 토큰) | `middleware/audit.go` |
| A10: SSRF | Gateway가 허용된 OpenStack 엔드포인트만 호출, URL 화이트리스트 | `config/config.go` |

### 7.2 인증/인가 흐름

```mermaid
sequenceDiagram
    participant C as CLI/WebUI
    participant GW as Gateway
    participant DB as PostgreSQL
    participant OS as OpenStack

    C->>GW: POST /api/v1/auth/login (username, password)
    GW->>DB: 사용자 조회 + bcrypt 검증
    DB->>GW: 사용자 정보
    GW->>GW: JWT 토큰 생성 (15분 만료)
    GW->>DB: 세션 저장
    GW->>DB: 감사 로그 기록 (LOGIN)
    GW->>C: 200 { token, expiresAt }

    Note over C,OS: --- 이후 API 요청 ---

    C->>GW: GET /api/v1/compute/servers + Bearer Token
    GW->>GW: JWT 검증 (middleware/auth.go)
    GW->>GW: 감사 로그 기록 (middleware/audit.go)
    GW->>OS: OpenStack Nova API 호출 (kcp-sdk)
    OS->>GW: 서버 목록
    GW->>C: 200 { servers }
```

### 7.3 민감 데이터 처리 정책

| 데이터 | 보호 방식 | 비고 |
|---|---|---|
| 사용자 비밀번호 | bcrypt 해싱 (비용 인자 12) | 평문 저장 절대 금지 |
| JWT 토큰 | HMAC-SHA256 서명, 15분 만료 | 서명 키는 환경변수 |
| CLI 설정 파일 | 파일 권한 600 (소유자만 읽기/쓰기) | `~/.kcp/config` |
| OpenStack 인증 정보 | Gateway 서버 설정 파일 (환경변수) | 소스코드 하드코딩 금지 |
| 감사 로그 request_summary | 비밀번호, 토큰 마스킹 | `p***`, `eyJ***` |
| 에러 로그 | 스택 트레이스에 민감 정보 미포함 | 프로덕션 환경 |

### 7.4 환경변수 관리

```bash
# kcp-gateway/.env.example (절대 커밋 금지)
DATABASE_URL=postgresql://user:pass@localhost:5432/kcp
JWT_SECRET=your-secret-key
JWT_EXPIRY=15m
OPENSTACK_AUTH_URL=https://keystone.example.com:5000/v3
OPENSTACK_USERNAME=admin
OPENSTACK_PASSWORD=secret
OPENSTACK_PROJECT_ID=project-uuid
OPENSTACK_DOMAIN_ID=default
TLS_ENABLED=false
TLS_CERT_PATH=
TLS_KEY_PATH=
```

### 7.5 의존성 라이선스 검토

| 패키지 | 라이선스 | 비고 |
|---|---|---|
| gin-gonic/gin | MIT | 허용 |
| spf13/cobra | Apache 2.0 | 허용 |
| charmbracelet/bubbletea | MIT | 허용 |
| lib/pq | MIT | 허용 |
| golang-jwt/jwt | MIT | 허용 |
| React / Next.js | MIT | 허용 |
| Tailwind CSS | MIT | 허용 |
| shadcn/ui | MIT | 허용 |
| Zustand | MIT | 허용 |
| axios | MIT | 허용 |

> GPL류 라이선스 패키지는 사용하지 않음.

---

## 8. 예외 처리 및 에러 전략

### 8.1 에러 분류 체계

| 분류 | HTTP 코드 | 예시 | 재시도 |
|---|---|---|---|
| 입력 검증 에러 | 400 | 필수 필드 누락, 잘못된 CIDR 형식 | 불가 |
| 인증 에러 | 401 | JWT 만료, 토큰 없음 | 토큰 갱신 후 재시도 |
| 인가 에러 | 403 | 권한 부족 | 불가 |
| 리소스 미존재 | 404 | 존재하지 않는 VM ID | 불가 |
| 비즈니스 규칙 위반 | 409/422 | 이미 연결된 볼륨, 중복 이름 | 불가 |
| Rate Limit 초과 | 429 | 로그인 시도 초과 | Retry-After 헤더 확인 후 재시도 |
| OpenStack 타임아웃 | 504 | OpenStack API 무응답 | 자동 재시도 (SDK 3회) |
| OpenStack 인증 실패 | 502 | Keystone 인증 실패 | 불가 — 설정 확인 필요 |
| OpenStack 서비스 다운 | 503 | Nova/Neutron 등 서비스 다운 | 자동 재시도 (SDK 3회) |
| 내부 서버 에러 | 500 | DB 연결 실패, 예기치 못한 에러 | 자동 재시도 (최대 3회) |

### 8.2 에러 응답 표준 포맷

**Gateway API 에러:**
```json
{
  "error": {
    "code": "OPENSTACK_TIMEOUT",
    "message": "OpenStack Compute 서비스가 응답하지 않습니다.",
    "status": 504,
    "detail": "Nova API 호출 30초 타임아웃",
    "traceId": "abc-123-def"
  }
}
```

**CLI 에러 출력:**
```
Error: OpenStack Compute 서비스가 응답하지 않습니다. (504)
Detail: Nova API 호출 30초 타임아웃
Trace: abc-123-def

재시도하려면: kcp vm list --retry
```

### 8.3 로깅 전략

| 레벨 | 용도 | 마스킹 대상 |
|---|---|---|
| ERROR | 서버 에러, OpenStack 장애 | 비밀번호, 토큰 |
| WARN | Rate Limit 도달, 재시도 발생 | — |
| INFO | API 요청/응답 (상태코드, 소요시간) | Body 내용 제외 |
| DEBUG | 상세 실행 흐름 (개발 환경만) | — |

### 8.4 재시도 정책

| 구간 | 재시도 | 전략 |
|---|---|---|
| kcp-sdk → OpenStack API | 3회 | 지수 백오프 (1s, 2s, 4s) |
| CLI → Gateway | 수동 (`--retry` 플래그) | 사용자 제어 |
| WebUI → Gateway | 자동 (401만) | 토큰 갱신 후 1회 재시도 |

### 8.5 멱등성 보장

| API | 멱등성 | 전략 |
|---|---|---|
| GET 전체 | O | 읽기 전용 |
| POST /servers (VM 생성) | X | 클라이언트 중복 방지 (debounce) |
| POST /servers/{id}/action | X | 상태 확인 후 액션 수행 |
| DELETE 전체 | O | 이미 삭제된 경우 204 반환 |

---

## 9. 3줄 요약 및 비유

> **3줄 요약**
> 1. 관리자는 CLI, TUI, 웹콘솔 중 어떤 도구를 사용하든 동일한 Gateway API를 통해 OpenStack 인프라를 관리합니다.
> 2. kcp-sdk가 모든 OpenStack 통신을 담당하여, 재시도/에러 처리/HTTPS 등의 로직이 한 곳에서 관리됩니다.
> 3. 모든 관리 작업은 감사 로그에 자동 기록되며, 1년간 보관됩니다.
>
> **비유로 이해하기**
> KCP CLI는 **리모컨이 세 개인 TV**와 비슷합니다.
> 일반 리모컨(CLI), 음성 리모컨(TUI), 스마트폰 앱(웹콘솔) 중 어떤 걸 써도
> 모두 같은 수신기(Gateway)를 거쳐 TV(OpenStack)를 제어합니다.
> 수신기 안에는 만능 번역기(kcp-sdk)가 있어서 어떤 리모컨이든 같은 방식으로 TV와 통신하고,
> 누가 언제 어떤 채널을 돌렸는지 자동으로 기록부(감사 로그)에 남깁니다.
