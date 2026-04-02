# 프로그래밍 규칙

이 프로젝트에서 코드를 작성할 때 반드시 따라야 하는 규칙이다.
사용자의 요청에 의해 확립된 규칙이므로 예외 없이 준수한다.

## 아키텍처 원칙

### Thin Client
- CLI와 WebUI는 thin client로 유지한다 — 비즈니스 로직을 최소화한다
- 데이터 조합/enrichment(ID→이름 변환, 여러 API 결과 병합 등)는 반드시 Gateway에서 처리한다
- CLI/WebUI는 Gateway 응답을 그대로 출력/렌더링하는 역할만 담당한다
- Gateway가 반환하는 JSON에 클라이언트가 필요한 모든 정보가 포함되어야 한다

### 통신 규칙
- CLI ↔ Gateway, WebUI ↔ Gateway 간 통신은 반드시 JSON 포맷을 사용한다
- 요청 헤더에 `Content-Type: application/json`과 `Accept: application/json`을 포함한다
- OpenStack API의 HTML 에러 응답은 Gateway에서 통일된 JSON 에러 포맷으로 변환한다
- 비정상 응답(HTML, 빈 값 등)이 클라이언트에 그대로 전달되지 않도록 Gateway에서 방어한다

## 설정 파일 규칙

### 형식
- 설정 파일은 YAML 형식을 사용한다 (.env 형식 사용 금지)
- Gateway 설정: `kcp-gateway-config.yaml` (기본: 현재 디렉토리)
- CLI 설정: `kcp-config.yaml` (기본: `~/.kcp/kcp-config.yaml`)
- 설정 파일 경로는 `--config` 플래그 또는 환경변수로 변경 가능해야 한다

### 우선순위
- 환경변수 > YAML 파일 값 > 기본값 순서로 적용한다
- 설정 파일이 없으면 환경변수만으로 실행할 수 있어야 한다

### OpenStack 연동
- openrc 환경변수명(OS_AUTH_URL, OS_USERNAME 등)을 그대로 지원한다
- YAML에서도 openrc 필드명과 호환되는 키를 사용한다 (auth_url, username, project_name 등)
- Keystone Auth URL에 `/v3`가 포함되지 않으면 자동으로 추가한다

### 보안
- 설정 파일의 토큰은 파일 권한 600으로 보호한다
- 실제 설정 파일(kcp-gateway-config.yaml, kcp-config.yaml)은 .gitignore에 포함한다
- `.example` 파일만 Git에 커밋한다

## 인증 규칙

### CLI 로그인
- username은 설정 파일에 저장하지 않는다 — 로그인 시 항상 화면에서 입력받는다
- 비밀번호 입력은 화면에 표시하지 않는다 (term.ReadPassword 사용)
- 로그인 시 서버 URL은 설정 파일에서 읽어 표시만 한다 (입력 프롬프트 없음)

### Gateway 시작
- OpenStack 인증 실패 시에도 Gateway가 시작되어야 한다 (지연 인증)
- API 호출 시점에 자동 재인증을 시도한다
- 초기 관리자 계정이 없으면 시작 시 자동 생성한다

## CLI 출력 규칙

### 테이블 형식
- OpenStack CLI와 동일한 컬럼/형식으로 출력한다
- 한국어 등 전각 문자의 터미널 표시 폭을 올바르게 계산한다 (go-runewidth 사용)
- 테이블이 깨지지 않도록 바이트 수가 아닌 표시 폭 기준으로 정렬한다

### OpenStack CLI 호환 컬럼
- server list: ID, Name, Status, Networks, Image, Flavor
- flavor list: ID, Name, RAM, Disk, VCPUs, Is Public
- project list: ID, Name, Domain ID, Description, Enabled
- user list: ID, Name, Domain ID, Enabled
- network list: ID, Name, Subnets, Status, Shared, External
- subnet list: ID, Name, Network ID, CIDR, IP Version, Gateway IP, DHCP
- router list: ID, Name, Status, Admin State Up, HA, Project ID
- security group list: ID, Name, Rules, Project ID
- image list: ID, Name, Status, Disk Format, Size

## 구현 완결성 규칙

- 함수/메서드를 생성할 때 반드시 실제 동작하는 로직까지 구현한다
- `// TODO`, `501 NOT_IMPLEMENTED`, 빈 함수 본문 등 스텁(stub) 상태로 남기지 않는다
- 코드를 수정한 후에는 반드시 `make build`로 바이너리를 재빌드한다
- 외부 서비스 연동 시 실제 API를 호출하는 코드를 작성한다 (목 응답 금지)

## 매직 넘버/하드코딩 금지

- 타임아웃, 만료 시간, 재시도 횟수 등 동작에 영향을 주는 값을 코드에 직접 하드코딩하지 않는다
- 설정 파일(YAML)에서 읽을 수 있는 값은 반드시 설정을 통해 주입한다
- 기본값(fallback)은 `const`로 선언하여 명명된 상수를 사용한다
- 테스트 코드에서도 프로덕션 상수를 참조한다 (테스트에 매직 넘버를 직접 쓰지 않는다)

## TDD 필수 규칙 (NON-NEGOTIABLE)

- 프롬프트, skill, agent 등 어떤 방식으로든 프로그래밍을 수행할 때 반드시 tdd-cycle skill의 TDD 사이클을 따른다
- "빠르게 해줘", "자동으로 진행해줘", "끝까지 진행해줘", "프로그램을 작성해줘", "코드를 작성해줘", "코드를 변경해줘", "구현해줘", "수정해줘", "고쳐줘", "fix 해줘" 등 프로덕션 코드가 변경되는 어떤 요청이든 TDD 사이클을 절대 생략하지 않는다
- 테스트 파일이 없는 패키지가 발견되면 테스트를 작성한다
- 프로덕션 코드 파일(.go, .ts, .tsx)을 생성하면 반드시 같은 작업 내에서 테스트 파일도 함께 생성한다
- agent에게 코드 생성을 위임할 때 프롬프트에 "테스트도 함께 작성하라"는 지시를 반드시 포함한다
- 작업 완료 전 `make test` 또는 `npx vitest run`으로 전체 테스트 통과를 확인한다

## 테스트 커버리지 규칙

- 모든 프로덕션 코드의 테스트 커버리지는 100%를 목표로 작성한다
- 새로운 코드를 작성하면 해당 코드에 대한 테스트를 반드시 함께 작성한다
- 기존 코드를 수정하면 커버리지가 떨어지지 않도록 테스트도 함께 수정/추가한다
- Go 프로젝트: `go test ./... -cover`로 커버리지를 확인한다
- WebUI 프로젝트: `npx vitest run --coverage`로 커버리지를 확인한다
- 커버리지 리포트에서 Uncovered Line이 발견되면 해당 라인을 커버하는 테스트를 추가한다
- API 호출이 필요한 컴포넌트는 모킹(vi.mock)을 사용하여 테스트한다
