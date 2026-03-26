# 상세 설계 문서 (SDD): AnonVoice - 사내 익명 피드백 시스템

| 항목 | 내용 |
|---|---|
| 문서 버전 | 1.0 |
| 작성일 | 2026-03-26 |
| 기반 PRD | test-prd-2.md |
| 상태 | Draft |

---

## 1. 시스템 아키텍처 개요

### 1.1 전체 구성도

```
[사용자 브라우저]
    │  HTTPS (TLS 1.3)
    ▼
[Vercel CDN / Next.js SSR]  ──── NextAuth.js ──── [사내 SSO IdP]
    │
    │  HTTPS (내부 통신)
    ▼
[AWS ALB (WAF 적용)]
    │
    ▼
[AWS ECS - Express API 서버] ──── [Redis (세션/캐시)]
    │
    ▼
[PostgreSQL 16 (RDS, 암호화)]
    │
    ▼
[S3 (리포트 PDF 저장, SSE-S3)]
```

### 1.2 주요 설계 원칙

- **익명성 최우선**: 피드백 작성자와 피드백 간 매핑 정보를 서버에 절대 저장하지 않음
- **제로 트러스트**: 모든 외부 입력은 서버에서 철저히 검증
- **최소 권한**: 각 역할은 필요한 데이터에만 접근 가능
- **방어적 프로그래밍**: 모든 계층에서 입력 검증 및 예외 처리

---

## 2. 데이터 모델 설계

### 2.1 ERD

```
┌──────────────┐     ┌──────────────────┐     ┌─────────────────┐
│    User       │     │    Feedback       │     │    Comment       │
├──────────────┤     ├──────────────────┤     ├─────────────────┤
│ id (PK)       │     │ id (PK)           │     │ id (PK)          │
│ employeeId    │     │ targetTeamId (FK) │     │ feedbackId (FK)  │
│ name          │     │ category          │     │ anonymousToken   │
│ email         │     │ content           │     │ content          │
│ role          │     │ sentimentTag      │     │ authorType       │
│ teamId (FK)   │     │ anonymousToken    │     │ createdAt        │
│ createdAt     │     │ createdAt         │     └─────────────────┘
│ updatedAt     │     │ expiresAt         │
└──────────────┘     └──────────────────┘
                           │
        ┌──────────────┐   │   ┌──────────────────┐
        │    Team       │◄──┘  │   AuditLog        │
        ├──────────────┤       ├──────────────────┤
        │ id (PK)       │       │ id (PK)           │
        │ name          │       │ userId (FK)       │
        │ department    │       │ action            │
        │ managerId(FK) │       │ resourceType      │
        │ createdAt     │       │ resourceId        │
        └──────────────┘       │ ipAddress         │
                               │ userAgent         │
        ┌──────────────┐       │ createdAt         │
        │ Notification  │       └──────────────────┘
        ├──────────────┤
        │ id (PK)       │       ┌──────────────────┐
        │ userId (FK)   │       │ AlertKeyword      │
        │ type          │       ├──────────────────┤
        │ content       │       │ id (PK)           │
        │ isRead        │       │ keyword           │
        │ metadata      │       │ severity          │
        │ createdAt     │       │ isActive          │
        └──────────────┘       └──────────────────┘
```

### 2.2 Prisma 스키마 핵심 정의

```prisma
model User {
  id         String   @id @default(uuid())
  employeeId String   @unique
  name       String
  email      String   @unique
  role       Role     @default(EMPLOYEE)
  teamId     String
  team       Team     @relation(fields: [teamId], references: [id])
  createdAt  DateTime @default(now())
  updatedAt  DateTime @updatedAt

  auditLogs     AuditLog[]
  notifications Notification[]
}

enum Role {
  EMPLOYEE
  TEAM_LEADER
  HR_ADMIN
}

model Feedback {
  id             String       @id @default(uuid())
  targetTeamId   String
  targetTeam     Team         @relation(fields: [targetTeamId], references: [id])
  category       Category
  content        String       @db.VarChar(1000)
  sentimentTag   Sentiment
  anonymousToken String       @unique
  createdAt      DateTime     @default(now())
  expiresAt      DateTime     // createdAt + 2년

  comments Comment[]

  @@index([targetTeamId, createdAt])
  @@index([category])
  @@index([sentimentTag])
}

enum Category {
  CULTURE
  PROCESS
  LEADERSHIP
  OTHER
}

enum Sentiment {
  POSITIVE
  NEUTRAL
  NEGATIVE
}

model Comment {
  id             String     @id @default(uuid())
  feedbackId     String
  feedback       Feedback   @relation(fields: [feedbackId], references: [id])
  anonymousToken String?    // 작성자가 코멘트할 때만 사용
  content        String     @db.VarChar(500)
  authorType     AuthorType
  createdAt      DateTime   @default(now())

  @@index([feedbackId, createdAt])
}

enum AuthorType {
  ANONYMOUS_AUTHOR  // 피드백 원작성자
  MANAGER           // 팀 리더/관리자
}

model AuditLog {
  id           String   @id @default(uuid())
  userId       String
  user         User     @relation(fields: [userId], references: [id])
  action       String   // VIEW_FEEDBACK, REPLY_FEEDBACK 등
  resourceType String
  resourceId   String
  ipAddress    String
  userAgent    String
  createdAt    DateTime @default(now())

  @@index([userId, createdAt])
  @@index([resourceType, resourceId])
}
```

