# 시스템 설계서: AnonVoice

## 1. 프로젝트 개요

| 항목 | 내용 |
|---|---|
| 프로젝트명 | AnonVoice |
| 한 줄 설명 | 직원이 익명으로 팀/조직에 피드백을 남기고, 관리자가 대시보드로 확인하는 사내 서비스 |
| 기술 스택 | React 18 + Next.js 14, Zustand, Tailwind CSS + shadcn/ui, Node.js + Express, PostgreSQL 16, Prisma, NextAuth.js, Vercel + AWS ECS, Sentry + Datadog |
| 작성일 | 2026-03-26 |
| 기반 문서 | prd.md |

---

## 2. 전체 시스템 아키텍처

```mermaid
flowchart TB
    subgraph Client["클라이언트 (Vercel)"]
        NextApp["Next.js 14 App Router<br/>React 18 + Zustand"]
    end

    subgraph Backend["백엔드 (AWS ECS)"]
        API["Express API 서버"]
        AuthModule["NextAuth.js<br/>인증 모듈"]
        KeywordDetector["위험 키워드<br/>감지 엔진"]
        ReportGenerator["리포트 생성<br/>엔진 (PDF)"]
        NotificationService["알림 서비스"]
    end

    subgraph DataLayer["데이터 레이어"]
        DB[(PostgreSQL 16)]
        Cache[(Redis<br/>세션/캐시)]
    end

    subgraph External["외부 서비스"]
        SSO["사내 SSO"]
        EmailSMTP["이메일 SMTP"]
        Sentry["Sentry<br/>에러 트래킹"]
        Datadog["Datadog<br/>모니터링"]
    end

    NextApp -->|"HTTPS API 호출"| API
    NextApp -->|"SSO 인증 요청"| AuthModule
    AuthModule -->|"OAuth/SAML"| SSO
    API --> DB
    API --> Cache
    API --> KeywordDetector
    API --> ReportGenerator
    API --> NotificationService
    NotificationService -->|"이메일 발송"| EmailSMTP
    API -->|"에러 보고"| Sentry
    API -->|"메트릭 전송"| Datadog
```

### 아키텍처 설명

- **클라이언트**: Next.js App Router 기반 SSR/CSR 혼합. Zustand로 클라이언트 상태 관리.
- **API 서버**: Express 기반 RESTful API. 인증/인가 미들웨어 적용.
- **데이터 레이어**: PostgreSQL을 메인 DB로 사용하고, Redis를 세션 저장소 및 캐시로 활용.
- **알림 서비스**: 이메일(SMTP) 및 인앱 알림 처리. 위험 키워드 감지 시 HR에 즉시 알림.
- **모니터링**: Sentry(에러 트래킹) + Datadog(인프라/APM 모니터링).

---

## 3. 데이터베이스 모델링

### 3.1 ERD

```mermaid
erDiagram
    USERS {
        uuid id PK
        varchar email "UK, 사내 이메일"
        varchar name
        varchar role "EMPLOYEE | TEAM_LEADER | HR_ADMIN"
        uuid team_id FK
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at "소프트 삭제"
    }

    TEAMS {
        uuid id PK
        varchar name "UK"
        varchar department
        timestamp created_at
    }

    FEEDBACKS {
        uuid id PK
        uuid target_team_id FK "피드백 대상 팀"
        varchar category "CULTURE | PROCESS | LEADERSHIP | OTHER"
        text content "AES-256 암호화 저장"
        varchar sentiment "POSITIVE | NEUTRAL | NEGATIVE"
        varchar anonymous_token_hash "SHA-256 해시, 익명 대화용"
        timestamp created_at
        timestamp expires_at "2년 후 자동 삭제 기준일"
    }

    CONVERSATIONS {
        uuid id PK
        uuid feedback_id FK
        text content "AES-256 암호화 저장"
        varchar author_type "MANAGER | ANONYMOUS_AUTHOR"
        uuid manager_id FK "NULL if ANONYMOUS_AUTHOR"
        int sequence_number "1~10, 최대 5회 왕복"
        timestamp created_at
    }

    NOTIFICATIONS {
        uuid id PK
        uuid user_id FK "알림 수신자"
        varchar type "NEW_FEEDBACK | NEW_REPLY | RISK_KEYWORD"
        varchar channel "EMAIL | IN_APP"
        jsonb metadata "관련 feedback_id 등"
        boolean is_read "기본 false"
        timestamp created_at
    }

    AUDIT_LOGS {
        uuid id PK
        uuid actor_id FK "행위자"
        varchar action "VIEW_FEEDBACK | REPLY_FEEDBACK | EXPORT_REPORT"
        uuid resource_id "대상 리소스 ID"
        varchar resource_type "FEEDBACK | REPORT"
        jsonb context "IP, User-Agent 등"
        timestamp created_at
    }

    RISK_ALERTS {
        uuid id PK
        uuid feedback_id FK
        varchar detected_keywords "감지된 키워드 목록"
        varchar status "PENDING | REVIEWED | DISMISSED"
        uuid reviewed_by FK "NULL if PENDING"
        timestamp created_at
        timestamp reviewed_at
    }

    REPORTS {
        uuid id PK
        uuid team_id FK "NULL이면 전사 리포트"
        varchar period "2026-Q1 형식"
        varchar report_type "QUARTERLY | MONTHLY"
        varchar file_path "S3 경로 (암호화 저장)"
        timestamp generated_at
    }

    TEAMS ||--o{ USERS : "소속"
    TEAMS ||--o{ FEEDBACKS : "대상"
    FEEDBACKS ||--o{ CONVERSATIONS : "쓰레드"
    FEEDBACKS ||--o{ RISK_ALERTS : "위험 감지"
    USERS ||--o{ NOTIFICATIONS : "수신"
    USERS ||--o{ AUDIT_LOGS : "행위"
    TEAMS ||--o{ REPORTS : "리포트"
```

