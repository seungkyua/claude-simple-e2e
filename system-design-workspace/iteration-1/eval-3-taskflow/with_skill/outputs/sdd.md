# 시스템 설계서: TaskFlow

## 1. 프로젝트 개요

| 항목 | 내용 |
|---|---|
| 프로젝트명 | TaskFlow |
| 한 줄 설명 | 소규모 팀(2~10명)을 위한 실시간 협업 칸반보드 기반 할일 관리 웹앱 |
| 기술 스택 | React 18 + Next.js 14, Node.js + Express, PostgreSQL, Redis, Socket.IO, Tailwind CSS |
| 작성일 | 2026-03-26 |
| 기반 문서 | prd.md |

### 기술 스택 선정 근거

| 영역 | 기술 | 선정 근거 | 라이선스 |
|---|---|---|---|
| 프론트엔드 | React 18 + Next.js 14 | 드래그앤드롭, 실시간 UI 업데이트에 최적화된 생태계. App Router의 서버 컴포넌트로 초기 로딩 성능 확보 | MIT |
| 상태 관리 | Zustand + TanStack Query v5 | 경량 전역 상태(Zustand) + 서버 상태 캐싱/동기화(TanStack Query) 분리 | MIT |
| 드래그앤드롭 | @dnd-kit/core | 접근성(a11y) 기본 지원, 경량, 커스텀 용이 | MIT |
| 스타일링 | Tailwind CSS v3 | 유틸리티 기반 빠른 반응형 UI 개발 | MIT |
| 백엔드 | Node.js + Express | WebSocket과 동일 런타임으로 실시간 처리 단순화 | MIT |
| 실시간 통신 | Socket.IO v4 | WebSocket 기반 실시간 양방향 통신, 자동 재연결, 룸 기반 이벤트 분리 | MIT |
| ORM | Prisma v5 | 타입 안전 쿼리, 마이그레이션 관리 용이 | Apache 2.0 |
| 데이터베이스 | PostgreSQL 16 | 관계형 데이터(팀-멤버-보드-태스크) 무결성 보장, JSON 컬럼 지원 | PostgreSQL License (MIT 유사) |
| 캐시/Pub-Sub | Redis 7 | 실시간 Pub/Sub, 세션 캐시, 알림 큐 | BSD-3-Clause |
| 인증 | JWT (jose 라이브러리) | 무상태 토큰 인증, Refresh Token 회전 | MIT |
| 유효성 검증 | zod | 런타임 스키마 검증, TypeScript 타입 추론 | MIT |
| 인프라 | Docker + Docker Compose | 사내 도구이므로 셀프 호스팅 최적화, 무료 운영 | Apache 2.0 |

---

## 2. 전체 시스템 아키텍처

```mermaid
flowchart TB
    subgraph Client["클라이언트 (브라우저)"]
        NextApp["Next.js App<br/>(React 18)"]
        DnD["@dnd-kit<br/>드래그앤드롭"]
        SocketClient["Socket.IO Client<br/>실시간 수신"]
    end

    subgraph Server["백엔드 서버"]
        Express["Express API 서버<br/>(REST API)"]
        SocketServer["Socket.IO Server<br/>실시간 브로드캐스트"]
        AuthMiddleware["인증 미들웨어<br/>(JWT 검증)"]
        NotificationService["알림 서비스<br/>(마감일/할당 알림)"]
    end

    subgraph DataLayer["데이터 레이어"]
        PG[(PostgreSQL<br/>영속 데이터)]
        RedisCache[(Redis<br/>캐시 + Pub/Sub)]
    end

    subgraph External["외부 서비스"]
        SMTP["SMTP 서버<br/>(초대/알림 이메일)"]
    end

    NextApp -- "REST API 요청" --> AuthMiddleware
    NextApp -- "드래그앤드롭 이벤트" --> DnD
    DnD -- "상태 변경 API" --> AuthMiddleware
    AuthMiddleware --> Express
    SocketClient <-- "실시간 이벤트" --> SocketServer
    Express --> PG
    Express --> RedisCache
    SocketServer --> RedisCache
    NotificationService --> RedisCache
    NotificationService --> SMTP
    RedisCache -- "Pub/Sub 이벤트" --> SocketServer
```

### 아키텍처 설명

- **클라이언트**: Next.js App Router 기반 SPA. REST API로 CRUD 수행, Socket.IO로 실시간 변경사항 수신
- **API 서버**: Express 기반 RESTful API. 모든 요청은 JWT 인증 미들웨어를 통과
- **Socket.IO 서버**: 동일 Express 서버에 마운트. 팀(보드)별 Room으로 이벤트 격리
- **Redis**: Pub/Sub으로 다중 서버 인스턴스 간 실시간 이벤트 동기화, 세션/캐시 저장
- **PostgreSQL**: 핵심 비즈니스 데이터 영속 저장
- **알림 서비스**: Redis 큐 기반 비동기 처리 (마감일 체크는 cron job)

---

## 3. 데이터베이스 모델링

### 3.1 ERD