### 2.3 핵심 설계 결정: 익명성 보장

**Feedback 테이블에 `userId` 컬럼이 없다.** 이것이 익명성 보장의 핵심이다.

- 피드백 생성 시 `anonymousToken`을 서버에서 생성하여 클라이언트에 반환
- 서버는 이 토큰과 사용자 ID 간의 매핑을 **어디에도 저장하지 않음**
- 클라이언트(브라우저)의 localStorage에만 토큰 저장
- 후속 대화 시 이 토큰으로 본인 확인

---

## 3. API 설계

### 3.1 인증 관련

| Method | Path | 설명 | 인증 |
|--------|------|------|------|
| GET | `/api/auth/signin` | SSO 로그인 리다이렉트 | 불필요 |
| GET | `/api/auth/callback` | SSO 콜백 처리 | 불필요 |
| POST | `/api/auth/signout` | 로그아웃 | 필요 |

### 3.2 피드백 API

| Method | Path | 설명 | 권한 |
|--------|------|------|------|
| POST | `/api/feedbacks` | 피드백 작성 | EMPLOYEE 이상 |
| GET | `/api/feedbacks/mine` | 내 피드백 목록 (토큰 기반) | EMPLOYEE 이상 |
| GET | `/api/feedbacks/team/:teamId` | 팀 피드백 목록 | TEAM_LEADER (본인 팀) |
| GET | `/api/feedbacks/all` | 전사 피드백 통계 | HR_ADMIN |

### 3.3 댓글/대화 API

| Method | Path | 설명 | 권한 |
|--------|------|------|------|
| POST | `/api/feedbacks/:id/comments` | 댓글 작성 | TEAM_LEADER 또는 토큰 보유자 |
| GET | `/api/feedbacks/:id/comments` | 댓글 조회 | TEAM_LEADER (본인 팀) 또는 토큰 보유자 |

### 3.4 통계/리포트 API

| Method | Path | 설명 | 권한 |
|--------|------|------|------|
| GET | `/api/stats/team/:teamId` | 팀 통계 | TEAM_LEADER (본인 팀) |
| GET | `/api/stats/organization` | 전사 통계 | HR_ADMIN |
| GET | `/api/reports/quarterly` | 분기별 리포트 다운로드 | HR_ADMIN |
| GET | `/api/stats/team/:teamId/trend` | 월별 트렌드 | TEAM_LEADER (본인 팀) |

### 3.5 알림 API

| Method | Path | 설명 | 권한 |
|--------|------|------|------|
| GET | `/api/notifications` | 내 알림 목록 | 로그인 사용자 |
| PATCH | `/api/notifications/:id/read` | 읽음 처리 | 로그인 사용자 (본인) |

### 3.6 API 요청/응답 예시

**POST /api/feedbacks** (피드백 작성)

요청:
```json
{
  "targetTeamId": "uuid-team-abc",
  "category": "LEADERSHIP",
  "content": "최근 팀 미팅이 더 효율적으로 변한 것 같습니다.",
  "sentimentTag": "POSITIVE"
}
```

응답 (201 Created):
```json
{
  "feedbackId": "uuid-feedback-xyz",
  "anonymousToken": "anon_tk_a1b2c3d4e5f6...",
  "message": "피드백이 성공적으로 등록되었습니다."
}
```

**주의**: `anonymousToken`은 이 응답에서 **한 번만** 반환된다. 클라이언트는 이를 안전하게 localStorage에 저장해야 한다.

