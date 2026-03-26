# 시스템 설계 문서 (SDD): TaskFlow - 실시간 협업 할일 관리

## 1. 개요

소규모 팀(2~10명)을 위한 실시간 협업 칸반보드 웹앱의 시스템 아키텍처를 정의한다.
실시간 동기화, 드래그앤드롭 칸반보드, 팀 관리, 알림 기능을 핵심으로 한다.

---

## 2. 추천 기술 스택

### 2.1 프론트엔드

| 기술 | 버전 | 선정 사유 |
|------|------|-----------|
| **React** | 19.x | 컴포넌트 기반 UI, 풍부한 생태계, 드래그앤드롭 라이브러리 지원 |
| **TypeScript** | 5.x | 타입 안전성으로 협업 시 버그 감소 |
| **Vite** | 6.x | 빠른 개발 서버 및 빌드 |
| **Tailwind CSS** | 4.x | 유틸리티 기반 스타일링, 빠른 UI 개발 |
| **@dnd-kit/core** | 6.x | 접근성 우수한 드래그앤드롭 라이브러리 (MIT) |
| **Zustand** | 5.x | 경량 상태 관리, 실시간 데이터와 궁합 우수 (MIT) |
| **React Router** | 7.x | 클라이언트 라우팅 (MIT) |

### 2.2 백엔드

| 기술 | 버전 | 선정 사유 |
|------|------|-----------|
| **Node.js** | 22 LTS | 이벤트 기반 비동기 I/O, WebSocket과 자연스러운 통합 |
| **Fastify** | 5.x | 고성능 HTTP 프레임워크, 스키마 기반 검증 내장 (MIT) |
| **Socket.IO** | 4.x | 실시간 양방향 통신, 자동 재연결/폴백 지원 (MIT) |
| **Prisma** | 6.x | 타입 안전 ORM, 마이그레이션 관리 용이 (Apache 2.0) |
| **TypeScript** | 5.x | 프론트엔드와 타입 공유 가능 |

### 2.3 데이터베이스 및 인프라

| 기술 | 선정 사유 |
|------|-----------|
| **PostgreSQL 16** | ACID 트랜잭션, JSON 지원, 무료 (PostgreSQL License) |
| **Redis 7** | 실시간 Pub/Sub, 세션 캐시, 알림 큐 (BSD-3-Clause) |
| **MinIO** | S3 호환 오브젝트 스토리지, 이미지 첨부파일 저장 (AGPL-3.0 - 서버 사이드 독립 배포이므로 소스 공개 의무 해당 없음, 대안으로 로컬 파일시스템도 가능) |
| **Docker / Docker Compose** | 로컬 개발 및 배포 환경 통일 (Apache 2.0) |

### 2.4 라이선스 검토 요약

모든 핵심 의존성은 MIT, Apache 2.0, BSD-3-Clause, PostgreSQL License로 GPL 소스 공개 의무 없음을 확인하였다. MinIO(AGPL-3.0)는 독립 서비스로 배포되므로 자체 코드에 소스 공개 의무가 전파되지 않으나, 우려 시 로컬 파일시스템 + Nginx 정적 파일 서빙으로 대체 가능하다.

---

## 3. 시스템 아키텍처

### 3.1 전체 구조

```
┌─────────────────────────────────────────────────┐
│                   클라이언트                      │
│  React + TypeScript + Tailwind + @dnd-kit       │
│  Zustand (로컬 상태) + Socket.IO Client          │
└──────────┬──────────────────┬────────────────────┘
           │ HTTPS/REST       │ WSS (Socket.IO)
           ▼                  ▼
┌──────────────────────────────────────────────────┐
│              API Gateway / Fastify               │
│  ┌──────────┐  ┌──────────┐  ┌───────────────┐  │
│  │ REST API │  │ WS 핸들러 │  │ 인증 미들웨어  │  │
│  └────┬─────┘  └────┬─────┘  └───────────────┘  │
│       │              │                            │
│  ┌────┴──────────────┴─────┐                     │
│  │     서비스 레이어         │                     │
│  │  Task / Board / Team    │                     │
│  │  Notification / Auth    │                     │
│  └────┬──────────────┬─────┘                     │
└───────┼──────────────┼───────────────────────────┘
        │              │
   ┌────▼────┐    ┌────▼────┐    ┌──────────┐
   │PostgreSQL│    │  Redis  │    │  MinIO   │
   │ (데이터) │    │(Pub/Sub │    │ (이미지) │
   │         │    │ + 캐시) │    │          │
   └─────────┘    └─────────┘    └──────────┘
```