```mermaid
erDiagram
    USER ||--o{ TEAM_MEMBER : "belongs to"
    TEAM ||--o{ TEAM_MEMBER : "has"
    TEAM ||--o{ BOARD : "owns"
    BOARD ||--o{ BOARD_COLUMN : "contains"
    BOARD_COLUMN ||--o{ TASK : "contains"
    TASK ||--o{ TASK_LABEL : "tagged with"
    TASK ||--o{ COMMENT : "has"
    TASK }o--o| USER : "assigned to"
    LABEL ||--o{ TASK_LABEL : "applied to"
    USER ||--o{ COMMENT : "writes"
    USER ||--o{ NOTIFICATION : "receives"

    USER {
        uuid id PK
        varchar email UK
        varchar password_hash
        varchar display_name
        varchar avatar_url
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    TEAM {
        uuid id PK
        varchar name
        uuid created_by FK
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    TEAM_MEMBER {
        uuid id PK
        uuid team_id FK
        uuid user_id FK
        enum role
        timestamp joined_at
    }

    BOARD {
        uuid id PK
        uuid team_id FK
        varchar name
        timestamp created_at
        timestamp updated_at
    }

    BOARD_COLUMN {
        uuid id PK
        uuid board_id FK
        varchar name
        int position
        timestamp created_at
        timestamp updated_at
    }

    TASK {
        uuid id PK
        uuid column_id FK
        uuid assignee_id FK
        varchar title
        text description
        enum priority
        int position
        date due_date
        varchar image_url
        int version
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    LABEL {
        uuid id PK
        uuid board_id FK
        varchar name
        varchar color
    }

    TASK_LABEL {
        uuid task_id FK
        uuid label_id FK
    }

    COMMENT {
        uuid id PK
        uuid task_id FK
        uuid author_id FK
        text content
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    NOTIFICATION {
        uuid id PK
        uuid user_id FK
        enum type
        jsonb payload
        boolean is_read
        timestamp created_at
    }
```

### 3.2 테이블 상세 명세

#### users

| 컬럼명 | 타입 | 제약조건 | 설명 |
|---|---|---|---|
| id | UUID | PK, DEFAULT gen_random_uuid() | 사용자 고유 식별자 |
| email | VARCHAR(255) | UNIQUE, NOT NULL | 로그인 이메일 |
| password_hash | VARCHAR(255) | NOT NULL | 비밀번호 해시 (PBKDF2, 최소 10만 반복) |
| display_name | VARCHAR(100) | NOT NULL | 표시 이름 |
| avatar_url | VARCHAR(500) | NULLABLE | 프로필 이미지 URL |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | 생성일 |
| updated_at | TIMESTAMPTZ | DEFAULT NOW() | 수정일 |
| deleted_at | TIMESTAMPTZ | NULLABLE | 소프트 삭제 일시 |

- **인덱스**: `idx_users_email` (email), `idx_users_deleted_at` (deleted_at)
- **암호화**: `password_hash`는 PBKDF2(SHA-256, salt 32바이트, 반복 100,000회) 적용

#### teams

| 컬럼명 | 타입 | 제약조건 | 설명 |
|---|---|---|---|
| id | UUID | PK | 팀 고유 식별자 |
| name | VARCHAR(100) | NOT NULL | 팀 이름 |
| created_by | UUID | FK(users.id), NOT NULL | 팀 생성자 |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | 생성일 |
| updated_at | TIMESTAMPTZ | DEFAULT NOW() | 수정일 |
| deleted_at | TIMESTAMPTZ | NULLABLE | 소프트 삭제 일시 |

#### team_members

| 컬럼명 | 타입 | 제약조건 | 설명 |
|---|---|---|---|
| id | UUID | PK | 멤버십 식별자 |
| team_id | UUID | FK(teams.id), NOT NULL | 팀 참조 |
| user_id | UUID | FK(users.id), NOT NULL | 사용자 참조 |
| role | ENUM('admin','member') | NOT NULL, DEFAULT 'member' | 역할 |
| joined_at | TIMESTAMPTZ | DEFAULT NOW() | 가입일 |

- **인덱스**: `idx_team_members_team_user` UNIQUE (team_id, user_id)
- **제약**: 팀당 최대 10명 (애플리케이션 레벨 검증)

#### boards

| 컬럼명 | 타입 | 제약조건 | 설명 |
|---|---|---|---|
| id | UUID | PK | 보드 고유 식별자 |
| team_id | UUID | FK(teams.id), NOT NULL | 소속 팀 |
| name | VARCHAR(100) | NOT NULL | 보드 이름 |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | 생성일 |
| updated_at | TIMESTAMPTZ | DEFAULT NOW() | 수정일 |

- **인덱스**: `idx_boards_team_id` (team_id)

#### board_columns

| 컬럼명 | 타입 | 제약조건 | 설명 |
|---|---|---|---|
| id | UUID | PK | 컬럼 식별자 |
| board_id | UUID | FK(boards.id), NOT NULL | 소속 보드 |
| name | VARCHAR(50) | NOT NULL | 컬럼 이름 |
| position | INTEGER | NOT NULL | 정렬 순서 |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | 생성일 |
| updated_at | TIMESTAMPTZ | DEFAULT NOW() | 수정일 |

- **인덱스**: `idx_board_columns_board_position` (board_id, position)

#### tasks

| 컬럼명 | 타입 | 제약조건 | 설명 |
|---|---|---|---|
| id | UUID | PK | 태스크 식별자 |
| column_id | UUID | FK(board_columns.id), NOT NULL | 소속 컬럼 |
| assignee_id | UUID | FK(users.id), NULLABLE | 담당자 |
| title | VARCHAR(200) | NOT NULL | 제목 |
| description | TEXT | NULLABLE | 설명 |
| priority | ENUM('low','medium','high','urgent') | NOT NULL, DEFAULT 'medium' | 우선순위 |
| position | INTEGER | NOT NULL | 컬럼 내 정렬 순서 |
| due_date | DATE | NULLABLE | 마감일 |
| image_url | VARCHAR(500) | NULLABLE | 첨부 이미지 URL |
| version | INTEGER | NOT NULL, DEFAULT 1 | 낙관적 잠금용 버전 |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | 생성일 |
| updated_at | TIMESTAMPTZ | DEFAULT NOW() | 수정일 |
| deleted_at | TIMESTAMPTZ | NULLABLE | 소프트 삭제 일시 |

- **인덱스**: `idx_tasks_column_position` (column_id, position), `idx_tasks_assignee` (assignee_id), `idx_tasks_due_date` (due_date WHERE deleted_at IS NULL)
- **제약**: 보드당 최대 500개 (애플리케이션 레벨 검증)

#### labels