### 3.7 에러 응답 형식 (RFC 7807)

```json
{
  "type": "https://anonvoice.internal/errors/validation",
  "title": "입력값 검증 실패",
  "status": 400,
  "detail": "피드백 본문은 10자 이상이어야 합니다.",
  "instance": "/api/feedbacks"
}
```

---

## 4. 보안 설계 (상세)

### 4.1 위협 모델링 (STRIDE)

| 위협 유형 | 시나리오 | 대응책 |
|-----------|----------|--------|
| **Spoofing** | 사내 SSO 토큰 위조 | NextAuth.js의 JWT 서명 검증 + SSO IdP 연동, 토큰 만료 시간 단축 (15분) |
| **Tampering** | 피드백 내용 변조 | DB 레코드 immutable 설계 (수정/삭제 불가), 감사 로그 별도 테이블 |
| **Repudiation** | 관리자가 피드백 조회 사실 부인 | AuditLog 테이블에 모든 관리자 행위 기록 (IP, User-Agent 포함) |
| **Information Disclosure** | 익명 작성자 신원 노출 | userId-feedback 매핑 미저장, 서버 로그에서 사용자 식별 정보 제거 |
| **Denial of Service** | 대량 피드백 스팸 | Rate Limiting (사용자당 1일 10건), WAF 적용 |
| **Elevation of Privilege** | 일반 직원이 관리자 API 접근 | RBAC 미들웨어, 팀 소속 검증 이중 체크 |

### 4.2 익명성 보장 아키텍처

```
[작성자 브라우저]                      [서버]                    [DB]
     │                                  │                        │
     │  POST /api/feedbacks             │                        │
     │  (JWT: userId 포함)              │                        │
     │ ─────────────────────────────►   │                        │
     │                                  │ 1. JWT에서 userId 추출  │
     │                                  │ 2. userId로 팀 소속 검증│
     │                                  │ 3. anonymousToken 생성  │
     │                                  │    (crypto.randomBytes) │
     │                                  │ 4. userId 폐기 ◄────── │ 핵심!
     │                                  │ 5. 피드백 저장           │
     │                                  │    (userId 없이!)       │
     │                                  │ ─────────────────────► │
     │                                  │                        │ INSERT feedback
     │  응답: { anonymousToken }        │                        │ (no userId column)
     │ ◄───────────────────────────── │                        │
     │                                  │                        │
     │  localStorage에 토큰 저장        │                        │
     │                                  │                        │
```

**핵심 보안 조치:**

1. **서버 메모리에서 userId 즉시 폐기**: 피드백 저장 트랜잭션 완료 후 userId 변수를 null로 설정
2. **요청 로그에서 userId 제거**: 피드백 작성 API의 접근 로그에는 userId를 기록하지 않음 (별도 로그 필터 적용)
3. **anonymousToken 생성**: `crypto.randomBytes(32)`로 256비트 무작위 토큰 생성 (KISA 권장 수준)
4. **네트워크 레벨 보호**: 피드백 작성 시점의 IP 주소도 서버 로그에서 마스킹 처리

### 4.3 인증 및 인가

#### 4.3.1 인증 흐름

```
[브라우저] → [Next.js] → [NextAuth.js] → [사내 SSO IdP (OIDC)]
                                              │
                                              ▼
                                    [JWT 발급 (서명: RS256)]
                                              │
                                              ▼
                              [httpOnly, Secure, SameSite=Strict 쿠키]
```

- **JWT 알고리즘**: RS256 (비대칭 서명, 공개키로만 검증 가능)
- **Access Token 유효기간**: 15분
- **Refresh Token 유효기간**: 7일, httpOnly 쿠키 저장
- **세션 저장소**: Redis (서버 사이드 세션 무효화 지원)

#### 4.3.2 RBAC (Role-Based Access Control)

```typescript
// 미들웨어 체인 구조
const authMiddleware = [
  verifyJWT,           // 1. JWT 서명 및 만료 검증
  extractUser,         // 2. 사용자 정보 추출
  checkRole(allowedRoles), // 3. 역할 확인
  checkTeamAccess,     // 4. 팀 소속 검증 (관리자 API)
  auditLog,            // 5. 감사 로그 기록
];
```

**역할별 접근 매트릭스:**

