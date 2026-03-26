# TDD 제약 규칙 (상세)

## Red 단계 규칙

- 테스트는 반드시 실패해야 진행 가능
- 컴파일 에러도 "Red"로 인정하지 않는다 — 테스트가 실행되어 실패해야 함
- 테스트 이름 형식: `should[ExpectedBehavior]When[Condition]`
  - 예: `shouldReturnZeroWhenListIsEmpty`
  - 예: `shouldThrowExceptionWhenInputIsNegative`
- 한 번에 하나의 assert만 작성 (가능한 경우)

## Green 단계 규칙

- 테스트를 통과시키는 **가장 단순한** 코드를 작성한다
- "너무 단순하지 않나?" 싶을 정도가 적당하다
- 하드코딩도 허용 — 이후 테스트가 일반화를 강제할 것이다
- 프로덕션 코드의 우아함은 이 단계의 목표가 아니다

## Refactor 단계 규칙

- 반드시 모든 테스트가 통과한 상태에서만 시작
- 한 번에 하나의 리팩토링만 수행
- 각 리팩토링 후 즉시 테스트 재실행
- 사용할 리팩토링 패턴 이름을 명시한다
  - 예: "Extract Method", "Rename Variable", "Inline Temp"

## 결함 수정 규칙

결함 발견 시 순서:
1. API 레벨 실패 테스트 먼저 작성
2. 결함을 재현하는 최소 단위 테스트 작성
3. 두 테스트를 모두 통과시키는 코드 구현

## 테스트 범위 규칙

- 단위 테스트: 매 사이클마다 실행
- 통합 테스트: Green 달성 후 실행
- Long-running 테스트: 커밋 전에만 실행
- 테스트 커버리지: 새로운 동작은 반드시 테스트로 커버