| 컬럼명 | 타입 | 제약조건 | 설명 |
|---|---|---|---|
| id | UUID | PK | 라벨 식별자 |
| board_id | UUID | FK(boards.id), NOT NULL | 소속 보드 |
| name | VARCHAR(30) | NOT NULL | 라벨 이름 |
| color | VARCHAR(7) | NOT NULL | HEX 컬러 코드 |

- **인덱스**: `idx_labels_board_id` (board_id)

#### task_labels

| 컬럼명 | 타입 | 제약조건 | 설명 |
|---|---|---|---|
| task_id | UUID | FK(tasks.id), NOT NULL | 태스크 참조 |
| label_id | UUID | FK(labels.id), NOT NULL | 라벨 참조 |

- **인덱스**: PK (task_id, label_id)

#### comments

| 컬럼명 | 타입 | 제약조건 | 설명 |
|---|---|---|---|
| id | UUID | PK | 댓글 식별자 |
| task_id | UUID | FK(tasks.id), NOT NULL | 소속 태스크 |
| author_id | UUID | FK(users.id), NOT NULL | 작성자 |
| content | TEXT | NOT NULL, 최대 2000자 | 댓글 내용 |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | 생성일 |
| updated_at | TIMESTAMPTZ | DEFAULT NOW() | 수정일 |
| deleted_at | TIMESTAMPTZ | NULLABLE | 소프트 삭제 일시 |

- **인덱스**: `idx_comments_task_id` (task_id)

#### notifications

| 컬럼명 | 타입 | 제약조건 | 설명 |
|---|---|---|---|
| id | UUID | PK | 알림 식별자 |
| user_id | UUID | FK(users.id), NOT NULL | 수신자 |
| type | ENUM('task_assigned','due_soon','due_today','comment') | NOT NULL | 알림 유형 |
| payload | JSONB | NOT NULL | 알림 상세 데이터 |
| is_read | BOOLEAN | NOT NULL, DEFAULT false | 읽음 여부 |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | 생성일 |

- **인덱스**: `idx_notifications_user_unread` (user_id, is_read) WHERE is_read = false

### 3.3 소프트 삭제 대상 테이블

| 테이블 | 사유 |
|---|---|
| users | 계정 복구 가능성, 기존 태스크/댓글 참조 유지 |
| teams | 팀 복구 가능성, 히스토리 보존 |
| tasks | 실수로 삭제한 태스크 복원, 히스토리 추적 |
| comments | 댓글 복원 가능성 |

---

## 4. 핵심 API 인터페이스 명세

### 4.1 인증 API

#### POST /api/v1/auth/register
**설명**: 회원가입
**인증**: 불필요
**Request Body**:
```json
{
  "email": "string (required, 이메일 형식, 최대 255자)",
  "password": "string (required, 최소 8자, 대소문자+숫자+특수문자 포함)",
  "displayName": "string (required, 1-100자)"
}
```
**Response 201**:
```json
{
  "id": "uuid",
  "email": "string",
  "displayName": "string",
  "createdAt": "ISO 8601"
}
```
**에러 응답**:

| 코드 | 타입 | 설명 |
|---|---|---|
| 400 | VALIDATION_ERROR | 입력값 검증 실패 |
| 409 | DUPLICATE_EMAIL | 이미 등록된 이메일 |

#### POST /api/v1/auth/login
**설명**: 로그인
**인증**: 불필요
**Request Body**:
```json
{
  "email": "string (required)",
  "password": "string (required)"
}
```
**Response 200**:
```json
{
  "accessToken": "string (JWT, 15분 만료)",
  "refreshToken": "string (7일 만료)"
}
```
**에러 응답**:

| 코드 | 타입 | 설명 |
|---|---|---|
| 401 | INVALID_CREDENTIALS | 이메일 또는 비밀번호 불일치 |
| 423 | ACCOUNT_LOCKED | 로그인 5회 실패로 계정 잠금 (15분) |

#### POST /api/v1/auth/refresh
**설명**: Access Token 갱신
**인증**: Refresh Token 필수
**Request Body**:
```json
{
  "refreshToken": "string (required)"
}
```
**Response 200**:
```json
{
  "accessToken": "string (JWT, 15분 만료)",
  "refreshToken": "string (새로 발급, 기존 토큰 무효화)"
}
```

---

### 4.2 팀 관리 API

#### POST /api/v1/teams
**설명**: 팀 생성
**인증**: Bearer Token 필수
**Request Body**:
```json
{
  "name": "string (required, 1-100자)"
}
```
**Response 201**:
```json
{
  "id": "uuid",
  "name": "string",
  "createdBy": "uuid",
  "createdAt": "ISO 8601"
}
```

#### POST /api/v1/teams/:teamId/members/invite
**설명**: 팀 멤버 초대 (이메일)
**인증**: Bearer Token 필수
**인가**: 팀 admin만 가능
**Request Body**:
```json
{
  "email": "string (required, 이메일 형식)"
}
```
**Response 200**:
```json
{
  "message": "초대 이메일이 발송되었습니다",
  "invitedEmail": "string"
}
```
**에러 응답**:

| 코드 | 타입 | 설명 |
|---|---|---|
| 403 | FORBIDDEN | admin 권한 없음 |
| 404 | USER_NOT_FOUND | 등록되지 않은 이메일 |
| 409 | ALREADY_MEMBER | 이미 팀 멤버 |
| 422 | TEAM_FULL | 팀 인원 10명 초과 |

#### GET /api/v1/teams/:teamId/members
**설명**: 팀 멤버 목록 조회
**인증**: Bearer Token 필수
**인가**: 팀 멤버만 가능
**Response 200**:
```json
{
  "members": [
    {
      "id": "uuid",
      "userId": "uuid",
      "displayName": "string",
      "email": "string",
      "role": "admin | member",
      "joinedAt": "ISO 8601",
      "isOnline": true
    }
  ]
}
```

---

### 4.3 보드 API