| 리소스 | EMPLOYEE | TEAM_LEADER | HR_ADMIN |
|--------|----------|-------------|----------|
| 피드백 작성 | O | O | O |
| 내 피드백 조회 | O | O | O |
| 팀 피드백 조회 | X | 본인 팀만 | 전체 |
| 피드백 답변 | X | 본인 팀만 | X |
| 팀 통계 | X | 본인 팀만 | 전체 |
| 전사 통계 | X | X | O |
| 위험 알림 관리 | X | X | O |
| 리포트 다운로드 | X | X | O |

### 4.4 입력 검증 (Zero Trust Input)

모든 입력은 서버에서 재검증한다. 프론트엔드 검증은 UX 개선 목적일 뿐, 보안 경계로 취급하지 않는다.

```typescript
// 피드백 작성 입력 검증 스키마 (Zod 사용)
const createFeedbackSchema = z.object({
  targetTeamId: z.string().uuid("유효한 팀 ID가 아닙니다"),
  category: z.enum(["CULTURE", "PROCESS", "LEADERSHIP", "OTHER"]),
  content: z
    .string()
    .min(10, "피드백은 최소 10자 이상이어야 합니다")
    .max(1000, "피드백은 최대 1000자까지 가능합니다")
    .refine(
      (val) => !containsSQLInjection(val),
      "허용되지 않는 문자가 포함되어 있습니다"
    ),
  sentimentTag: z.enum(["POSITIVE", "NEUTRAL", "NEGATIVE"]),
});
```

**추가 검증 항목:**

- XSS 방지: 모든 사용자 입력에 DOMPurify 적용 (출력 시점)
- SQL Injection 방지: Prisma ORM의 파라미터 바인딩 강제 사용
- 경로 조작 방지: teamId, feedbackId 등 UUID 형식만 허용
- Content-Type 검증: `application/json`만 허용

### 4.5 Rate Limiting 정책

| API 그룹 | 제한 | 윈도우 | 대상 |
|-----------|------|--------|------|
| 피드백 작성 | 10회 | 24시간 | 사용자별 |
| 댓글 작성 | 20회 | 24시간 | 사용자별 |
| 대시보드 조회 | 100회 | 1시간 | 사용자별 |
| 인증 시도 | 5회 | 15분 | IP별 |
| 전체 API | 1000회 | 1시간 | IP별 |

구현: Express Rate Limiter + Redis 슬라이딩 윈도우 방식

### 4.6 데이터 암호화

| 구분 | 방식 | 상세 |
|------|------|------|
| 전송 중 (In Transit) | TLS 1.3 | 모든 HTTP 통신 HTTPS 강제, HSTS 헤더 적용 |
| 저장 중 (At Rest) | AES-256 | RDS 스토리지 암호화 (AWS KMS 관리 키) |
| 피드백 본문 | AES-256-GCM | 애플리케이션 레벨 암호화 (키는 AWS KMS에서 관리) |
| anonymousToken | SHA-256 해시 | DB에는 해시값만 저장, 원본은 클라이언트에만 존재 |
| 비밀번호 | 해당 없음 | SSO 연동이므로 자체 비밀번호 저장 없음 |

**anonymousToken 보안 처리 흐름:**
1. 서버에서 `crypto.randomBytes(32)`로 원본 토큰 생성
2. `SHA-256(token)`을 DB에 저장
3. 원본 토큰을 클라이언트에 반환
4. 클라이언트가 요청 시 원본 토큰 전송 → 서버에서 해시 후 DB 조회

### 4.7 감사 로그 설계

```typescript
interface AuditLogEntry {
  userId: string;         // 행위자
  action: AuditAction;    // 수행 행위
  resourceType: string;   // 대상 리소스 종류
  resourceId: string;     // 대상 리소스 ID
  ipAddress: string;      // 접속 IP (X-Forwarded-For 검증)
  userAgent: string;      // 브라우저 정보
  timestamp: Date;        // 발생 시각
}

enum AuditAction {
  VIEW_FEEDBACK = "VIEW_FEEDBACK",
  REPLY_FEEDBACK = "REPLY_FEEDBACK",
  VIEW_DASHBOARD = "VIEW_DASHBOARD",
  DOWNLOAD_REPORT = "DOWNLOAD_REPORT",
  VIEW_ALERT = "VIEW_ALERT",
}
```

**감사 로그 정책:**
- 관리자(TEAM_LEADER, HR_ADMIN)의 모든 조회/변경 행위 기록
- 일반 직원의 피드백 작성은 **기록하지 않음** (익명성 보장)
- 감사 로그는 별도 DB 인스턴스 또는 별도 스키마에 저장 (분리 원칙)
- 보관 기간: 5년 (내부 감사 기준)

