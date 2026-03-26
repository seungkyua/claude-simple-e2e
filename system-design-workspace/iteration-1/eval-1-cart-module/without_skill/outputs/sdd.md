# 시스템 설계서: 장바구니 합계 계산기

## 1. 개요

쇼핑몰 장바구니에 담긴 상품들의 최종 결제 금액을 산출하는 모듈이다.
상품 합계, 할인, 세금, 배송비를 순차적으로 계산하여 최종 금액을 반환한다.

---

## 2. 아키텍처 개요

```
[CartItem] ──▶ [CartCalculator] ──▶ 최종 결제 금액
                    │
                    ├── subtotal (상품 합계)
                    ├── discount (할인 적용)
                    ├── tax (세금 계산)
                    └── shipping (배송비 계산)
```

단일 모듈 구조로 설계하며, 외부 의존성 없이 순수 계산 로직으로 구성한다.

---

## 3. 도메인 모델

### 3.1 CartItem (장바구니 항목)

| 필드 | 타입 | 제약 조건 | 설명 |
|------|------|-----------|------|
| `name` | `string` | 필수, 비어있지 않을 것 | 상품명 |
| `price` | `number` | 0 초과 | 상품 단가 (원) |
| `quantity` | `integer` | 1 이상 | 수량 |

### 3.2 Coupon (쿠폰)

| 필드 | 타입 | 제약 조건 | 설명 |
|------|------|-----------|------|
| `code` | `string` | 필수 | 쿠폰 코드 |
| `discountRate` | `integer` | 1 ~ 100 | 할인율 (%) |

### 3.3 CartSummary (계산 결과)

| 필드 | 타입 | 설명 |
|------|------|------|
| `subtotal` | `number` | 상품 합계 (할인 전) |
| `discountAmount` | `number` | 할인 금액 |
| `discountedSubtotal` | `number` | 할인 후 상품 합계 |
| `tax` | `number` | 부가세 (VAT 10%) |
| `shippingFee` | `number` | 배송비 |
| `total` | `number` | 최종 결제 금액 |

---

## 4. 핵심 계산 흐름

### 4.1 계산 순서

```
1. 상품 합계 (subtotal) = SUM(price * quantity)
2. 할인 적용 (discountedSubtotal) = subtotal - (subtotal * discountRate / 100)
   - 할인 후 금액은 최소 0원 (음수 불가)
3. 세금 계산 (tax) = ROUND(discountedSubtotal * 0.1)
   - 소수점 이하 반올림
4. 배송비 결정 (shippingFee)
   - discountedSubtotal >= 50,000 → 0원
   - discountedSubtotal < 50,000 → 3,000원
5. 최종 금액 (total) = discountedSubtotal + tax + shippingFee
```

### 4.2 계산 예시

| 항목 | 값 |
|------|-----|
| 상품: A(10,000원 x 2), B(20,000원 x 1) | - |
| subtotal | 40,000원 |
| 쿠폰 할인 10% | -4,000원 |
| discountedSubtotal | 36,000원 |
| tax (10%) | 3,600원 |
| shippingFee (36,000 < 50,000) | 3,000원 |
| **total** | **42,600원** |

---

## 5. 모듈 인터페이스 설계

### 5.1 CartCalculator

```
class CartCalculator {
  // 상품 합계 계산
  calculateSubtotal(items: CartItem[]): number

  // 할인 적용
  applyDiscount(subtotal: number, couponCode: string): { discountAmount: number, discountedSubtotal: number }

  // 세금 계산
  calculateTax(discountedSubtotal: number): number

  // 배송비 계산
  calculateShippingFee(discountedSubtotal: number): number

  // 최종 결제 금액 산출 (통합)
  calculateTotal(items: CartItem[], couponCode?: string): CartSummary
}
```

### 5.2 메서드별 책임

| 메서드 | 입력 | 출력 | 책임 |
|--------|------|------|------|
| `calculateSubtotal` | CartItem[] | number | 각 항목의 price * quantity 합산 |
| `applyDiscount` | subtotal, couponCode | { discountAmount, discountedSubtotal } | 쿠폰 유효성 검증 및 할인 금액 계산 |
| `calculateTax` | discountedSubtotal | number | VAT 10% 반올림 계산 |
| `calculateShippingFee` | discountedSubtotal | number | 무료 배송 조건 판별 |
| `calculateTotal` | CartItem[], couponCode? | CartSummary | 전체 계산 흐름 조합 |

---

## 6. 입력 검증 규칙

모든 외부 입력은 처리 전에 반드시 검증한다 (Zero Trust Input 원칙).