#### GET /api/v1/teams/:teamId/boards
**설명**: 팀 보드 목록 조회
**인증**: Bearer Token 필수
**인가**: 팀 멤버만 가능
**Response 200**:
```json
{
  "boards": [
    {
      "id": "uuid",
      "name": "string",
      "createdAt": "ISO 8601"
    }
  ]
}
```

#### POST /api/v1/teams/:teamId/boards
**설명**: 보드 생성 (기본 3개 컬럼 자동 생성: To Do, In Progress, Done)
**인증**: Bearer Token 필수
**인가**: 팀 멤버
**Request Body**:
```json
{
  "name": "string (required, 1-100자)"
}
```
**Response 201**:
```json
{
  "id": "uuid",
  "name": "string",
  "columns": [
    { "id": "uuid", "name": "To Do", "position": 0 },
    { "id": "uuid", "name": "In Progress", "position": 1 },
    { "id": "uuid", "name": "Done", "position": 2 }
  ],
  "createdAt": "ISO 8601"
}
```

---

### 4.4 컬럼 API

#### POST /api/v1/boards/:boardId/columns
**설명**: 커스텀 컬럼 추가
**인증**: Bearer Token 필수
**Request Body**:
```json
{
  "name": "string (required, 1-50자)",
  "position": "number (optional, 미지정 시 마지막 위치)"
}
```
**Response 201**:
```json
{
  "id": "uuid",
  "name": "string",
  "position": 0
}
```

#### PATCH /api/v1/boards/:boardId/columns/:columnId
**설명**: 컬럼 이름 변경 또는 위치 이동
**인증**: Bearer Token 필수
**Request Body**:
```json
{
  "name": "string (optional, 1-50자)",
  "position": "number (optional)"
}
```

#### DELETE /api/v1/boards/:boardId/columns/:columnId
**설명**: 컬럼 삭제 (하위 태스크가 있으면 거부)
**인증**: Bearer Token 필수
**에러 응답**:

| 코드 | 타입 | 설명 |
|---|---|---|
| 422 | COLUMN_NOT_EMPTY | 하위 태스크가 존재하여 삭제 불가 |

---

### 4.5 태스크 API

#### POST /api/v1/boards/:boardId/tasks
**설명**: 태스크 생성
**인증**: Bearer Token 필수
**Request Body**:
```json
{
  "columnId": "uuid (required)",
  "title": "string (required, 1-200자)",
  "description": "string (optional, 최대 5000자)",
  "priority": "string (enum: low, medium, high, urgent, default: medium)",
  "assigneeId": "uuid (optional)",
  "dueDate": "string (optional, YYYY-MM-DD 형식)",
  "labelIds": ["uuid (optional)"]
}
```
**Response 201**:
```json
{
  "id": "uuid",
  "columnId": "uuid",
  "title": "string",
  "description": "string",
  "priority": "string",
  "assigneeId": "uuid | null",
  "dueDate": "string | null",
  "position": 0,
  "version": 1,
  "labels": [],
  "createdAt": "ISO 8601"
}
```
**에러 응답**:

| 코드 | 타입 | 설명 |
|---|---|---|
| 422 | TASK_LIMIT_EXCEEDED | 보드당 500개 태스크 초과 |

#### PATCH /api/v1/boards/:boardId/tasks/:taskId
**설명**: 태스크 수정 (낙관적 잠금 적용)
**인증**: Bearer Token 필수
**Request Body**:
```json
{
  "title": "string (optional)",
  "description": "string (optional)",
  "priority": "string (optional)",
  "assigneeId": "uuid | null (optional)",
  "dueDate": "string | null (optional)",
  "version": "number (required, 현재 버전)"
}
```
**에러 응답**:

| 코드 | 타입 | 설명 |
|---|---|---|
| 409 | VERSION_CONFLICT | 다른 사용자가 이미 수정함. 최신 데이터와 함께 반환 |

#### PATCH /api/v1/boards/:boardId/tasks/:taskId/move
**설명**: 태스크 이동 (드래그앤드롭 - 컬럼 간/컬럼 내 이동)
**인증**: Bearer Token 필수
**Request Body**:
```json
{
  "columnId": "uuid (required, 이동할 컬럼)",
  "position": "number (required, 새 위치)",
  "version": "number (required, 낙관적 잠금)"
}
```

#### DELETE /api/v1/boards/:boardId/tasks/:taskId
**설명**: 태스크 삭제 (소프트 삭제)
**인증**: Bearer Token 필수
**Response 204**: No Content

#### POST /api/v1/boards/:boardId/tasks/:taskId/image
**설명**: 태스크 이미지 첨부
**인증**: Bearer Token 필수
**Request**: multipart/form-data
**제약**: 이미지 파일만(JPEG, PNG, GIF, WebP), 최대 5MB
**Response 200**:
```json
{
  "imageUrl": "string"
}
```
**에러 응답**:

| 코드 | 타입 | 설명 |
|---|---|---|
| 400 | INVALID_FILE_TYPE | 이미지 파일만 허용 |
| 413 | FILE_TOO_LARGE | 5MB 초과 |

---

### 4.6 필터링 API

#### GET /api/v1/boards/:boardId/tasks
**설명**: 보드의 태스크 목록 조회 (필터링 지원)
**인증**: Bearer Token 필수
**Query Parameters**:

| 파라미터 | 타입 | 설명 |
|---|---|---|
| assigneeId | uuid | 담당자 필터 |
| labelId | uuid | 라벨 필터 |
| priority | string | 우선순위 필터 |
| search | string | 제목 검색 (최대 100자) |

**Response 200**:
```json
{
  "columns": [
    {
      "id": "uuid",
      "name": "string",
      "position": 0,
      "tasks": [
        {
          "id": "uuid",
          "title": "string",
          "priority": "string",
          "assignee": { "id": "uuid", "displayName": "string" },
          "dueDate": "string | null",
          "labels": [{ "id": "uuid", "name": "string", "color": "string" }],
          "commentCount": 0,
          "position": 0,
          "version": 1
        }
      ]
    }
  ]
}
```