### 3.2 계층 구조 (Layered Architecture)

```
Controller (Route Handler)
    │
    ▼
Service Layer (비즈니스 로직)
    │
    ▼
Repository Layer (Prisma ORM)
    │
    ▼
Database (PostgreSQL)
```

각 계층은 단방향 의존성을 가지며, 서비스 레이어에서 비즈니스 규칙을 집중 관리한다.

---

## 4. 데이터 모델

### 4.1 ERD

```
┌──────────┐     ┌──────────────┐     ┌──────────┐
│   User   │────<│  TeamMember  │>────│   Team   │
│──────────│     │──────────────│     │──────────│
│ id (PK)  │     │ userId (FK)  │     │ id (PK)  │
│ email    │     │ teamId (FK)  │     │ name     │
│ name     │     │ role (enum)  │     │ createdAt│
│ password │     │ joinedAt     │     └────┬─────┘
│ createdAt│     └──────────────┘          │
└────┬─────┘                               │
     │                              ┌──────┴─────┐
     │                              │   Board    │
     │                              │────────────│
     │                              │ id (PK)    │
     │                              │ teamId(FK) │
     │                              │ name       │
     │                              └──────┬─────┘
     │                                     │
     │                              ┌──────┴─────┐
     │         ┌──────────┐         │   Column   │
     │         │TaskLabel │         │────────────│
     │         │──────────│         │ id (PK)    │
     │         │taskId(FK)│         │ boardId(FK)│
     │         │labelId(FK│         │ name       │
     │         └──────────┘         │ position   │
     │                              └──────┬─────┘
     │                                     │
     │                              ┌──────┴─────┐
     └──────────────────────────────│   Task     │
                                    │────────────│
                                    │ id (PK)    │
                                    │ columnId(FK│
                                    │ title      │
                                    │ description│
                                    │ priority   │
                                    │ assigneeId │
                                    │ dueDate    │
                                    │ position   │
                                    │ version    │ ← 낙관적 잠금용
                                    │ createdAt  │
                                    │ updatedAt  │
                                    └──────┬─────┘
                                           │
                    ┌──────────┐     ┌──────┴─────┐
                    │  Label   │     │  Comment   │
                    │──────────│     │────────────│
                    │ id (PK)  │     │ id (PK)    │
                    │ boardId  │     │ taskId(FK) │
                    │ name     │     │ authorId   │
                    │ color    │     │ content    │
                    └──────────┘     │ createdAt  │
                                     └────────────┘

┌──────────────┐
│ Notification │
│──────────────│
│ id (PK)      │
│ userId (FK)  │
│ type (enum)  │
│ payload(JSON)│
│ isRead       │
│ createdAt    │
└──────────────┘

┌──────────────┐
│  Attachment  │
│──────────────│
│ id (PK)      │
│ taskId (FK)  │
│ fileName     │
│ fileKey      │ ← 오브젝트 스토리지 키
│ fileSize     │
│ mimeType     │
│ uploadedBy   │
│ createdAt    │
└──────────────┘
```

### 4.2 주요 Enum 정의