### 3.2 테이블 상세 명세

#### USERS 테이블

| 컬럼명 | 타입 | 제약조건 | 설명 |
|---|---|---|---|
| id | UUID | PK, DEFAULT gen_random_uuid() | 사용자 고유 ID |
| email | VARCHAR(255) | NOT NULL, UNIQUE | 사내 이메일 (SSO 연동 식별자) |
| name | VARCHAR(100) | NOT NULL | 사용자 이름 |
| role | VARCHAR(20) | NOT NULL, CHECK(IN) | EMPLOYEE, TEAM_LEADER, HR_ADMIN |
| team_id | UUID | FK(teams.id), NOT NULL | 소속 팀 |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() | 생성일 |
| updated_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() | 수정일 |
| deleted_at | TIMESTAMPTZ | NULL | 소프트 삭제 |

- 인덱스: `idx_users_email` (UNIQUE), `idx_users_team_id`, `idx_users_role`

#### FEEDBACKS 테이블

| 컬럼명 | 타입 | 제약조건 | 설명 |
|---|---|---|---|
| id | UUID | PK, DEFAULT gen_random_uuid() | 피드백 고유 ID |
| target_team_id | UUID | FK(teams.id), NOT NULL | 피드백 대상 팀 |
| category | VARCHAR(20) | NOT NULL, CHECK(IN) | CULTURE, PROCESS, LEADERSHIP, OTHER |
| content | TEXT | NOT NULL | 본문 (AES-256-GCM 암호화) |
| sentiment | VARCHAR(10) | NOT NULL, CHECK(IN) | POSITIVE, NEUTRAL, NEGATIVE |
| anonymous_token_hash | VARCHAR(64) | NOT NULL | 1회용 익명 토큰의 SHA-256 해시 |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() | 생성일 |
| expires_at | TIMESTAMPTZ | NOT NULL | created_at + 2년 (자동 삭제 기준) |

- 인덱스: `idx_feedbacks_target_team_id`, `idx_feedbacks_category`, `idx_feedbacks_created_at`, `idx_feedbacks_expires_at`
- **핵심 보안**: `author_id` 컬럼은 의도적으로 존재하지 않음. 작성자와 피드백의 매핑 정보를 서버에 저장하지 않아 익명성을 보장.

#### CONVERSATIONS 테이블

| 컬럼명 | 타입 | 제약조건 | 설명 |
|---|---|---|---|
| id | UUID | PK | 대화 고유 ID |
| feedback_id | UUID | FK(feedbacks.id), NOT NULL | 피드백 참조 |
| content | TEXT | NOT NULL | 대화 내용 (AES-256-GCM 암호화) |
| author_type | VARCHAR(20) | NOT NULL, CHECK(IN) | MANAGER, ANONYMOUS_AUTHOR |
| manager_id | UUID | FK(users.id), NULL | 관리자인 경우만 |
| sequence_number | INT | NOT NULL, CHECK(1..10) | 순서 번호 (최대 5회 왕복 = 10개) |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() | 생성일 |

- 인덱스: `idx_conversations_feedback_id`
- 유니크 제약: `(feedback_id, sequence_number)` -- 순서 중복 방지

#### AUDIT_LOGS 테이블

| 컬럼명 | 타입 | 제약조건 | 설명 |
|---|---|---|---|
| id | UUID | PK | 로그 고유 ID |
| actor_id | UUID | FK(users.id), NOT NULL | 행위자 |
| action | VARCHAR(50) | NOT NULL | 행위 유형 |
| resource_id | UUID | NOT NULL | 대상 리소스 |
| resource_type | VARCHAR(30) | NOT NULL | 리소스 유형 |
| context | JSONB | NULL | IP, User-Agent 등 메타 정보 |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() | 생성일 |

- 인덱스: `idx_audit_logs_actor_id`, `idx_audit_logs_created_at`
- **보안 참고**: 감사 로그는 UPDATE/DELETE 불가. 삽입 전용(append-only) 테이블.

### 3.3 데이터 보관 및 삭제 정책

- `feedbacks.expires_at` 기준으로 PostgreSQL `pg_cron` 또는 애플리케이션 스케줄러가 매일 만료 데이터를 배치 삭제
- 관련 `conversations`, `risk_alerts`도 CASCADE 삭제
- 감사 로그(`audit_logs`)는 별도 보관 정책 적용 (5년 보관)
- 삭제 작업 자체도 감사 로그에 기록

---

## 4. 핵심 API 인터페이스 명세

### 4.1 인증

#### POST /api/v1/auth/login
**설명**: SSO를 통한 로그인 (NextAuth.js 처리)
**인증**: 불필요 (로그인 엔드포인트)
**Request Body**:
```json
{
  "provider": "string (required, 'sso')",
  "callbackUrl": "string (optional)"
}
```
**Response 200**:
```json
{
  "accessToken": "string (JWT)",
  "refreshToken": "string",
  "expiresIn": 3600,
  "user": {
    "id": "uuid",
    "name": "string",
    "role": "EMPLOYEE | TEAM_LEADER | HR_ADMIN",
    "teamId": "uuid"
  }
}
```
**에러 응답**:

