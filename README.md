# claude-simple-e2e
Simple end‑to‑end development using Claude Code



최종 프로젝트 구조 요약

  ┌─────────────┬────────────────────┬───────────────────────────────────────────────────────────────────────────────────────┬────────────────────┐
  │  프로젝트   │        언어        │                                       주요 파일                                       │        상태        │
  ├─────────────┼────────────────────┼───────────────────────────────────────────────────────────────────────────────────────┼────────────────────┤
  │ kcp-cli     │ Go                 │ SDK 5개 클라이언트, CLI 8개 명령어, TUI 프레임워크, 설정 관리                         │ 빌드 + 테스트 통과 │
  ├─────────────┼────────────────────┼───────────────────────────────────────────────────────────────────────────────────────┼────────────────────┤
  │ kcp-gateway │ Go + Gin           │ 인증 API, 7개 리소스 핸들러, 3개 미들웨어, Repository/Service 레이어, DB 마이그레이션 │ 빌드 통과          │
  ├─────────────┼────────────────────┼───────────────────────────────────────────────────────────────────────────────────────┼────────────────────┤
  │ kcp-webui   │ Next.js + Tailwind │ 로그인, 대시보드, 6개 리소스 관리 화면, 감사 로그, 공통 DataTable/StatCard            │ 패키지 설치 완료   │
  ├─────────────┼────────────────────┼───────────────────────────────────────────────────────────────────────────────────────┼────────────────────┤
  │ db          │ SQL/Shell          │ PostgreSQL nerdctl 실행 스크립트, 초기 스키마                                         │ 생성 완료          │
  └─────────────┴────────────────────┴───────────────────────────────────────────────────────────────────────────────────────┴────────────────────┘

  실행/테스트 방법

  # Go 프로젝트 빌드
  make build

  # 테스트 실행
  make test

  # PostgreSQL 컨테이너 실행
  chmod +x db/nerdctl-postgres.sh
  make db-up

  # Gateway 실행
  make run-gateway

  # WebUI 개발 서버
  make run-webui

 


---

 로그인 방법

  사전 준비:
  # 1. PostgreSQL 컨테이너 실행
  chmod +x db/nerdctl-postgres.sh
  make db-up

  # 2. Gateway 실행 (마이그레이션 + 초기 관리자 계정 자동 생성)
  make run-gateway

  Gateway가 시작되면 자동으로:
  - 테이블 생성 (users, sessions, audit_logs)
  - 초기 관리자 계정 생성: admin / admin123

  CLI 로그인:
  # 빌드
  make build

  # 로그인
  ./bin/kcp login
  # → Server URL: http://localhost:8080/api/v1
  # → Username: admin
  # → Password: admin123

  WebUI 로그인:
  http://localhost:3000/login
  Username: admin
  Password: admin123

  관리자 계정 커스터마이징 (환경변수):
  KCP_ADMIN_USER=myadmin KCP_ADMIN_PASSWORD=mypassword make run-gateway


---

1. 설정 파일 형식: JSON → YAML
  - 기본 경로: ~/.kcp/kcp-config.yaml
  - 아규먼트: kcp --config /other/path/config.yaml login
  - 환경변수: KCP_CONFIG=/other/path/config.yaml kcp login

  2. 설정 파일 자동 생성: 파일이 없으면 기본값으로 반환
  # ~/.kcp/kcp-config.yaml (자동 생성)
  server_url: http://localhost:8080/api/v1
  token: ""
  auth_type: JWT
  username: ""

  3. 로그인 시 서버 URL 입력 불필요: config에서 자동으로 읽음
  # 기본 config 사용
  ./bin/kcp login
  # 서버 URL: http://localhost:8080/api/v1   ← 자동 표시
  # 사용자명: admin
  # 비밀번호: admin123

  # 다른 config 파일 지정
  ./bin/kcp --config /path/to/other-config.yaml login


---
   3. 사용법:
  # 기본: 현재 디렉토리의 kcp-gateway-config.yaml 로드
  make run-gateway

  # 경로 지정
  make run-gateway CONFIG=/path/to/config.yaml