```typescript
enum Priority {
  LOW = 'LOW',
  MEDIUM = 'MEDIUM',
  HIGH = 'HIGH',
  URGENT = 'URGENT'
}

enum TeamRole {
  ADMIN = 'ADMIN',
  MEMBER = 'MEMBER'
}

enum NotificationType {
  TASK_ASSIGNED = 'TASK_ASSIGNED',
  DUE_DATE_APPROACHING = 'DUE_DATE_APPROACHING',
  DUE_DATE_TODAY = 'DUE_DATE_TODAY',
  COMMENT_ADDED = 'COMMENT_ADDED'
}
```

---

## 5. API 설계

### 5.1 인증

| Method | Endpoint | 설명 |
|--------|----------|------|
| POST | `/api/auth/register` | 회원가입 |
| POST | `/api/auth/login` | 로그인 (JWT 발급) |
| POST | `/api/auth/refresh` | 토큰 갱신 |

### 5.2 팀 관리

| Method | Endpoint | 설명 |
|--------|----------|------|
| POST | `/api/teams` | 팀 생성 |
| GET | `/api/teams/:teamId` | 팀 정보 조회 |
| POST | `/api/teams/:teamId/invite` | 멤버 초대 |
| DELETE | `/api/teams/:teamId/members/:userId` | 멤버 제거 |
| PATCH | `/api/teams/:teamId/members/:userId/role` | 역할 변경 |

### 5.3 보드 / 컬럼

| Method | Endpoint | 설명 |
|--------|----------|------|
| GET | `/api/teams/:teamId/boards` | 보드 목록 |
| POST | `/api/teams/:teamId/boards` | 보드 생성 |
| GET | `/api/boards/:boardId` | 보드 상세 (컬럼+태스크 포함) |
| POST | `/api/boards/:boardId/columns` | 컬럼 추가 |
| PATCH | `/api/columns/:columnId` | 컬럼 수정 |
| DELETE | `/api/columns/:columnId` | 컬럼 삭제 |
| PATCH | `/api/boards/:boardId/columns/reorder` | 컬럼 순서 변경 |

### 5.4 태스크

| Method | Endpoint | 설명 |
|--------|----------|------|
| POST | `/api/columns/:columnId/tasks` | 태스크 생성 |
| PATCH | `/api/tasks/:taskId` | 태스크 수정 |
| DELETE | `/api/tasks/:taskId` | 태스크 삭제 |
| PATCH | `/api/tasks/:taskId/move` | 태스크 이동 (컬럼 변경 + 순서) |
| GET | `/api/boards/:boardId/tasks?filter=...` | 필터링 조회 |

### 5.5 댓글 / 첨부파일

| Method | Endpoint | 설명 |
|--------|----------|------|
| GET | `/api/tasks/:taskId/comments` | 댓글 목록 |
| POST | `/api/tasks/:taskId/comments` | 댓글 작성 |
| POST | `/api/tasks/:taskId/attachments` | 이미지 업로드 |
| DELETE | `/api/attachments/:attachmentId` | 첨부파일 삭제 |

### 5.6 알림

| Method | Endpoint | 설명 |
|--------|----------|------|
| GET | `/api/notifications` | 내 알림 목록 |
| PATCH | `/api/notifications/:id/read` | 읽음 처리 |
| PATCH | `/api/notifications/read-all` | 전체 읽음 |

---

## 6. 실시간 동기화 설계

### 6.1 Socket.IO 이벤트 설계

```
[클라이언트 → 서버]
- board:join        { boardId }          // 보드 룸 참여
- board:leave       { boardId }          // 보드 룸 퇴장
- task:create       { columnId, data }
- task:update       { taskId, data, version }
- task:move         { taskId, toColumnId, position, version }
- task:delete       { taskId }
- column:create     { boardId, name }
- column:update     { columnId, name }
- column:reorder    { boardId, columnIds[] }

[서버 → 클라이언트]
- board:updated     { type, payload }    // 보드 변경 브로드캐스트
- task:conflict     { taskId, serverVersion, serverData }
- user:online       { userId, boardId }
- user:offline      { userId, boardId }
- notification:new  { notification }
```