### 4.8 보안 헤더

```typescript
// Express 보안 헤더 (helmet 사용)
app.use(helmet({
  contentSecurityPolicy: {
    directives: {
      defaultSrc: ["'self'"],
      scriptSrc: ["'self'"],
      styleSrc: ["'self'", "'unsafe-inline'"],  // Tailwind 인라인 스타일
      imgSrc: ["'self'", "data:"],
      connectSrc: ["'self'"],
      frameSrc: ["'none'"],
      objectSrc: ["'none'"],
    },
  },
  hsts: { maxAge: 31536000, includeSubDomains: true, preload: true },
  referrerPolicy: { policy: "strict-origin-when-cross-origin" },
  crossOriginEmbedderPolicy: true,
  crossOriginOpenerPolicy: { policy: "same-origin" },
}));
```

### 4.9 OWASP Top 10 대응 매핑

| OWASP 2021 | 대응 방안 |
|-------------|-----------|
| A01: Broken Access Control | RBAC 미들웨어 + 팀 소속 이중 검증 + 감사 로그 |
| A02: Cryptographic Failures | TLS 1.3, AES-256-GCM, SHA-256 토큰 해시, AWS KMS 키 관리 |
| A03: Injection | Prisma 파라미터 바인딩, Zod 입력 검증, DOMPurify XSS 방지 |
| A04: Insecure Design | STRIDE 위협 모델링, 익명성 보장 아키텍처 |
| A05: Security Misconfiguration | Helmet 보안 헤더, 환경변수 기반 설정, 불필요 API 비활성화 |
| A06: Vulnerable Components | npm audit 자동화, Dependabot 알림, 허용 라이선스 검증 |
| A07: Auth Failures | SSO 연동, JWT RS256, Rate Limiting, 세션 만료 관리 |
| A08: Data Integrity Failures | 피드백 immutable 설계, JWT 서명 검증 |
| A09: Logging Failures | 구조화된 감사 로그, Sentry/Datadog 모니터링 |
| A10: SSRF | 외부 요청 없음 (내부 시스템), allowlist 기반 URL 제한 |

---

## 5. 주요 컴포넌트 설계

### 5.1 백엔드 레이어 구조

```
src/
├── middlewares/
│   ├── auth.middleware.ts        # JWT 검증
│   ├── rbac.middleware.ts        # 역할 기반 접근 제어
│   ├── teamAccess.middleware.ts  # 팀 소속 검증
│   ├── rateLimiter.middleware.ts # 요청 제한
│   ├── auditLog.middleware.ts    # 감사 로그
│   ├── inputSanitizer.middleware.ts # 입력 정제
│   └── errorHandler.middleware.ts   # 에러 처리 (RFC 7807)
├── routes/
│   ├── feedback.routes.ts
│   ├── comment.routes.ts
│   ├── stats.routes.ts
│   ├── notification.routes.ts
│   └── report.routes.ts
├── services/
│   ├── feedback.service.ts
│   ├── comment.service.ts
│   ├── stats.service.ts
│   ├── notification.service.ts
│   ├── report.service.ts
│   ├── keyword.service.ts       # 위험 키워드 감지
│   └── encryption.service.ts    # 암호화 유틸리티
├── validators/
│   ├── feedback.validator.ts
│   └── comment.validator.ts
├── repositories/
│   ├── feedback.repository.ts
│   ├── comment.repository.ts
│   └── auditLog.repository.ts
└── config/
    ├── database.ts
    ├── redis.ts
    └── security.ts
```

### 5.2 피드백 작성 서비스 (핵심 로직)