---

### 4.7 댓글 API

#### POST /api/v1/tasks/:taskId/comments
**설명**: 댓글 작성
**인증**: Bearer Token 필수
**Request Body**:
```json
{
  "content": "string (required, 1-2000자)"
}
```
**Response 201**:
```json
{
  "id": "uuid",
  "content": "string",
  "author": { "id": "uuid", "displayName": "string" },
  "createdAt": "ISO 8601"
}
```

#### GET /api/v1/tasks/:taskId/comments
**설명**: 태스크 댓글 목록 조회
**인증**: Bearer Token 필수
**Query Parameters**: `cursor` (페이지네이션), `limit` (기본 20, 최대 50)

---

### 4.8 알림 API

#### GET /api/v1/notifications
**설명**: 내 알림 목록 조회
**인증**: Bearer Token 필수
**Query Parameters**: `unreadOnly` (boolean), `cursor`, `limit`
**Response 200**:
```json
{
  "notifications": [
    {
      "id": "uuid",
      "type": "task_assigned | due_soon | due_today | comment",
      "payload": {},
      "isRead": false,
      "createdAt": "ISO 8601"
    }
  ],
  "nextCursor": "string | null"
}
```

#### PATCH /api/v1/notifications/:notificationId/read
**설명**: 알림 읽음 처리
**인증**: Bearer Token 필수
**Response 204**: No Content

---

### 4.9 라벨 API

#### POST /api/v1/boards/:boardId/labels
**설명**: 라벨 생성
**인증**: Bearer Token 필수
**Request Body**:
```json
{
  "name": "string (required, 1-30자)",
  "color": "string (required, HEX 형식 #RRGGBB)"
}
```

#### DELETE /api/v1/boards/:boardId/labels/:labelId
**설명**: 라벨 삭제 (연관 task_labels도 함께 삭제)
**인증**: Bearer Token 필수
**Response 204**: No Content

---

### 4.10 실시간 WebSocket 이벤트

#### 연결 및 인증

```
클라이언트 → 서버: connection (headers: { Authorization: "Bearer {accessToken}" })
서버 → 클라이언트: authenticated | auth_error
클라이언트 → 서버: join_board { boardId: "uuid" }
서버 → 클라이언트: board_joined { boardId: "uuid", onlineMembers: [...] }
```

#### 이벤트 목록

| 이벤트명 | 방향 | 페이로드 | 설명 |
|---|---|---|---|
| task:created | 서버→클라이언트 | Task 객체 | 태스크 생성됨 |
| task:updated | 서버→클라이언트 | Task 변경분 + version | 태스크 수정됨 |
| task:moved | 서버→클라이언트 | { taskId, fromColumnId, toColumnId, position, version } | 태스크 이동됨 |
| task:deleted | 서버→클라이언트 | { taskId } | 태스크 삭제됨 |
| column:created | 서버→클라이언트 | Column 객체 | 컬럼 추가됨 |
| column:updated | 서버→클라이언트 | Column 변경분 | 컬럼 수정됨 |
| column:deleted | 서버→클라이언트 | { columnId } | 컬럼 삭제됨 |
| comment:created | 서버→클라이언트 | Comment 객체 | 댓글 추가됨 |
| member:online | 서버→클라이언트 | { userId, displayName } | 멤버 접속 |
| member:offline | 서버→클라이언트 | { userId } | 멤버 접속 해제 |
| notification:new | 서버→클라이언트 | Notification 객체 | 새 알림 도착 |

---

## 5. 폴더 구조 및 컴포넌트 분리 전략

