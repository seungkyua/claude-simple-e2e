# 코드 생성 계획 — 장바구니 합계 계산기

## 프로젝트 정보
- **프로젝트명**: cart-calculator
- **코드 위치**: `cart/`
- **기술 스택**: Go + go test
- **TDD 적용**: Red → Green → Refactor

## 코드 생성 순서

### Phase 1: 프로젝트 구조 및 기반

- [x] Step 1: 프로젝트 초기화 — go.mod, 디렉토리 구조, Item 구조체

### Phase 2: 상품 합계 계산 (cart/)

- [x] Step 2: 상품 합계 구현 — CalculateSubtotal, 빈 장바구니/단일/복수 상품, 수량 곱셈
- [ ] Step 3: 입력 검증 구현 — 가격 0 이하 에러, 수량 0 이하 에러

### Phase 3: 할인 적용 (cart/)

- [ ] Step 4: 할인 구현 — ApplyDiscount, 쿠폰 검증, 정률 할인, 할인 후 최소 0원

### Phase 4: 배송비 계산 (cart/)

- [ ] Step 5: 배송비 구현 — CalculateShippingFee, 조건부 무료 배송

### Phase 5: 세금 및 최종 금액 (cart/)

- [ ] Step 6: 세금 구현 — CalculateVat, VAT 10%, 소수점 반올림
- [ ] Step 7: 최종 금액 구현 — CalculateTotal, 할인+세금+배송비 통합

## 스토리 매핑

| 기능 요구사항 | 구현 Step |
|--------------|----------|
| 1. 상품 합계 계산 | Step 1-3 |
| 2. 세금 계산 | Step 6 |
| 3. 할인 적용 | Step 4 |
| 4. 배송비 계산 | Step 5 |
| 최종 결제 금액 산출 | Step 7 |