| 대상 | 검증 규칙 | 위반 시 처리 |
|------|-----------|-------------|
| `price` | `price > 0` | `InvalidPriceError` 발생 |
| `quantity` | `Number.isInteger(quantity) && quantity >= 1` | `InvalidQuantityError` 발생 |
| `couponCode` | 등록된 쿠폰 코드와 일치 여부 | `InvalidCouponError` 발생 |
| `discountRate` | `1 <= rate <= 100`, 정수 | `InvalidDiscountRateError` 발생 |
| `items` | 빈 배열 허용 (합계 0 반환) | - |

---

## 7. 에러 처리

각 에러는 명확한 커스텀 에러 타입으로 구분한다.

| 에러 타입 | 발생 조건 | 메시지 예시 |
|-----------|-----------|-------------|
| `InvalidPriceError` | 상품 가격이 0 이하 | "상품 가격은 0보다 커야 합니다" |
| `InvalidQuantityError` | 수량이 1 미만 또는 정수가 아닌 경우 | "수량은 1 이상의 정수여야 합니다" |
| `InvalidCouponError` | 등록되지 않은 쿠폰 코드 | "유효하지 않은 쿠폰 코드입니다" |
| `InvalidDiscountRateError` | 할인율이 1~100 범위 밖 | "할인율은 1~100 사이여야 합니다" |

---

## 8. 보안 고려사항

- **입력 검증**: 모든 수치 입력에 대해 타입과 범위를 서버 측에서 검증한다.
- **정수 연산 정밀도**: 통화 계산 시 부동소수점 오차에 주의하며, 최종 결과에 반올림을 적용한다.
- **쿠폰 코드 보안**: 쿠폰 코드 검증 시 타이밍 공격 방지를 위해 상수 시간 비교를 고려한다.
- **로그 기록**: 에러 로그에 사용자 개인정보나 결제 정보를 포함하지 않는다.
- **민감정보 하드코딩 금지**: 쿠폰 데이터는 외부 저장소(환경변수, DB 등)에서 주입받는다.

---

## 9. TDD 테스트 계획

PRD의 기능 요구사항과 제약 조건을 기반으로 한 테스트 목록이다.

### 9.1 상품 합계 계산

| # | 테스트명 | 설명 |
|---|---------|------|
| 1 | `shouldReturnZeroWhenCartIsEmpty` | 빈 장바구니 → 합계 0 |
| 2 | `shouldCalculateSubtotalForSingleItem` | 단일 상품의 price * quantity |
| 3 | `shouldCalculateSubtotalForMultipleItems` | 복수 상품 합산 |

### 9.2 입력 검증

| # | 테스트명 | 설명 |
|---|---------|------|
| 4 | `shouldThrowErrorWhenPriceIsZeroOrNegative` | 가격 0 이하 → 에러 |
| 5 | `shouldThrowErrorWhenQuantityIsLessThanOne` | 수량 1 미만 → 에러 |
| 6 | `shouldThrowErrorWhenQuantityIsNotInteger` | 수량 비정수 → 에러 |

### 9.3 할인 적용

| # | 테스트명 | 설명 |
|---|---------|------|
| 7 | `shouldApplyPercentageDiscount` | 정률 할인 적용 |
| 8 | `shouldThrowErrorWhenCouponIsInvalid` | 미등록 쿠폰 → 에러 |
| 9 | `shouldNotAllowDiscountedPriceBelowZero` | 할인 후 금액 최소 0 |

### 9.4 세금 계산

| # | 테스트명 | 설명 |
|---|---------|------|
| 10 | `shouldCalculateTenPercentVAT` | VAT 10% 계산 |
| 11 | `shouldRoundTaxToNearestInteger` | 세금 소수점 반올림 |
| 12 | `shouldApplyTaxAfterDiscount` | 할인 후 금액 기준 세금 계산 |

### 9.5 배송비 계산

| # | 테스트명 | 설명 |
|---|---------|------|
| 13 | `shouldChargeBasicShippingFee` | 기본 배송비 3,000원 |
| 14 | `shouldWaiveShippingWhenSubtotalIsAboveThreshold` | 50,000원 이상 → 무료 배송 |
| 15 | `shouldNotIncludeShippingInTaxCalculation` | 배송비는 세금 미포함 |

### 9.6 최종 금액 통합

| # | 테스트명 | 설명 |
|---|---------|------|
| 16 | `shouldCalculateTotalWithAllComponents` | 전체 흐름 통합 검증 |
| 17 | `shouldCalculateTotalWithoutCoupon` | 쿠폰 미적용 시 전체 계산 |

---

## 10. 기술 스택 고려사항

- **언어**: TypeScript (타입 안전성 확보)
- **테스트 프레임워크**: Vitest 또는 Jest
- **라이선스**: MIT, Apache 2.0 등 허용 라이선스만 사용
- **외부 의존성**: 최소화 (순수 계산 모듈이므로 외부 라이브러리 불필요)