```typescript
class FeedbackService {
  async createFeedback(
    userId: string,     // 인증 미들웨어에서 추출
    dto: CreateFeedbackDto
  ): Promise<CreateFeedbackResponse> {
    // 1. 대상 팀 존재 여부 검증
    const team = await this.teamRepository.findById(dto.targetTeamId);
    if (!team) throw new NotFoundError("존재하지 않는 팀입니다");

    // 2. anonymousToken 생성 (256비트 암호학적 난수)
    const rawToken = crypto.randomBytes(32).toString("hex");
    const hashedToken = crypto
      .createHash("sha256")
      .update(rawToken)
      .digest("hex");

    // 3. 피드백 본문 암호화
    const encryptedContent = await this.encryptionService.encrypt(dto.content);

    // 4. 피드백 저장 (userId는 저장하지 않음)
    const feedback = await this.feedbackRepository.create({
      targetTeamId: dto.targetTeamId,
      category: dto.category,
      content: encryptedContent,
      sentimentTag: dto.sentimentTag,
      anonymousToken: hashedToken,
      expiresAt: addYears(new Date(), 2),
    });

    // 5. 위험 키워드 감지 (비동기)
    this.keywordService.detectAndAlert(dto.content, feedback.id);

    // 6. 관리자 알림 발송 (비동기)
    this.notificationService.notifyTeamManager(dto.targetTeamId, feedback.id);

    // 7. userId 참조 제거 (방어적 프로그래밍)
    // 이 시점부터 userId는 사용되지 않음

    return {
      feedbackId: feedback.id,
      anonymousToken: rawToken,  // 원본 토큰 (최초 1회만 반환)
      message: "피드백이 성공적으로 등록되었습니다.",
    };
  }
}
```

### 5.3 위험 키워드 감지 서비스

```typescript
class KeywordService {
  // 감지 대상 키워드 (DB에서 관리, 캐시 적용)
  private keywords: AlertKeyword[];

  async detectAndAlert(content: string, feedbackId: string): Promise<void> {
    const normalizedContent = content.toLowerCase().replace(/\s+/g, " ");

    const matchedKeywords = this.keywords.filter(
      (kw) => kw.isActive && normalizedContent.includes(kw.keyword)
    );

    if (matchedKeywords.length > 0) {
      const maxSeverity = Math.max(...matchedKeywords.map((k) => k.severity));

      // HR 관리자에게 즉시 알림
      await this.notificationService.alertHRAdmins({
        type: "RISK_KEYWORD_DETECTED",
        feedbackId,
        keywords: matchedKeywords.map((k) => k.keyword),
        severity: maxSeverity,
      });

      // 심각도 높은 경우 이메일도 발송
      if (maxSeverity >= 3) {
        await this.emailService.sendUrgentAlert(feedbackId, matchedKeywords);
      }
    }
  }
}
```

### 5.4 익명 대화 시스템

```typescript
class CommentService {
  async createComment(
    feedbackId: string,
    dto: CreateCommentDto,
    requester: { userId?: string; anonymousToken?: string }
  ): Promise<Comment> {
    const feedback = await this.feedbackRepository.findById(feedbackId);
    if (!feedback) throw new NotFoundError("피드백을 찾을 수 없습니다");

    // 대화 횟수 제한 확인 (최대 5회 왕복 = 10개 코멘트)
    const commentCount = await this.commentRepository.countByFeedbackId(feedbackId);
    if (commentCount >= 10) {
      throw new BadRequestError("대화 횟수 제한(5회 왕복)에 도달했습니다");
    }

    let authorType: AuthorType;

    if (requester.anonymousToken) {
      // 익명 작성자: 토큰 해시 비교로 본인 확인
      const hashedToken = hashToken(requester.anonymousToken);
      if (hashedToken !== feedback.anonymousToken) {
        throw new ForbiddenError("권한이 없습니다");
      }
      authorType = AuthorType.ANONYMOUS_AUTHOR;
    } else if (requester.userId) {
      // 관리자: 역할 및 팀 소속 검증 (미들웨어에서 처리됨)
      authorType = AuthorType.MANAGER;
    } else {
      throw new UnauthorizedError("인증이 필요합니다");
    }

    return this.commentRepository.create({
      feedbackId,
      content: await this.encryptionService.encrypt(dto.content),
      authorType,
      anonymousToken: requester.anonymousToken
        ? hashToken(requester.anonymousToken)
        : null,
    });
  }
}
```

---

## 6. 프론트엔드 설계

### 6.1 페이지 구조 (Next.js App Router)

```
app/
├── layout.tsx                    # 루트 레이아웃
├── page.tsx                      # 랜딩/로그인
├── (auth)/
│   └── login/page.tsx            # SSO 로그인 페이지
├── (protected)/
│   ├── layout.tsx                # 인증 필요 레이아웃
│   ├── feedback/
│   │   ├── new/page.tsx          # 피드백 작성
│   │   └── mine/page.tsx         # 내 피드백 목록
│   ├── dashboard/
│   │   └── page.tsx              # 팀 리더 대시보드
│   └── hr/
│       ├── page.tsx              # HR 대시보드
│       └── reports/page.tsx      # 리포트 관리
└── api/
    └── auth/[...nextauth]/route.ts
```

### 6.2 Zustand 스토어 구조