```
taskflow/
├── docker-compose.yml              # PostgreSQL, Redis, 앱 서버 구성
├── .env.example                     # 환경변수 템플릿 (실제 .env는 gitignore)
│
├── client/                          # 프론트엔드 (Next.js)
│   ├── next.config.js
│   ├── tailwind.config.js
│   ├── tsconfig.json
│   ├── public/                      # 정적 파일
│   ├── src/
│   │   ├── app/                     # Next.js App Router 페이지
│   │   │   ├── layout.tsx           # 루트 레이아웃
│   │   │   ├── page.tsx             # 랜딩/로그인 페이지
│   │   │   ├── (auth)/              # 인증 관련 페이지 그룹
│   │   │   │   ├── login/page.tsx
│   │   │   │   └── register/page.tsx
│   │   │   └── (dashboard)/         # 인증 필요 페이지 그룹
│   │   │       ├── layout.tsx       # 사이드바 포함 레이아웃
│   │   │       ├── teams/page.tsx   # 팀 목록
│   │   │       └── boards/
│   │   │           └── [boardId]/
│   │   │               └── page.tsx # 칸반보드 메인 뷰
│   │   ├── components/              # 재사용 UI 컴포넌트
│   │   │   ├── ui/                  # 기본 UI (Button, Input, Modal, Toast)
│   │   │   ├── board/               # 보드 관련
│   │   │   │   ├── KanbanBoard.tsx  # 칸반보드 컨테이너
│   │   │   │   ├── BoardColumn.tsx  # 컬럼 컴포넌트
│   │   │   │   ├── TaskCard.tsx     # 태스크 카드
│   │   │   │   ├── TaskDetail.tsx   # 태스크 상세 모달
│   │   │   │   └── FilterBar.tsx    # 필터링 바
│   │   │   ├── team/                # 팀 관련
│   │   │   │   ├── TeamList.tsx
│   │   │   │   ├── MemberList.tsx
│   │   │   │   └── InviteModal.tsx
│   │   │   └── notification/        # 알림 관련
│   │   │       ├── NotificationBell.tsx
│   │   │       └── NotificationList.tsx
│   │   ├── hooks/                   # 커스텀 훅
│   │   │   ├── useAuth.ts           # 인증 상태 관리
│   │   │   ├── useSocket.ts         # Socket.IO 연결 관리
│   │   │   ├── useBoard.ts          # 보드 데이터 훅
│   │   │   └── useDragAndDrop.ts    # 드래그앤드롭 로직
│   │   ├── services/                # API 통신 레이어
│   │   │   ├── apiClient.ts         # Axios 인스턴스 (인터셉터, 토큰 자동 갱신)
│   │   │   ├── authService.ts
│   │   │   ├── teamService.ts
│   │   │   ├── boardService.ts
│   │   │   ├── taskService.ts
│   │   │   └── notificationService.ts
│   │   ├── stores/                  # Zustand 전역 상태
│   │   │   ├── authStore.ts         # 인증 상태
│   │   │   ├── boardStore.ts        # 현재 보드 상태
│   │   │   └── notificationStore.ts # 알림 상태
│   │   ├── types/                   # TypeScript 타입 정의
│   │   │   ├── auth.ts
│   │   │   ├── board.ts
│   │   │   ├── task.ts
│   │   │   └── api.ts               # API 응답 공통 타입
│   │   ├── utils/                   # 유틸리티 함수
│   │   │   ├── validation.ts        # 입력 검증
│   │   │   └── date.ts              # 날짜 포맷팅
│   │   └── lib/                     # 외부 라이브러리 설정
│   │       └── socket.ts            # Socket.IO 클라이언트 초기화
│   └── __tests__/                   # 프론트엔드 테스트
│
├── server/                          # 백엔드 (Express)
│   ├── tsconfig.json
│   ├── src/
│   │   ├── index.ts                 # 서버 엔트리포인트
│   │   ├── app.ts                   # Express 앱 설정
│   │   ├── config/                  # 설정
│   │   │   ├── env.ts               # 환경변수 로딩 및 검증
│   │   │   ├── database.ts          # Prisma 클라이언트 초기화
│   │   │   └── redis.ts             # Redis 클라이언트 초기화
│   │   ├── middleware/              # Express 미들웨어
│   │   │   ├── auth.ts              # JWT 인증 미들웨어
│   │   │   ├── authorize.ts         # 역할 기반 인가 미들웨어
│   │   │   ├── validate.ts          # zod 스키마 검증 미들웨어
│   │   │   ├── rateLimiter.ts       # Rate Limiting
│   │   │   ├── errorHandler.ts      # 전역 에러 핸들러
│   │   │   └── security.ts          # Helmet, CORS, CSP 설정
│   │   ├── routes/                  # 라우트 정의
│   │   │   ├── authRoutes.ts
│   │   │   ├── teamRoutes.ts
│   │   │   ├── boardRoutes.ts
│   │   │   ├── taskRoutes.ts
│   │   │   ├── commentRoutes.ts
│   │   │   └── notificationRoutes.ts
│   │   ├── controllers/             # 요청 처리
│   │   │   ├── authController.ts
│   │   │   ├── teamController.ts
│   │   │   ├── boardController.ts
│   │   │   ├── taskController.ts
│   │   │   ├── commentController.ts
│   │   │   └── notificationController.ts
│   │   ├── services/                # 비즈니스 로직
│   │   │   ├── authService.ts
│   │   │   ├── teamService.ts
│   │   │   ├── boardService.ts
│   │   │   ├── taskService.ts
│   │   │   ├── commentService.ts
│   │   │   ├── notificationService.ts
│   │   │   └── emailService.ts
│   │   ├── socket/                  # Socket.IO 관련
│   │   │   ├── index.ts             # Socket.IO 서버 초기화
│   │   │   ├── authHandler.ts       # 소켓 인증
│   │   │   ├── boardHandler.ts      # 보드 이벤트 핸들러
│   │   │   └── presenceHandler.ts   # 온라인/오프라인 상태
│   │   ├── schemas/                 # zod 검증 스키마
│   │   │   ├── authSchema.ts
│   │   │   ├── teamSchema.ts
│   │   │   ├── boardSchema.ts
│   │   │   └── taskSchema.ts
│   │   ├── errors/                  # 커스텀 에러 클래스
│   │   │   ├── AppError.ts          # 기본 에러 클래스
│   │   │   ├── ValidationError.ts
│   │   │   ├── AuthError.ts
│   │   │   └── ConflictError.ts     # 낙관적 잠금 충돌
│   │   ├── utils/                   # 유틸리티
│   │   │   ├── logger.ts            # 구조화된 로깅 (pino)
│   │   │   ├── crypto.ts            # 비밀번호 해싱, 토큰 생성
│   │   │   └── fileUpload.ts        # 이미지 업로드 처리 (multer)
│   │   └── jobs/                    # 스케줄 작업
│   │       └── dueDateChecker.ts    # 마감일 알림 cron (매일 09:00)
│   ├── prisma/
│   │   ├── schema.prisma            # Prisma 스키마 정의
│   │   └── migrations/              # DB 마이그레이션
│   └── __tests__/                   # 백엔드 테스트
│       ├── unit/
│       └── integration/
│
└── shared/                          # 프론트/백엔드 공유
    └── types/                       # 공유 타입 정의
        ├── events.ts                # WebSocket 이벤트 타입
        └── enums.ts                 # 공통 Enum (Priority, Role 등)
```

### 컴포넌트 분리 기준

- **프론트엔드**: 페이지(app/) → 도메인별 컴포넌트(components/) → UI 원자 컴포넌트(components/ui/) 순으로 계층 분리. 비즈니스 로직은 hooks/에, API 통신은 services/에 격리
- **백엔드**: Route → Controller → Service → Prisma(Model) 레이어 분리. Controller는 요청/응답 변환만, Service에 비즈니스 로직 집중

---

## 6. 상태 관리 및 전역 상태 흐름도

### 6.1 상태 분리 기준

