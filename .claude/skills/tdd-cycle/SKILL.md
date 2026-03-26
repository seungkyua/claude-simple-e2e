---
name: tdd-cycle
description: Kent Beck의 TDD 사이클(Red→Green→Refactor)을 단계별로 실행할 때 사용. "go", "다음 테스트", "TDD로 구현해줘", "다음 단계" 같은 요청 시 반드시 이 skill을 참조. plan.md가 존재하는 프로젝트에서 코드 작성 요청이 오면 항상 이 skill을 트리거. 테스트 작성, 구현, 리팩토링 요청이 오면 반드시 이 절차를 따른다.
---

# TDD Cycle Skill

## "go" 명령 실행 절차

### Step 1: plan.md에서 다음 테스트 찾기

```
1. plan.md를 읽는다
2. [ ] 로 표시된 항목 중 첫 번째를 선택한다
3. 선택한 항목을 사용자에게 확인한다:
   "다음 테스트: [항목 내용] — 진행할까요?"
4. 확인 후 진행
```

### Step 2: RED — 실패하는 테스트 작성

```
1. 테스트 파일에 새 테스트 함수 추가
2. 테스트 이름: should[ExpectedBehavior]When[Condition]
3. 구현 코드는 작성하지 않는다
4. 테스트 실행 → 실패 메시지 출력
5. "❌ RED: [실패 내용]" 형식으로 보고
```

**출력 예시:**
```
❌ RED
테스트: shouldReturnSumWhenGivenTwoNumbers
실패 이유: TypeError: add is not a function
```

### Step 3: GREEN — 최소 구현

```
1. 테스트를 통과시키는 가장 단순한 코드 작성
2. 우아함 불필요 — 동작만 맞으면 됨
3. 테스트 실행 → 통과 확인
4. "✅ GREEN: 모든 테스트 통과" 형식으로 보고
5. 테스트 통과 실패 시 Step 3 반복 (구현 수정)
```

**출력 예시:**
```
✅ GREEN
구현: add(a, b) { return a + b }
통과: 3/3 tests passing
```

### Step 4: REFACTOR — 구조 개선

```
중복이나 불명확한 이름이 있는 경우에만 수행:
1. 적용할 리팩토링 패턴 이름 명시
2. 변경 사항 설명
3. 테스트 재실행 → 여전히 통과 확인
4. "🔧 REFACTOR: [패턴명]" 형식으로 보고

필요 없으면 이 단계 건너뜀
```

**출력 예시:**
```
🔧 REFACTOR: Extract Method
변경: 검증 로직을 validateInput()으로 추출
결과: 3/3 tests still passing
```

### Step 5: 커밋

```
테스트 + 구현이 함께인 경우:
  git add -A
  git commit -m "[behavioral] should[ExpectedBehavior]When[Condition]"

리팩토링이 있는 경우 별도 커밋:
  git add -A
  git commit -m "[structural] Extract Method: validateInput"
```

### Step 6: plan.md 업데이트

```
1. 완료된 항목의 [ ]를 [x]로 변경
2. 완료 보고:

📋 진행 현황
완료: [x] 항목명
남은 테스트: N개
다음: "go"를 입력하면 다음 테스트를 진행합니다.
```

---

## 흐름 요약

```
plan.md 읽기
    ↓
[ ] 항목 선택 → 사용자 확인
    ↓
❌ RED: 실패 테스트 작성 & 실행
    ↓
✅ GREEN: 최소 구현 & 테스트 통과
    ↓
🔧 REFACTOR: 필요 시 구조 개선 (선택)
    ↓
💾 COMMIT: [behavioral] + 필요 시 [structural] 별도
    ↓
📋 plan.md [x] 마킹 & 현황 보고
```

---

## 예외 상황 처리

| 상황 | 대응 |
|---|---|
| plan.md가 없음 | "plan.md가 없습니다. 생성해드릴까요?" |
| 모든 테스트 완료 | "🎉 plan.md의 모든 테스트가 완료되었습니다." |
| 테스트가 예상과 다르게 실패 | 사용자에게 보고 후 방향 확인 |
| 구현 방향이 불명확 | "이 테스트의 의도를 확인해주세요: [내용]" |