```typescript
// 익명 토큰 관리 스토어 (persist 미들웨어 사용)
interface AnonymousTokenStore {
  tokens: Record<string, string>;  // feedbackId → anonymousToken
  addToken: (feedbackId: string, token: string) => void;
  getToken: (feedbackId: string) => string | undefined;
  removeToken: (feedbackId: string) => void;
}

// 알림 스토어
interface NotificationStore {
  notifications: Notification[];
  unreadCount: number;
  fetchNotifications: () => Promise<void>;
  markAsRead: (id: string) => Promise<void>;
}
```

### 6.3 클라이언트 측 익명 토큰 보안

- **저장 위치**: localStorage (IndexedDB 대안 검토 가능)
- **위험 요소**: XSS 공격 시 토큰 탈취 가능
- **대응:**
  - CSP 헤더로 인라인 스크립트 차단
  - DOMPurify로 모든 렌더링 데이터 정제
  - 토큰에 만료 시간 포함 (2년, 피드백 보관 기간과 동일)
  - 토큰 사용 시 Fingerprinting (브라우저 특성 해시) 추가 검증 고려

---

## 7. 데이터 생명주기 및 보관 정책

### 7.1 데이터 보관 주기

| 데이터 | 보관 기간 | 삭제 방식 |
|--------|-----------|-----------|
| 피드백 | 2년 | 스케줄러 기반 자동 삭제 (매일 새벽 3시) |
| 댓글 | 피드백과 동일 | CASCADE 삭제 |
| 감사 로그 | 5년 | 별도 아카이빙 후 삭제 |
| 알림 | 90일 | 스케줄러 기반 자동 삭제 |
| 리포트 PDF | 3년 | S3 Lifecycle 정책 |

### 7.2 삭제 스케줄러

```typescript
// 매일 03:00 KST 실행 (cron: 0 18 * * * UTC)
async function cleanupExpiredFeedbacks(): Promise<void> {
  const expiredFeedbacks = await prisma.feedback.findMany({
    where: { expiresAt: { lte: new Date() } },
    select: { id: true },
  });

  // 배치 삭제 (1000건 단위)
  for (const batch of chunk(expiredFeedbacks, 1000)) {
    await prisma.feedback.deleteMany({
      where: { id: { in: batch.map((f) => f.id) } },
    });
  }
}
```

---

## 8. 인프라 및 배포 설계

### 8.1 배포 아키텍처

```
[Route 53 DNS]
      │
      ▼
[CloudFront + WAF]
      │
      ├── /api/* ──► [ALB] ──► [ECS Fargate (Express API)]
      │                              │
      │                              ├── [RDS PostgreSQL (Multi-AZ)]
      │                              ├── [ElastiCache Redis]
      │                              └── [S3 (리포트 저장)]
      │
      └── /* ──► [Vercel (Next.js)]
```

### 8.2 환경변수 관리

모든 시크릿은 AWS Secrets Manager 또는 Parameter Store에서 관리한다. 소스코드에 하드코딩 금지.

```
# 필수 환경변수 목록
DATABASE_URL            # PostgreSQL 연결 문자열
REDIS_URL               # Redis 연결 문자열
NEXTAUTH_SECRET         # NextAuth 세션 암호화 키
SSO_CLIENT_ID           # SSO OIDC Client ID
SSO_CLIENT_SECRET       # SSO OIDC Client Secret
ENCRYPTION_KEY_ARN      # AWS KMS 키 ARN
SENTRY_DSN              # Sentry 에러 추적
S3_REPORT_BUCKET        # 리포트 저장 S3 버킷
SMTP_HOST               # 이메일 발송 서버
SMTP_USER               # 이메일 계정
SMTP_PASS               # 이메일 비밀번호
```

### 8.3 모니터링 및 알림

| 도구 | 용도 | 알림 채널 |
|------|------|-----------|
| Sentry | 애플리케이션 에러 추적 | Slack #ops-alerts |
| Datadog | 인프라 메트릭, APM | PagerDuty (P1), Slack (P2+) |
| CloudWatch | AWS 리소스 모니터링 | SNS → Slack |
| Custom | 위험 키워드 감지 | 이메일 + 인앱 알림 |

**주요 모니터링 지표:**
- API 응답 시간 p99 < 500ms
- 에러율 < 0.1%
- DB 커넥션 풀 사용률 < 80%
- Rate Limit 초과 횟수 (비정상 사용 패턴 감지)

---

## 9. 성능 고려사항