| 상태 유형 | 관리 도구 | 예시 |
|---|---|---|
| 전역 클라이언트 상태 | Zustand | 현재 사용자 정보, 인증 토큰, 알림 배지 카운트 |
| 서버 상태 (캐시) | TanStack Query | 보드 데이터, 태스크 목록, 팀 멤버 목록 |
| 실시간 상태 | Socket.IO + Zustand | 온라인 멤버, 실시간 태스크 변경 |
| 로컬 UI 상태 | React useState | 모달 열림/닫힘, 필터 선택, 드래그 중 상태 |

### 6.2 실시간 동기화 흐름도

```mermaid
sequenceDiagram
    participant A as 사용자 A (브라우저)
    participant API as API 서버
    participant DB as PostgreSQL
    participant Redis as Redis Pub/Sub
    participant WS as Socket.IO 서버
    participant B as 사용자 B (브라우저)

    Note over A,B: 같은 보드를 보고 있는 팀원

    A->>API: PATCH /tasks/:id/move (version: 3)
    API->>DB: UPDATE task SET column_id, position, version=4 WHERE version=3
    DB-->>API: 성공 (1 row affected)
    API-->>A: 200 OK (새 버전: 4)
    API->>Redis: PUBLISH board:{boardId} task:moved
    Redis->>WS: 이벤트 전달
    WS->>B: task:moved { taskId, toColumnId, position, version: 4 }
    B->>B: TanStack Query 캐시 업데이트 (invalidation)
```

### 6.3 낙관적 업데이트 + 충돌 해결 흐름

```mermaid
stateDiagram-v2
    [*] --> Idle: 보드 로딩 완료
    Idle --> Dragging: 드래그 시작
    Dragging --> OptimisticUpdate: 드롭
    OptimisticUpdate --> Idle: API 성공 (서버 버전 반영)
    OptimisticUpdate --> ConflictDetected: API 409 (버전 충돌)
    ConflictDetected --> ServerSync: 서버 최신 데이터 fetch
    ServerSync --> Idle: UI에 최신 상태 반영 + 사용자 알림
```

### 6.4 인증 상태 흐름

```mermaid
stateDiagram-v2
    [*] --> LoggedOut
    LoggedOut --> LoggingIn: 로그인 요청
    LoggingIn --> LoggedIn: 성공 (토큰 저장)
    LoggingIn --> LoggedOut: 실패
    LoggedIn --> TokenRefreshing: Access Token 만료
    TokenRefreshing --> LoggedIn: Refresh 성공
    TokenRefreshing --> LoggedOut: Refresh 실패 (재로그인)
    LoggedIn --> LoggedOut: 로그아웃
```

---

## 7. 보안 설계

### 7.1 OWASP Top 10 대응

| 위협 | 대응 방안 | 구현 위치 |
|---|---|---|
| A01: Broken Access Control | 팀 멤버십 + 역할(admin/member) 기반 인가 검증. 모든 API에서 리소스 소유권 확인 | authorize 미들웨어, 각 서비스 레이어 |
| A02: Cryptographic Failures | 비밀번호 PBKDF2 해싱(SHA-256, 10만회), JWT 서명(HS256→환경변수 시크릿), HTTPS 강제 | crypto.ts, auth 미들웨어 |
| A03: Injection | Prisma ORM(Parameterized Query 강제), zod 스키마 검증으로 입력 타입/길이 제한 | validate 미들웨어, Prisma |
| A05: Security Misconfiguration | Helmet.js(보안 헤더), CORS 화이트리스트, CSP 설정, 불필요한 응답 헤더 제거 | security.ts 미들웨어 |
| A07: XSS | 프론트엔드 React 자동 이스케이핑 + DOMPurify로 사용자 입력 sanitize, CSP 헤더 | 프론트엔드 컴포넌트, security.ts |
| A08: SSRF | 이미지 업로드 시 파일 타입 검증(매직바이트), URL 기반 업로드 미지원 | fileUpload.ts |
| A09: Logging & Monitoring | pino 구조화 로깅, 민감 정보(password, token) 자동 마스킹, 요청 ID 추적 | logger.ts |

### 7.2 인증/인가 흐름

```mermaid
sequenceDiagram
    participant C as 클라이언트
    participant M as Auth 미들웨어
    participant A as Auth 서비스
    participant DB as 데이터베이스
    participant R as Redis

    Note over C,R: 로그인 흐름
    C->>A: POST /auth/login { email, password }
    A->>R: 로그인 실패 카운터 확인 (5회 초과 시 423 반환)
    A->>DB: 이메일로 사용자 조회
    A->>A: PBKDF2 비밀번호 검증
    A->>R: Refresh Token 저장 (7일 TTL)
    A-->>C: { accessToken (15분), refreshToken (7일) }

    Note over C,R: API 요청 흐름
    C->>M: GET /api/v1/boards/:id (Authorization: Bearer {token})
    M->>M: JWT 서명 검증 + 만료 확인
    M->>M: 팀 멤버십 + 역할 검증
    M-->>C: 200 OK 또는 401/403
```

### 7.3 민감 데이터 처리 정책

| 데이터 | 저장 방식 | 로그 출력 | 전송 방식 |
|---|---|---|---|
| 비밀번호 | PBKDF2 해시 (salt 32바이트, 10만회 반복) | 절대 출력 금지 | HTTPS만 허용 |
| Access Token | 클라이언트 메모리(Zustand) | 마스킹 (앞 10자만) | Authorization 헤더 |
| Refresh Token | httpOnly + Secure + SameSite=Strict 쿠키 + Redis | 절대 출력 금지 | httpOnly 쿠키 |
| 이메일 | 평문 저장 (검색 필요) | 마스킹 (a**@**.com) | HTTPS |

### 7.4 환경변수 관리