| 코드 | 타입 | 설명 |
|---|---|---|
| 401 | SSO_AUTH_FAILED | SSO 인증 실패 |
| 503 | SSO_UNAVAILABLE | SSO 서버 연결 불가 |

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
  "accessToken": "string (JWT)",
  "expiresIn": 3600
}
```

---

### 4.2 피드백

#### POST /api/v1/feedbacks
**설명**: 익명 피드백 작성
**인증**: Bearer Token 필수 (작성자 정보는 인증 확인 후 즉시 폐기, 피드백과 연결하지 않음)
**Request Body**:
```json
{
  "targetTeamId": "uuid (required)",
  "category": "string (required, enum: CULTURE | PROCESS | LEADERSHIP | OTHER)",
  "content": "string (required, 10-1000자)",
  "sentiment": "string (required, enum: POSITIVE | NEUTRAL | NEGATIVE)"
}
```
**Response 201**:
```json
{
  "feedbackId": "uuid",
  "anonymousToken": "string (1회용 익명 토큰, 클라이언트만 보관)",
  "createdAt": "ISO 8601"
}
```
**에러 응답**:

| 코드 | 타입 | 설명 |
|---|---|---|
| 400 | INVALID_INPUT | content가 10자 미만 또는 1000자 초과 |
| 400 | INVALID_CATEGORY | 유효하지 않은 카테고리 |
| 400 | INVALID_TEAM | 존재하지 않는 팀 ID |
| 401 | UNAUTHORIZED | 인증 토큰 누락/만료 |

**보안 핵심 로직**:
1. 인증 토큰으로 `유효한 사내 직원인지`만 확인
2. 확인 후 요청 컨텍스트에서 사용자 식별 정보를 즉시 폐기
3. `anonymous_token`을 생성하여 클라이언트에 반환 (서버에는 SHA-256 해시만 저장)

#### GET /api/v1/feedbacks
**설명**: 팀별 피드백 목록 조회 (관리자용)
**인증**: Bearer Token 필수 (TEAM_LEADER 이상)
**Query Params**:

| 파라미터 | 타입 | 필수 | 설명 |
|---|---|---|---|
| teamId | uuid | No | 대상 팀 (기본: 본인 팀). HR_ADMIN은 다른 팀도 조회 가능 |
| category | string | No | 카테고리 필터 |
| sentiment | string | No | 감정 필터 |
| page | int | No | 페이지 번호 (기본: 1) |
| limit | int | No | 페이지당 항목 수 (기본: 20, 최대: 100) |
| sort | string | No | 정렬 (기본: createdAt:desc) |

**Response 200**:
```json
{
  "data": [
    {
      "id": "uuid",
      "targetTeamId": "uuid",
      "category": "CULTURE",
      "content": "string",
      "sentiment": "POSITIVE",
      "conversationCount": 3,
      "createdAt": "ISO 8601"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "totalPages": 8
  }
}
```
**에러 응답**:

| 코드 | 타입 | 설명 |
|---|---|---|
| 401 | UNAUTHORIZED | 인증 토큰 누락/만료 |
| 403 | FORBIDDEN | 해당 팀 피드백 열람 권한 없음 |

---

### 4.3 익명 대화

#### POST /api/v1/feedbacks/:feedbackId/conversations
**설명**: 피드백에 대한 답변/후속 코멘트 작성
**인증**: 관리자는 Bearer Token, 익명 작성자는 Anonymous Token
**Request Body**:
```json
{
  "content": "string (required, 1-500자)",
  "anonymousToken": "string (optional, 익명 작성자인 경우)"
}
```
**Response 201**:
```json
{
  "id": "uuid",
  "feedbackId": "uuid",
  "authorType": "MANAGER | ANONYMOUS_AUTHOR",
  "sequenceNumber": 3,
  "content": "string",
  "createdAt": "ISO 8601"
}
```
**에러 응답**:

| 코드 | 타입 | 설명 |
|---|---|---|
| 400 | CONVERSATION_LIMIT_REACHED | 최대 왕복 횟수(5회) 초과 |
| 401 | UNAUTHORIZED | 인증 실패 |
| 403 | INVALID_ANONYMOUS_TOKEN | 익명 토큰이 해당 피드백과 불일치 |
| 404 | FEEDBACK_NOT_FOUND | 존재하지 않는 피드백 |

#### GET /api/v1/feedbacks/:feedbackId/conversations
**설명**: 피드백의 대화 목록 조회
**인증**: Bearer Token (관리자) 또는 Anonymous Token (작성자)
**Response 200**:
```json
{
  "feedbackId": "uuid",
  "conversations": [
    {
      "id": "uuid",
      "authorType": "MANAGER",
      "content": "string",
      "sequenceNumber": 1,
      "createdAt": "ISO 8601"
    }
  ]
}
```

---

### 4.4 알림

#### GET /api/v1/notifications
**설명**: 사용자 알림 목록 조회
**인증**: Bearer Token 필수
**Query Params**:

| 파라미터 | 타입 | 필수 | 설명 |
|---|---|---|---|
| isRead | boolean | No | 읽음 여부 필터 |
| page | int | No | 페이지 번호 (기본: 1) |
| limit | int | No | 페이지당 항목 수 (기본: 20) |

**Response 200**:
```json
{
  "data": [
    {
      "id": "uuid",
      "type": "NEW_FEEDBACK",
      "channel": "IN_APP",
      "metadata": { "feedbackId": "uuid", "teamName": "string" },
      "isRead": false,
      "createdAt": "ISO 8601"
    }
  ],
  "unreadCount": 5,
  "pagination": { "page": 1, "limit": 20, "total": 30 }
}
```

#### PATCH /api/v1/notifications/:id/read
**설명**: 알림을 읽음 처리
**인증**: Bearer Token 필수
**Response 200**:
```json
{
  "id": "uuid",
  "isRead": true
}
```

---

### 4.5 통계 및 리포트

#### GET /api/v1/statistics/dashboard
**설명**: 대시보드용 통계 데이터 조회
**인증**: Bearer Token 필수 (TEAM_LEADER 이상)
**Query Params**:

| 파라미터 | 타입 | 필수 | 설명 |
|---|---|---|---|
| teamId | uuid | No | 대상 팀 (기본: 본인 팀) |
| period | string | No | 기간 (기본: 최근 30일, 형식: YYYY-MM) |

**Response 200**:
```json
{
  "totalFeedbacks": 45,
  "sentimentDistribution": {
    "positive": 20,
    "neutral": 15,
    "negative": 10
  },
  "categoryDistribution": {
    "culture": 12,
    "process": 18,
    "leadership": 10,
    "other": 5
  },
  "monthlyTrend": [
    { "month": "2026-01", "count": 15, "avgSentiment": 0.6 },
    { "month": "2026-02", "count": 18, "avgSentiment": 0.5 }
  ],
  "participationRate": 0.72,
  "topKeywords": [
    { "keyword": "소통", "count": 8 },
    { "keyword": "회의", "count": 6 }
  ]
}
```

#### POST /api/v1/reports/generate
**설명**: 분기별 리포트 수동 생성 요청
**인증**: Bearer Token 필수 (HR_ADMIN만)
**Request Body**:
```json
{
  "teamId": "uuid (optional, NULL이면 전사 리포트)",
  "period": "string (required, 예: 2026-Q1)"
}
```
**Response 202**:
```json
{
  "reportId": "uuid",
  "status": "GENERATING",
  "estimatedCompletionTime": "ISO 8601"
}
```

#### GET /api/v1/reports/:reportId
**설명**: 생성된 리포트 다운로드
**인증**: Bearer Token 필수 (HR_ADMIN만)
**Response 200**: PDF 파일 스트림 (Content-Type: application/pdf)

---

### 4.6 위험 키워드 알림 (HR 전용)

#### GET /api/v1/risk-alerts
**설명**: 위험 키워드 감지 알림 목록
**인증**: Bearer Token 필수 (HR_ADMIN만)
**Query Params**:

| 파라미터 | 타입 | 필수 | 설명 |
|---|---|---|---|
| status | string | No | PENDING, REVIEWED, DISMISSED |
| page | int | No | 페이지 번호 |

**Response 200**:
```json
{
  "data": [
    {
      "id": "uuid",
      "feedbackId": "uuid",
      "detectedKeywords": ["퇴사", "번아웃"],
      "feedbackPreview": "string (앞 50자)",
      "status": "PENDING",
      "createdAt": "ISO 8601"
    }
  ],
  "pagination": { "page": 1, "limit": 20, "total": 5 }
}
```

#### PATCH /api/v1/risk-alerts/:id
**설명**: 위험 알림 상태 변경
**인증**: Bearer Token 필수 (HR_ADMIN만)
**Request Body**:
```json
{
  "status": "string (required, enum: REVIEWED | DISMISSED)"
}
```

---

## 5. 폴더 구조 및 컴포넌트 분리 전략

```
anonvoice/
├── frontend/                    # Next.js 14 프론트엔드
│   ├── src/
│   │   ├── app/                 # App Router 페이지
│   │   │   ├── layout.tsx       # 루트 레이아웃 (인증 프로바이더 래핑)
│   │   │   ├── page.tsx         # 메인 페이지 (리다이렉트)
│   │   │   ├── login/
│   │   │   │   └── page.tsx     # SSO 로그인 페이지
│   │   │   ├── feedback/
│   │   │   │   ├── new/
│   │   │   │   │   └── page.tsx # 피드백 작성 페이지
│   │   │   │   └── my/
│   │   │   │       └── page.tsx # 내 피드백 목록
│   │   │   ├── dashboard/
│   │   │   │   ├── page.tsx     # 관리자 대시보드
│   │   │   │   └── layout.tsx   # 대시보드 레이아웃
│   │   │   └── hr/
│   │   │       ├── page.tsx     # HR 대시보드
│   │   │       ├── alerts/
│   │   │       │   └── page.tsx # 위험 알림 목록
│   │   │       └── reports/
│   │   │           └── page.tsx # 리포트 관리
│   │   ├── components/          # 재사용 UI 컴포넌트
│   │   │   ├── ui/              # shadcn/ui 기반 기본 컴포넌트
│   │   │   ├── feedback/        # 피드백 관련 컴포넌트
│   │   │   │   ├── FeedbackForm.tsx
│   │   │   │   ├── FeedbackCard.tsx
│   │   │   │   └── ConversationThread.tsx
│   │   │   ├── dashboard/       # 대시보드 관련 컴포넌트
│   │   │   │   ├── SentimentChart.tsx
│   │   │   │   ├── TrendGraph.tsx
│   │   │   │   └── WordCloud.tsx
│   │   │   ├── notification/    # 알림 컴포넌트
│   │   │   │   └── NotificationBell.tsx
│   │   │   └── layout/          # 레이아웃 컴포넌트
│   │   │       ├── Header.tsx
│   │   │       ├── Sidebar.tsx
│   │   │       └── Footer.tsx
│   │   ├── hooks/               # 커스텀 훅
│   │   │   ├── useFeedbacks.ts
│   │   │   ├── useNotifications.ts
│   │   │   ├── useStatistics.ts
│   │   │   └── useAuth.ts
│   │   ├── services/            # API 통신 레이어
│   │   │   ├── api.ts           # Axios/Fetch 인스턴스 (인터셉터 설정)
│   │   │   ├── feedbackService.ts
│   │   │   ├── notificationService.ts
│   │   │   ├── statisticsService.ts
│   │   │   └── authService.ts
│   │   ├── stores/              # Zustand 스토어
│   │   │   ├── authStore.ts
│   │   │   ├── notificationStore.ts
│   │   │   └── feedbackStore.ts
│   │   ├── types/               # TypeScript 타입 정의
│   │   │   ├── feedback.ts
│   │   │   ├── user.ts
│   │   │   ├── notification.ts
│   │   │   └── api.ts
│   │   ├── utils/               # 유틸리티 함수
│   │   │   ├── tokenStorage.ts  # 익명 토큰 로컬 저장 관리
│   │   │   ├── validation.ts    # 입력값 검증
│   │   │   └── formatting.ts    # 날짜/텍스트 포맷
│   │   └── lib/                 # 라이브러리 설정
│   │       └── auth.ts          # NextAuth 설정
│   ├── public/                  # 정적 파일
│   ├── .env.local               # 환경변수 (gitignore 대상)
│   ├── next.config.js
│   ├── tailwind.config.ts
│   └── tsconfig.json
│
├── backend/                     # Express 백엔드
│   ├── src/
│   │   ├── app.ts               # Express 앱 초기화
│   │   ├── server.ts            # 서버 엔트리포인트
│   │   ├── config/              # 설정 모듈
│   │   │   ├── database.ts      # DB 연결 설정
│   │   │   ├── env.ts           # 환경변수 검증 및 로드
│   │   │   └── redis.ts         # Redis 연결 설정
│   │   ├── middleware/          # Express 미들웨어
│   │   │   ├── auth.ts          # JWT 인증 미들웨어
│   │   │   ├── rbac.ts          # 역할 기반 접근 제어
│   │   │   ├── rateLimiter.ts   # Rate Limiting
│   │   │   ├── validator.ts     # 요청 검증 (Zod 기반)
│   │   │   ├── auditLogger.ts   # 감사 로그 자동 기록
│   │   │   ├── helmet.ts        # 보안 헤더 설정
│   │   │   └── errorHandler.ts  # 전역 에러 핸들러
│   │   ├── routes/              # 라우트 정의
│   │   │   ├── authRoutes.ts
│   │   │   ├── feedbackRoutes.ts
│   │   │   ├── conversationRoutes.ts
│   │   │   ├── notificationRoutes.ts
│   │   │   ├── statisticsRoutes.ts
│   │   │   ├── reportRoutes.ts
│   │   │   └── riskAlertRoutes.ts
│   │   ├── controllers/         # 요청/응답 처리
│   │   │   ├── authController.ts
│   │   │   ├── feedbackController.ts
│   │   │   ├── conversationController.ts
│   │   │   ├── notificationController.ts
│   │   │   ├── statisticsController.ts
│   │   │   ├── reportController.ts
│   │   │   └── riskAlertController.ts
│   │   ├── services/            # 비즈니스 로직
│   │   │   ├── feedbackService.ts
│   │   │   ├── conversationService.ts
│   │   │   ├── notificationService.ts
│   │   │   ├── statisticsService.ts
│   │   │   ├── reportService.ts
│   │   │   ├── keywordDetectionService.ts
│   │   │   └── encryptionService.ts
│   │   ├── models/              # Prisma 모델 (자동 생성)
│   │   ├── utils/               # 유틸리티
│   │   │   ├── anonymousToken.ts  # 익명 토큰 생성/검증
│   │   │   ├── crypto.ts          # AES-256-GCM 암복호화
│   │   │   └── logger.ts          # 로거 (민감 정보 마스킹)
│   │   └── types/               # 백엔드 타입 정의
│   │       ├── express.d.ts     # Express 타입 확장
│   │       └── index.ts
│   ├── prisma/
│   │   ├── schema.prisma        # Prisma 스키마
│   │   └── migrations/          # DB 마이그레이션
│   ├── .env                     # 환경변수 (gitignore 대상)
│   └── tsconfig.json
│
├── shared/                      # 프론트/백엔드 공유 타입
│   └── types/
│       ├── api.ts               # API 요청/응답 타입
│       └── enums.ts             # 공유 Enum 정의
│
├── .github/
│   └── workflows/
│       ├── ci.yml               # CI 파이프라인
│       └── deploy.yml           # CD 파이프라인
│
├── docker-compose.yml           # 로컬 개발 환경
├── .env.example                 # 환경변수 예시 (비밀값 미포함)
├── .gitignore
└── README.md
```

### 컴포넌트 분리 기준

- **프론트엔드**: 페이지 단위(App Router) > 도메인별 컴포넌트 > 공통 UI 컴포넌트 > 커스텀 훅 > 서비스 레이어
- **백엔드**: Route > Controller (요청/응답 매핑) > Service (비즈니스 로직) > Prisma Model (데이터 접근)
- **관심사 분리**: UI 로직과 비즈니스 로직을 명확히 분리. API 통신은 반드시 `services/` 레이어를 경유.

---

## 6. 상태 관리 및 전역 상태 흐름도

### 6.1 상태 분류 기준

| 구분 | 상태 관리 방식 | 예시 |
|---|---|---|
| 전역 상태 | Zustand | 인증 정보, 알림 목록, 알림 개수 |
| 서버 상태 | SWR 또는 React Query | 피드백 목록, 통계 데이터, 리포트 |
| 로컬 상태 | useState | 폼 입력값, 모달 열림/닫힘, 필터 선택 |
| 영속 상태 | localStorage (암호화) | 익명 토큰 (클라이언트만 보관) |

### 6.2 주요 상태 흐름도

```mermaid
stateDiagram-v2
    [*] --> 미인증

    미인증 --> SSO로그인중: SSO 로그인 버튼 클릭
    SSO로그인중 --> 인증완료: SSO 인증 성공
    SSO로그인중 --> 미인증: SSO 인증 실패

    인증완료 --> 일반직원화면: role = EMPLOYEE
    인증완료 --> 관리자화면: role = TEAM_LEADER
    인증완료 --> HR화면: role = HR_ADMIN

    state 일반직원화면 {
        [*] --> 피드백작성
        피드백작성 --> 제출완료: 피드백 제출
        제출완료 --> 토큰저장: 익명 토큰 localStorage 저장
        [*] --> 내피드백조회
        내피드백조회 --> 대화확인: 답변 존재시
    }

    state 관리자화면 {
        [*] --> 대시보드
        대시보드 --> 피드백상세: 피드백 클릭
        피드백상세 --> 답변작성: 답변 버튼 클릭
        답변작성 --> 대시보드: 답변 제출 완료
    }

    state HR화면 {
        [*] --> 전사통계
        전사통계 --> 위험알림확인: 위험 키워드 알림 클릭
        전사통계 --> 리포트생성: 리포트 생성 요청
    }

    인증완료 --> 미인증: 로그아웃
```

### 6.3 서버 상태 동기화 전략

- **피드백 목록**: SWR의 `revalidateOnFocus`로 탭 전환 시 자동 갱신. 5분 간격 폴링.
- **알림**: 30초 간격 폴링으로 읽지 않은 알림 카운트 갱신. 향후 WebSocket 전환 가능.
- **통계 데이터**: 캐시 TTL 5분. 대시보드 진입 시 stale-while-revalidate 패턴.
- **익명 토큰**: localStorage에 AES 암호화하여 저장. 피드백 ID를 키로 사용.

---

## 7. 보안 설계

### 7.1 OWASP Top 10 대응

| 위협 | 대응 방안 | 구현 위치 |
|---|---|---|
| A01: 접근 권한 취약 (Broken Access Control) | RBAC 미들웨어로 역할별 접근 제어. 팀 리더는 자기 팀만, HR은 전사 접근. 모든 접근에 서버측 권한 검증 | `middleware/rbac.ts` |
| A02: 암호화 실패 (Cryptographic Failures) | 피드백/대화 내용 AES-256-GCM 암호화. 암호화 키는 AWS KMS 관리. 전 구간 HTTPS 강제 | `utils/crypto.ts`, Nginx/ALB |
| A03: 인젝션 (Injection) | Prisma ORM의 Parameterized Query 강제. 사용자 입력 Zod 스키마 검증. CSP 헤더로 XSS 차단 | `middleware/validator.ts`, Prisma |
| A04: 불안전한 설계 (Insecure Design) | 피드백-작성자 매핑 정보를 서버에 저장하지 않는 아키텍처 수준의 익명성 보장 | 전체 아키텍처 |
| A05: 보안 설정 오류 (Security Misconfiguration) | Helmet.js로 보안 헤더 자동 설정. 프로덕션 환경에서 디버그 모드 비활성화. 불필요한 HTTP 메서드 차단 | `middleware/helmet.ts` |
| A06: 취약한 컴포넌트 (Vulnerable Components) | npm audit 자동 실행 (CI). Dependabot 알림 활성화. MIT/Apache 2.0 라이선스만 허용 | CI 파이프라인 |
| A07: 인증 실패 (Auth Failures) | JWT + Refresh Token 회전. Access Token 만료 1시간, Refresh Token 만료 7일. Refresh Token 1회 사용 후 폐기 | `middleware/auth.ts` |
| A08: 데이터 무결성 실패 (Software & Data Integrity) | CI/CD 파이프라인에서 의존성 무결성 검증. `npm ci --ignore-scripts` 후 수동 빌드 | CI 파이프라인 |
| A09: 보안 로깅 실패 (Security Logging) | 관리자의 모든 피드백 조회/답변을 감사 로그에 기록. 민감 정보 마스킹. 로그에 개인정보 노출 방지 | `middleware/auditLogger.ts` |
| A10: SSRF (Server-Side Request Forgery) | 외부 URL 입력 없음. SSO 콜백 URL 화이트리스트. 내부 네트워크 접근 차단 | `config/env.ts` |

### 7.2 추가 보안 대책

| 위협 | 대응 방안 | 구현 위치 |
|---|---|---|
| CSRF | SameSite=Strict 쿠키 + CSRF 토큰 이중 검증 | 미들웨어 |
| Rate Limiting | IP 기반 + 사용자 기반 이중 제한 (피드백 작성: 분당 5회, API 전체: 분당 100회) | `middleware/rateLimiter.ts` |
| 익명성 역추적 | 서버 로그에서 피드백 작성 API의 사용자 식별 정보를 제거. 타이밍 공격 방지를 위한 랜덤 지연 추가 | `feedbackController.ts` |
| 브루트포스 | 연속 5회 인증 실패 시 15분 잠금 | `middleware/rateLimiter.ts` |

### 7.3 인증/인가 흐름도

```mermaid
sequenceDiagram
    participant User as 사용자 브라우저
    participant Next as Next.js (프론트)
    participant API as Express API 서버
    participant SSO as 사내 SSO
    participant DB as PostgreSQL
    participant Redis as Redis (세션)

    Note over User,Redis: === SSO 로그인 흐름 ===
    User->>Next: SSO 로그인 버튼 클릭
    Next->>SSO: OAuth/SAML 인증 요청
    SSO-->>User: 사내 계정 인증 폼
    User->>SSO: 사내 계정 인증
    SSO-->>Next: 인증 코드 반환
    Next->>API: 인증 코드로 토큰 요청
    API->>SSO: 인증 코드 검증
    SSO-->>API: 사용자 정보 반환
    API->>DB: 사용자 정보 조회/생성
    API->>Redis: Refresh Token 저장
    API-->>Next: Access Token + Refresh Token
    Next-->>User: 로그인 완료, 대시보드로 이동

    Note over User,Redis: === 익명 피드백 작성 흐름 ===
    User->>Next: 피드백 작성 요청
    Next->>API: POST /feedbacks (Bearer Token)
    API->>API: JWT 검증 (유효한 직원인지만 확인)
    API->>API: 사용자 식별 정보 컨텍스트에서 즉시 제거
    API->>API: 익명 토큰 생성 (crypto.randomBytes)
    API->>DB: 피드백 저장 (작성자 ID 없음, 토큰 해시만 저장)
    API-->>Next: feedbackId + anonymousToken
    Next->>Next: anonymousToken을 localStorage에 암호화 저장
    Next-->>User: 제출 완료 안내

    Note over User,Redis: === 관리자 피드백 조회 흐름 ===
    User->>Next: 대시보드 접근
    Next->>API: GET /feedbacks (Bearer Token)
    API->>API: JWT 검증 + RBAC (TEAM_LEADER 이상)
    API->>API: 팀 소속 검증 (자기 팀만)
    API->>DB: 감사 로그 기록 (VIEW_FEEDBACK)
    API->>DB: 피드백 목록 조회 (복호화)
    API-->>Next: 피드백 목록 반환
```

### 7.4 익명성 보장 설계 (핵심 보안 아키텍처)

이 시스템의 가장 중요한 보안 요구사항은 **익명성 보장**이다. 다음 원칙을 준수한다:

1. **서버에 매핑 정보 없음**: `feedbacks` 테이블에 `author_id` 컬럼이 존재하지 않음. 어떤 조인으로도 작성자를 역추적할 수 없음.
2. **1회용 익명 토큰**: 피드백 작성 시 `crypto.randomBytes(32)`로 생성. 서버에는 `SHA-256` 해시만 저장.
3. **서버 로그 정화**: 피드백 작성 API 요청의 서버 로그에서 `userId`, `IP`, `User-Agent` 등 식별 가능한 정보를 모두 제거.
4. **타이밍 공격 방지**: 피드백 작성 API 응답에 50~200ms 랜덤 지연을 추가하여, 요청 시간 기반 역추적 방지.
5. **DB 관리자 접근 통제**: PostgreSQL Row-Level Security(RLS) 적용. DBA도 직접 쿼리로 익명성을 우회할 수 없도록 설계.

### 7.5 민감 데이터 처리 정책

| 데이터 | 처리 방식 | 저장 위치 |
|---|---|---|
| 피드백 본문 | AES-256-GCM 암호화 | PostgreSQL (feedbacks.content) |
| 대화 내용 | AES-256-GCM 암호화 | PostgreSQL (conversations.content) |
| 익명 토큰 | SHA-256 해시 (서버), 평문 (클라이언트 localStorage) | DB: 해시 / 브라우저: localStorage |
| 사용자 이메일 | 평문 저장 (사내 이메일) | PostgreSQL (users.email) |
| 암호화 키 | AWS KMS 관리 | AWS KMS |
| JWT Secret | 환경변수 | .env (배포시 AWS Secrets Manager) |
| 리포트 PDF | S3 서버측 암호화 (SSE-S3) | AWS S3 |

### 7.6 환경변수 관리 전략

```bash
# .env.example (비밀값 미포함, 저장소에 커밋)
DATABASE_URL=postgresql://user:password@localhost:5432/anonvoice
REDIS_URL=redis://localhost:6379
JWT_SECRET=<환경별_비밀키>
JWT_REFRESH_SECRET=<환경별_비밀키>
ENCRYPTION_KEY=<AWS_KMS_키_ARN>
SSO_CLIENT_ID=<SSO_클라이언트_ID>
SSO_CLIENT_SECRET=<SSO_클라이언트_시크릿>
SMTP_HOST=<SMTP_호스트>
SMTP_PORT=587
SMTP_USER=<SMTP_사용자>
SMTP_PASS=<SMTP_비밀번호>
SENTRY_DSN=<Sentry_DSN>
DATADOG_API_KEY=<Datadog_API_키>
S3_BUCKET_NAME=anonvoice-reports
AWS_REGION=ap-northeast-2
```

- 모든 비밀값은 소스코드에 하드코딩 절대 금지
- 로컬 개발: `.env.local` (gitignore 대상)
- 스테이징/프로덕션: AWS Secrets Manager 또는 ECS Task Definition의 환경변수
- CI/CD: GitHub Actions Secrets

### 7.7 의존성 패키지 라이선스 검토

| 패키지 | 라이선스 | 허용 여부 |
|---|---|---|
| React / Next.js | MIT | 허용 |
| Zustand | MIT | 허용 |
| Tailwind CSS | MIT | 허용 |
| shadcn/ui | MIT | 허용 |
| Express | MIT | 허용 |
| Prisma | Apache 2.0 | 허용 |
| NextAuth.js | ISC | 허용 (MIT 호환) |
| Helmet.js | MIT | 허용 |
| Zod | MIT | 허용 |
| Sentry SDK | MIT | 허용 |
| SWR | MIT | 허용 |
| jsonwebtoken | MIT | 허용 |

- GPL, LGPL, AGPL 등 소스 공개 의무 라이선스는 사용 금지
- 신규 패키지 추가 시 `license-checker` 또는 `license-report`로 자동 검증

---

## 8. 예외 처리 및 에러 전략

### 8.1 에러 분류 체계

| 분류 | 코드 범위 | 설명 | 예시 |
|---|---|---|---|
| 비즈니스 에러 | 400-499 | 사용자 요청의 문제 | 입력값 오류, 권한 없음, 리소스 없음 |
| 시스템 에러 | 500-599 | 서버 내부 문제 | DB 연결 실패, 외부 서비스 장애 |

### 8.2 에러 응답 표준 포맷 (RFC 7807)

```json
{
  "type": "https://anonvoice.internal/errors/invalid-input",
  "title": "Invalid Input",
  "status": 400,
  "detail": "피드백 본문은 10자 이상이어야 합니다.",
  "instance": "/api/v1/feedbacks",
  "errors": [
    {
      "field": "content",
      "message": "최소 10자 이상 입력해주세요.",
      "code": "MIN_LENGTH"
    }
  ],
  "traceId": "uuid (Sentry/Datadog 연동용)"
}
```

### 8.3 에러 코드 정의

| 에러 타입 | HTTP 코드 | 설명 |
|---|---|---|
| INVALID_INPUT | 400 | 입력값 검증 실패 |
| INVALID_CATEGORY | 400 | 유효하지 않은 카테고리 값 |
| INVALID_TEAM | 400 | 존재하지 않는 팀 ID |
| CONVERSATION_LIMIT_REACHED | 400 | 대화 왕복 횟수 초과 |
| UNAUTHORIZED | 401 | 인증 토큰 누락/만료 |
| SSO_AUTH_FAILED | 401 | SSO 인증 실패 |
| INVALID_ANONYMOUS_TOKEN | 403 | 익명 토큰 불일치 |
| FORBIDDEN | 403 | 접근 권한 없음 (역할 부족) |
| TEAM_ACCESS_DENIED | 403 | 해당 팀 데이터 접근 불가 |
| FEEDBACK_NOT_FOUND | 404 | 존재하지 않는 피드백 |
| REPORT_NOT_FOUND | 404 | 존재하지 않는 리포트 |
| DUPLICATE_RESOURCE | 409 | 중복 리소스 |
| RATE_LIMIT_EXCEEDED | 429 | 요청 제한 초과 |
| INTERNAL_ERROR | 500 | 서버 내부 오류 |
| SSO_UNAVAILABLE | 503 | SSO 서버 연결 불가 |
| DATABASE_ERROR | 503 | 데이터베이스 연결 오류 |

### 8.4 로깅 전략

| 로그 레벨 | 용도 | 예시 |
|---|---|---|
| ERROR | 시스템 장애, 예외 발생 | DB 연결 실패, 암호화 키 로드 실패 |
| WARN | 비정상이지만 처리 가능한 상황 | Rate limit 근접, 위험 키워드 감지 |
| INFO | 주요 비즈니스 이벤트 | 피드백 생성 (작성자 정보 제외), 리포트 생성 |
| DEBUG | 개발/디버깅 전용 | 요청/응답 상세 (프로덕션 비활성화) |

**민감 정보 마스킹 규칙**:
- 로그에 사용자 이메일 노출 금지 (필요시 `a***@company.com` 형태)
- JWT 토큰 전체를 로그에 기록하지 않음 (앞 8자만 기록)
- 피드백 본문은 로그에 절대 기록하지 않음
- IP 주소는 감사 로그에만 기록, 일반 로그에는 제외

### 8.5 재시도 정책

| API/작업 | 멱등성 | 재시도 전략 |
|---|---|---|
| GET 요청 전체 | 멱등 | 최대 3회, 지수 백오프 (1s, 2s, 4s) |
| POST /feedbacks | 비멱등 | 재시도 금지 (중복 피드백 방지) |
| POST /conversations | 비멱등 | 재시도 금지 (sequence_number 유니크 제약) |
| PATCH /notifications/:id/read | 멱등 | 최대 2회 재시도 |
| POST /reports/generate | 비멱등 | 재시도 금지 (클라이언트에서 상태 폴링) |
| 이메일 발송 | 멱등 (발송 자체) | 최대 3회, 지수 백오프. 실패 시 큐에 재등록 |

---

## 9. 3줄 요약 및 비유

> **3줄 요약**
> 1. 직원은 완전한 익명성이 보장된 환경에서 팀이나 조직에 솔직한 피드백을 작성하고, 관리자와 익명 대화를 나눌 수 있습니다.
> 2. 관리자는 대시보드에서 피드백 현황, 감정 트렌드, 키워드 분석을 한눈에 확인하고, HR은 전사 통계와 위험 신호를 실시간으로 모니터링합니다.
> 3. 서버에 작성자 정보를 저장하지 않는 아키텍처 설계와 다층 암호화로, 기술적으로도 익명성을 역추적할 수 없게 보장합니다.
>
> **비유로 이해하기**
> 이 시스템은 **익명 투표함**과 비슷합니다. 직원(투표자)이 피드백(투표지)을 작성하면,
> 입구에서 신분증(SSO 인증)을 확인해 "자격이 있는 직원인지"만 검증한 뒤, 신분증 정보를 즉시 잊어버립니다.
> 투표지는 잠금장치(AES-256 암호화)가 걸린 투표함(DB)에 들어가고, 투표지에는 이름 대신
> 1회용 번호표(익명 토큰)만 적혀 있어서, 나중에 그 번호표로 답변 확인만 할 수 있습니다.
> 관리자는 투표함을 열어 내용을 읽을 수 있지만, 누가 넣었는지는 절대 알 수 없습니다.
> HR은 전체 투표함들의 경향(통계)을 분석하고, "퇴사"나 "번아웃" 같은 위험 신호가 감지되면 즉시 알림을 받습니다.