### 6.2 낙관적 잠금 (Optimistic Locking)

```
1. 클라이언트가 task:update { taskId, data, version: 3 } 전송
2. 서버가 DB의 현재 version 확인
   - version == 3 → 업데이트 수행, version을 4로 증가, 브로드캐스트
   - version != 3 → task:conflict 이벤트로 최신 데이터 반환
3. 충돌 시 클라이언트가 최신 데이터를 표시하고 사용자에게 재시도 유도
```

### 6.3 Redis Pub/Sub 활용

다중 서버 인스턴스 환경을 고려하여 Socket.IO의 Redis Adapter를 사용한다.

```
Server A ──publish──> Redis Channel ──subscribe──> Server B
                                    ──subscribe──> Server C
```

---

## 7. 인증 및 보안 설계

### 7.1 인증 방식

- **JWT (Access Token + Refresh Token)**
  - Access Token: 15분 만료, 메모리 저장
  - Refresh Token: 7일 만료, HttpOnly Secure Cookie
  - Refresh Token Rotation 적용 (사용 시 새 토큰 발급, 기존 폐기)

### 7.2 보안 고려사항

| 항목 | 대응 |
|------|------|
| 비밀번호 저장 | bcrypt (cost factor 12) |
| 입력 검증 | Fastify JSON Schema 기반 서버 검증 + zod 공유 스키마 |
| SQL Injection | Prisma ORM 파라미터 바인딩 강제 |
| XSS | React 자동 이스케이프 + DOMPurify로 사용자 입력 sanitize |
| CSRF | SameSite=Strict Cookie + Origin 헤더 검증 |
| 파일 업로드 | MIME 타입 서버 검증 (image/* only), 5MB 제한, 파일명 UUID 치환 |
| Rate Limiting | fastify-rate-limit (로그인: 5회/분, API: 100회/분) |
| CORS | 허용 Origin 화이트리스트 |
| 환경 변수 | 모든 시크릿(DB URL, JWT Secret, Redis URL)은 환경변수로 관리, .env.example만 커밋 |
| 에러 응답 | RFC 7807 형식, 내부 에러 상세는 로그에만 기록 |

### 7.3 권한 모델

```
ADMIN: 팀 설정 변경, 멤버 초대/제거, 역할 변경, 보드/컬럼 관리, 태스크 전체 관리
MEMBER: 태스크 CRUD (본인 담당), 댓글 작성, 보드 조회
```

미들웨어에서 요청마다 팀 멤버십과 역할을 검증한다.

---

## 8. 알림 시스템 설계

### 8.1 알림 흐름

```
이벤트 발생 (태스크 할당, 댓글 작성 등)
    │
    ▼
서비스 레이어에서 Notification 레코드 생성 (PostgreSQL)
    │
    ▼
Redis Pub/Sub로 실시간 알림 전송
    │
    ▼
Socket.IO를 통해 대상 사용자에게 즉시 전달
(오프라인 시 → 다음 접속 시 미읽은 알림 목록에서 확인)
```

### 8.2 마감일 알림 스케줄러

- **node-cron**으로 매일 09:00에 실행
- D-1, D-day 태스크를 조회하여 담당자에게 알림 생성
- 향후 이메일 알림 확장 가능하도록 알림 채널을 추상화

---

## 9. 프론트엔드 구조

### 9.1 디렉토리 구조

```
src/
├── app/                    # 라우팅, 레이아웃
│   ├── routes/
│   └── layout.tsx
├── features/               # 기능 단위 모듈
│   ├── auth/
│   ├── board/
│   │   ├── components/     # KanbanBoard, Column, TaskCard
│   │   ├── hooks/          # useBoard, useDragAndDrop
│   │   └── api/
│   ├── team/
│   ├── task/
│   └── notification/
├── shared/                 # 공용 컴포넌트, 유틸, 타입
│   ├── components/
│   ├── hooks/
│   ├── lib/                # socket.ts, api-client.ts
│   └── types/
└── main.tsx
```

### 9.2 상태 관리 전략

| 상태 종류 | 관리 방법 |
|-----------|-----------|
| 서버 상태 (태스크, 보드) | Socket.IO 이벤트로 Zustand 스토어 직접 갱신 |
| UI 상태 (모달, 필터) | Zustand 또는 컴포넌트 로컬 상태 |
| 폼 상태 | React Hook Form + zod 검증 |
| URL 상태 (필터, 페이지) | React Router searchParams |

### 9.3 실시간 동기화 UX

- 다른 사용자의 변경 시 해당 카드에 부드러운 애니메이션 적용
- 충돌 발생 시 토스트 알림으로 "다른 사용자가 수정했습니다. 최신 내용을 확인하세요." 표시
- 온라인 사용자 아바타를 보드 상단에 표시

---

## 10. 백엔드 디렉토리 구조

```
src/
├── app.ts                  # Fastify 인스턴스 생성, 플러그인 등록
├── server.ts               # 서버 시작점
├── config/                 # 환경 변수 로딩, 검증
├── plugins/                # Fastify 플러그인 (auth, cors, rate-limit)
├── modules/
│   ├── auth/
│   │   ├── auth.controller.ts
│   │   ├── auth.service.ts
│   │   ├── auth.schema.ts   # 요청/응답 JSON Schema
│   │   └── auth.routes.ts
│   ├── team/
│   ├── board/
│   ├── task/
│   ├── comment/
│   ├── notification/
│   └── attachment/
├── socket/
│   ├── index.ts            # Socket.IO 초기화
│   ├── board.handler.ts    # 보드 관련 이벤트 핸들러
│   └── middleware.ts       # 소켓 인증 미들웨어
├── jobs/
│   └── due-date-notifier.ts # 마감일 알림 크론잡
├── shared/
│   ├── errors/             # 커스텀 에러 클래스
│   ├── middleware/          # 공통 미들웨어
│   └── utils/
└── prisma/
    ├── schema.prisma
    └── migrations/
```

---

## 11. 배포 구성

### 11.1 Docker Compose (개발/스테이징)

```yaml
services:
  app:
    build: .
    ports: ["3000:3000"]
    env_file: .env
    depends_on: [postgres, redis]

  client:
    build: ./client
    ports: ["5173:80"]

  postgres:
    image: postgres:16-alpine
    volumes: [pg_data:/var/lib/postgresql/data]

  redis:
    image: redis:7-alpine

  minio:
    image: minio/minio
    command: server /data
    ports: ["9000:9000"]
```

### 11.2 프로덕션 고려사항

- Nginx 리버스 프록시 (SSL 종단, 정적 파일 캐싱)
- PM2 또는 Docker 컨테이너로 Node.js 프로세스 관리
- PostgreSQL 일일 백업 (pg_dump)
- 사내 서비스이므로 클라우드 대신 온프레미스 서버도 적합

---

## 12. 제약 조건 대응

| 제약 조건 | 구현 방식 |
|-----------|-----------|
| 팀 최대 10명 | TeamMember 생성 시 카운트 검증 (서비스 레이어) |
| 보드당 태스크 500개 | Task 생성 시 해당 보드의 태스크 수 검증 |
| 이미지만 5MB | Multer 미들웨어에서 MIME 타입 + 파일 크기 검증, 매직바이트 추가 검증 |
| 무료 서비스 | 모든 인프라를 오픈소스/셀프 호스팅으로 구성 |

---

## 13. 비기능 요구사항

| 항목 | 목표 |
|------|------|
| 응답 시간 | REST API p95 < 200ms |
| 실시간 지연 | WebSocket 이벤트 전파 < 100ms (동일 서버) |
| 가용성 | 사내 서비스 기준 99.5% |
| 동시 접속 | 최소 50명 (10팀 x 5명 동시 접속 가정) |
| 데이터 백업 | 일일 1회 자동 백업 |