```
# .env.example (실제 값은 절대 커밋 금지)
DATABASE_URL=postgresql://user:password@localhost:5432/taskflow
REDIS_URL=redis://localhost:6379
JWT_SECRET=                          # 최소 256비트 랜덤 키
JWT_REFRESH_SECRET=                  # 별도 시크릿
SMTP_HOST=
SMTP_PORT=
SMTP_USER=
SMTP_PASS=
CORS_ORIGIN=http://localhost:3000
NODE_ENV=development
```

- `.env` 파일은 `.gitignore`에 반드시 포함
- 프로덕션 환경에서는 Docker secrets 또는 환경변수 주입 방식 사용
- 모든 환경변수는 서버 시작 시 zod 스키마로 존재 여부와 형식 검증

### 7.5 Rate Limiting

| 엔드포인트 | 제한 | 윈도우 |
|---|---|---|
| POST /auth/login | 5회 | 15분 |
| POST /auth/register | 3회 | 1시간 |
| POST /teams/:id/members/invite | 10회 | 1시간 |
| 기타 API | 100회 | 1분 |

### 7.6 의존성 라이선스 검토

모든 의존성 패키지는 MIT 또는 Apache 2.0 라이선스만 허용한다. GPL 계열(GPL, LGPL, AGPL) 라이선스는 소스 공개 의무가 있으므로 사용을 금지한다.

| 패키지 | 라이선스 | 상태 |
|---|---|---|
| react, next | MIT | 허용 |
| express | MIT | 허용 |
| socket.io | MIT | 허용 |
| prisma | Apache 2.0 | 허용 |
| zustand | MIT | 허용 |
| @tanstack/react-query | MIT | 허용 |
| @dnd-kit/core | MIT | 허용 |
| tailwindcss | MIT | 허용 |
| zod | MIT | 허용 |
| pino | MIT | 허용 |
| helmet | MIT | 허용 |
| jose | MIT | 허용 |
| redis (ioredis) | MIT | 허용 |
| multer | MIT | 허용 |

---

## 8. 예외 처리 및 에러 전략

### 8.1 에러 분류 체계

| 분류 | HTTP 코드 범위 | 예시 | 사용자 노출 |
|---|---|---|---|
| 비즈니스 에러 | 400-422 | 입력 검증 실패, 팀 인원 초과, 버전 충돌 | 구체적 메시지 노출 |
| 인증/인가 에러 | 401, 403 | 토큰 만료, 권한 없음 | 일반적 메시지만 노출 |
| 시스템 에러 | 500 | DB 연결 실패, 외부 서비스 오류 | "서버 오류가 발생했습니다" (내부 상세는 숨김) |

### 8.2 에러 응답 표준 포맷 (RFC 7807 Problem Details)

```json
{
  "type": "https://taskflow.internal/errors/version-conflict",
  "title": "VERSION_CONFLICT",
  "status": 409,
  "detail": "다른 팀원이 이 태스크를 이미 수정했습니다. 최신 데이터를 확인해주세요.",
  "instance": "/api/v1/boards/abc/tasks/xyz",
  "traceId": "req-uuid-1234"
}
```

### 8.3 로깅 전략

| 레벨 | 용도 | 예시 |
|---|---|---|
| error | 시스템 오류, 복구 불가 상황 | DB 연결 실패, 미처리 예외 |
| warn | 비정상적이지만 복구 가능 | 로그인 반복 실패, Rate Limit 도달 |
| info | 주요 비즈니스 이벤트 | 회원가입, 팀 생성, 태스크 이동 |
| debug | 개발 디버깅용 (프로덕션 비활성화) | SQL 쿼리, 요청/응답 상세 |

**민감 정보 마스킹 규칙**:
- `password`, `token`, `secret` 키가 포함된 필드는 자동으로 `[REDACTED]`로 대체
- 이메일은 `a**@**.com` 형식으로 마스킹
- 모든 로그에 `traceId`(요청 ID) 포함으로 추적 가능

### 8.4 재시도 정책

| 대상 | 재시도 횟수 | 간격 | 조건 |
|---|---|---|---|
| DB 쿼리 | 2회 | 지수 백오프 (100ms, 200ms) | 연결 오류일 때만 |
| 이메일 발송 | 3회 | 지수 백오프 (1s, 2s, 4s) | 타임아웃/연결 오류 |
| Socket 재연결 | 무제한 | 지수 백오프 (1s~30s) | 클라이언트 자동 처리 |

**멱등성 보장 필요 API**:
- `PATCH /tasks/:id/move` - version 필드로 멱등성 보장 (동일 version으로 재요청 시 무시)
- `POST /teams/:id/members/invite` - 이메일 중복 체크로 중복 초대 방지

---

## 9. 3줄 요약 및 비유

> **3줄 요약**
> 1. 팀원들이 브라우저에서 칸반보드를 열면, 태스크의 생성/이동/수정이 WebSocket을 통해 모든 팀원에게 즉시 동기화됩니다.
> 2. 두 사람이 동시에 같은 태스크를 수정하면 버전 번호로 충돌을 감지하여, 나중에 수정한 사람에게 최신 데이터를 보여주고 재시도를 안내합니다.
> 3. 모든 통신은 암호화되고, 팀 멤버만 해당 보드에 접근할 수 있으며, 비밀번호와 토큰은 철저히 보호됩니다.
>
> **비유로 이해하기**
> TaskFlow는 공유 화이트보드가 있는 회의실과 비슷합니다.
> 팀원(사용자)이 포스트잇(태스크)을 화이트보드(칸반보드)에 붙이거나 옮기면,
> 같은 회의실에 있는 모든 사람(같은 보드를 보는 팀원)이 그 변화를 즉시 볼 수 있습니다.
> 만약 두 사람이 동시에 같은 포스트잇을 잡으려 하면, 먼저 잡은 사람의 변경이 적용되고
> 나중 사람에게는 "이미 누가 옮겼어요, 다시 확인해주세요"라고 알려줍니다.
> 회의실 문(인증)은 초대받은 팀원만 열 수 있고, 관리자만 새 사람을 초대할 수 있습니다.