### 9.1 캐싱 전략

| 대상 | 캐시 위치 | TTL | 무효화 조건 |
|------|-----------|-----|-------------|
| 팀 목록 | Redis | 1시간 | 팀 정보 변경 시 |
| 통계 데이터 | Redis | 5분 | 새 피드백 등록 시 |
| 위험 키워드 목록 | 인메모리 + Redis | 10분 | HR 관리자 수정 시 |
| 사용자 역할/팀 정보 | Redis | 15분 | SSO 재인증 시 |

### 9.2 DB 인덱스 전략

```sql
-- 팀별 피드백 조회 최적화
CREATE INDEX idx_feedback_team_created ON feedback(target_team_id, created_at DESC);

-- 카테고리 필터링
CREATE INDEX idx_feedback_category ON feedback(category);

-- 감정 태그 통계
CREATE INDEX idx_feedback_sentiment ON feedback(sentiment_tag);

-- 댓글 조회
CREATE INDEX idx_comment_feedback ON comment(feedback_id, created_at);

-- 감사 로그 조회
CREATE INDEX idx_audit_user_time ON audit_log(user_id, created_at DESC);

-- 만료 피드백 삭제 최적화
CREATE INDEX idx_feedback_expires ON feedback(expires_at) WHERE expires_at IS NOT NULL;
```

---

## 10. 테스트 전략

| 테스트 유형 | 범위 | 도구 |
|-------------|------|------|
| 단위 테스트 | 서비스 로직, 검증 로직 | Jest |
| 통합 테스트 | API 엔드포인트, DB 연동 | Supertest + Testcontainers |
| E2E 테스트 | 주요 사용자 시나리오 | Playwright |
| 보안 테스트 | OWASP ZAP 스캔, 의존성 취약점 | ZAP, npm audit |
| 부하 테스트 | API 성능, 동시접속 | k6 |

**보안 관련 필수 테스트 케이스:**
- 피드백 생성 시 DB에 userId가 저장되지 않는지 검증
- 관리자가 타 팀 피드백에 접근할 때 403 반환 검증
- anonymousToken 없이 익명 대화 시도 시 거부 검증
- Rate Limit 초과 시 429 반환 검증
- XSS 페이로드가 sanitize 되는지 검증
- SQL Injection 패턴이 차단되는지 검증

---

## 11. 라이선스 검토

| 패키지 | 라이선스 | 허용 여부 |
|--------|----------|-----------|
| React / Next.js | MIT | 허용 |
| Express | MIT | 허용 |
| Prisma | Apache 2.0 | 허용 |
| Zustand | MIT | 허용 |
| Tailwind CSS | MIT | 허용 |
| shadcn/ui | MIT | 허용 |
| NextAuth.js | ISC | 허용 |
| helmet | MIT | 허용 |
| Zod | MIT | 허용 |
| DOMPurify | Apache 2.0 / MIT | 허용 |
| ioredis | MIT | 허용 |

GPL류 라이선스 패키지는 사용하지 않는다.

---

## 12. 리스크 및 고려사항

| 리스크 | 영향도 | 발생 가능성 | 대응 방안 |
|--------|--------|-------------|-----------|
| 익명 토큰 탈취 (XSS) | 높음 | 중간 | CSP 강화, DOMPurify 적용, 토큰 만료 관리 |
| 관리자 권한 남용 | 중간 | 낮음 | 감사 로그 기록, 정기 감사 리포트 |
| 피드백 스팸 | 중간 | 중간 | Rate Limiting, 최소 글자 수 제한, 이상 패턴 감지 |
| 서버 로그에서 작성자 추적 | 높음 | 낮음 | 피드백 API 로그에서 userId 마스킹, 로그 접근 제한 |
| 소규모 팀 역추적 | 높음 | 높음 | 최소 팀 인원 기준 설정 (5인 이하 팀은 부서 단위로 통합), 시간 지연 노출 |

---

## 부록: 용어 정의

| 용어 | 설명 |
|------|------|
| anonymousToken | 피드백 작성자 식별용 1회성 토큰 (클라이언트 보관) |
| RBAC | Role-Based Access Control, 역할 기반 접근 제어 |
| STRIDE | Spoofing, Tampering, Repudiation, Info Disclosure, DoS, Elevation of Privilege |
| RFC 7807 | HTTP API 에러 응답 표준 형식 (Problem Details for HTTP APIs) |
| CSP | Content Security Policy, 콘텐츠 보안 정책 |
