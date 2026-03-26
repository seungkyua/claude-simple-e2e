# 시스템 설계서: 장바구니 합계 계산기

## 1. 프로젝트 개요

| 항목 | 내용 |
|---|---|
| 프로젝트명 | 장바구니 합계 계산기 |
| 한 줄 설명 | 쇼핑몰 장바구니의 상품 합계, 세금, 할인, 배송비를 포함한 최종 결제 금액을 산출하는 순수 계산 모듈 |
| 기술 스택 | TypeScript, Vitest (테스트) |
| 작성일 | 2026-03-26 |
| 기반 문서 | prd.md |

---

## 2. 전체 시스템 아키텍처

본 모듈은 DB나 외부 서비스 없이 동작하는 순수 계산 모듈이다. 호출자(쇼핑몰 애플리케이션)가 장바구니 데이터와 쿠폰 정보를 전달하면, 모듈이 합계를 계산하여 결과를 반환한다.

```mermaid
flowchart LR
    Caller[호출자<br/>쇼핑몰 앱] -->|CartItem[] + 쿠폰코드| Module[장바구니 합계 계산기]
    Module -->|CartSummary| Caller

    subgraph Module[장바구니 합계 계산기]
        direction TB
        A[상품 합계 계산] --> B[할인 적용]
        B --> C[세금 계산]
        B --> D[배송비 계산]
        C --> E[최종 금액 산출]
        D --> E
    end
```

**데이터 흐름 순서:**
1. 호출자가 장바구니 상품 목록과 (선택적으로) 쿠폰 코드를 전달
2. 상품 합계 계산: 각 상품의 `단가 x 수량`을 합산
3. 할인 적용: 유효한 쿠폰이 있으면 상품 합계에서 정률 할인 적용
4. 세금 계산: 할인 후 금액에 부가세 10% 적용 (소수점 반올림)
5. 배송비 결정: 할인 후 금액 기준 50,000원 이상이면 무료, 미만이면 3,000원
6. 최종 금액 = 할인 후 금액 + 세금 + 배송비

---

## 3. 데이터 구조체 및 인터페이스 명세

> DB가 없는 순수 모듈이므로, ERD 대신 핵심 데이터 구조체를 정의한다.

### 3.1 CartItem (장바구니 항목)

| 필드명 | 타입 | 제약조건 | 설명 |
|---|---|---|---|
| `name` | `string` | 필수, 1자 이상 | 상품명 |
| `price` | `number` | 필수, 0 초과 | 상품 단가 (원) |
| `quantity` | `number` | 필수, 1 이상 정수 | 수량 |

### 3.2 Coupon (쿠폰)

| 필드명 | 타입 | 제약조건 | 설명 |
|---|---|---|---|
| `code` | `string` | 필수, 1자 이상 | 쿠폰 코드 |
| `discountRate` | `number` | 1~100 정수 | 할인율 (%) |

### 3.3 CartSummary (계산 결과)

| 필드명 | 타입 | 설명 |
|---|---|---|
| `subtotal` | `number` | 상품 합계 (할인 전) |
| `discountAmount` | `number` | 할인 금액 |
| `afterDiscount` | `number` | 할인 후 금액 |
| `tax` | `number` | 부가세 (10%, 반올림) |
| `shippingFee` | `number` | 배송비 (0 또는 3,000) |
| `total` | `number` | 최종 결제 금액 |

### 3.4 TypeScript 인터페이스

```typescript
interface CartItem {
  name: string;
  price: number;    // 0 초과
  quantity: number;  // 1 이상 정수
}

interface Coupon {
  code: string;
  discountRate: number;  // 1~100 정수(%)
}

interface CartSummary {
  subtotal: number;
  discountAmount: number;
  afterDiscount: number;
  tax: number;
  shippingFee: number;
  total: number;
}
```

---

## 4. 핵심 API 인터페이스 명세

> HTTP API가 아닌 순수 모듈이므로, public 함수 시그니처와 입출력 명세를 정의한다.

### 4.1 `calculateSubtotal(items: CartItem[]): number`

**설명**: 장바구니 상품 합계 계산

| 구분 | 내용 |
|---|---|
| 입력 | `CartItem[]` - 장바구니 상품 목록 |
| 출력 | `number` - 상품 합계 금액 |
| 빈 배열 | `0` 반환 |
| 예외 | `InvalidCartItemError` - price <= 0 또는 quantity < 1인 경우 |

### 4.2 `applyDiscount(subtotal: number, coupon?: Coupon): { afterDiscount: number; discountAmount: number }`

**설명**: 쿠폰 할인 적용

| 구분 | 내용 |
|---|---|
| 입력 | `subtotal` - 상품 합계, `coupon` - 쿠폰 (선택) |
| 출력 | `{ afterDiscount, discountAmount }` |
| 쿠폰 없음 | `{ afterDiscount: subtotal, discountAmount: 0 }` |
| 예외 | `InvalidCouponError` - 할인율이 1~100 범위 밖인 경우 |
| 비즈니스 규칙 | 할인 후 금액은 최소 0 (음수 불가) |

### 4.3 `calculateTax(amount: number): number`

**설명**: 부가세 계산

| 구분 | 내용 |
|---|---|
| 입력 | `amount` - 할인 후 금액 |
| 출력 | `number` - 세금 (10%, 소수점 반올림) |
| 계산식 | `Math.round(amount * 0.1)` |

### 4.4 `calculateShippingFee(afterDiscount: number): number`

**설명**: 배송비 계산

| 구분 | 내용 |
|---|---|
| 입력 | `afterDiscount` - 할인 후 상품 합계 |
| 출력 | `number` - 배송비 (0 또는 3,000) |
| 무료 기준 | 할인 후 금액 >= 50,000원 |
| 기본 배송비 | 3,000원 |

### 4.5 `calculateCart(items: CartItem[], coupon?: Coupon): CartSummary`

**설명**: 장바구니 전체 계산 (메인 퍼사드 함수)

| 구분 | 내용 |
|---|---|
| 입력 | `items` - 장바구니 상품 목록, `coupon` - 쿠폰 (선택) |
| 출력 | `CartSummary` - 전체 계산 결과 |
| 예외 | `InvalidCartItemError`, `InvalidCouponError` |

### 4.6 에러/예외 정의

| 에러 클래스 | 발생 조건 | 메시지 예시 |
|---|---|---|
| `InvalidCartItemError` | 상품 가격 <= 0 또는 수량 < 1 또는 수량이 정수가 아닌 경우 | `"상품 가격은 0보다 커야 합니다"` |
| `InvalidCouponError` | 쿠폰 할인율이 1~100 범위 밖이거나 유효하지 않은 코드 | `"유효하지 않은 쿠폰 코드입니다"` |

---

## 5. 폴더 구조 및 모듈 분리 전략

```
src/
├── cart/
│   ├── types.ts            # CartItem, Coupon, CartSummary 인터페이스
│   ├── errors.ts           # InvalidCartItemError, InvalidCouponError 정의
│   ├── calculateSubtotal.ts    # 상품 합계 계산
│   ├── applyDiscount.ts        # 할인 적용
│   ├── calculateTax.ts         # 세금 계산
│   ├── calculateShippingFee.ts # 배송비 계산
│   ├── calculateCart.ts        # 퍼사드 - 전체 계산 통합
│   └── index.ts            # public API 재수출
└── __tests__/
    └── cart/
        ├── calculateSubtotal.test.ts
        ├── applyDiscount.test.ts
        ├── calculateTax.test.ts
        ├── calculateShippingFee.test.ts
        └── calculateCart.test.ts
```

**모듈 분리 기준:**
- **단일 책임 원칙**: 각 계산 로직(합계, 할인, 세금, 배송비)은 독립 함수로 분리
- **퍼사드 패턴**: `calculateCart`가 개별 함수들을 조합하여 통합 결과 제공
- **타입 분리**: 인터페이스와 에러 클래스는 별도 파일로 관리하여 순환 참조 방지
- **테스트 대응**: 각 함수별 독립 테스트 파일로 TDD 사이클에 적합한 구조

---

## 6. 내부 데이터 흐름도

> 프론트엔드가 없는 순수 모듈이므로, 상태 관리 대신 데이터 변환 파이프라인을 시각화한다.

```mermaid
flowchart TD
    Start([입력: CartItem[] + Coupon?]) --> Validate{입력 검증}
    Validate -->|유효하지 않음| Error[에러 반환]
    Validate -->|유효| Subtotal[상품 합계 계산<br/>items.reduce: price × quantity]
    Subtotal --> HasCoupon{쿠폰 있음?}
    HasCoupon -->|예| Discount[할인 적용<br/>subtotal × discountRate / 100<br/>최소 0 보장]
    HasCoupon -->|아니오| NoDiscount[할인 없음<br/>afterDiscount = subtotal]
    Discount --> Tax[세금 계산<br/>Math.round: afterDiscount × 0.1]
    NoDiscount --> Tax
    Discount --> Shipping{afterDiscount >= 50000?}
    NoDiscount --> Shipping
    Shipping -->|예| Free[배송비 = 0]
    Shipping -->|아니오| Paid[배송비 = 3000]
    Tax --> Total[최종 금액 산출<br/>afterDiscount + tax + shippingFee]
    Free --> Total
    Paid --> Total
    Total --> Result([출력: CartSummary])
```

**핵심 데이터 변환 단계:**

| 단계 | 입력 | 출력 | 규칙 |
|---|---|---|---|
| 1. 검증 | `CartItem[]` | 유효한 `CartItem[]` | price > 0, quantity >= 1 (정수) |
| 2. 합계 | `CartItem[]` | `subtotal: number` | `SUM(price * quantity)` |
| 3. 할인 | `subtotal + Coupon?` | `afterDiscount: number` | `MAX(0, subtotal - subtotal * rate / 100)` |
| 4. 세금 | `afterDiscount` | `tax: number` | `Math.round(afterDiscount * 0.1)` |
| 5. 배송비 | `afterDiscount` | `shippingFee: number` | `afterDiscount >= 50000 ? 0 : 3000` |
| 6. 합산 | `afterDiscount + tax + shippingFee` | `total: number` | 단순 덧셈 |

---

## 7. 보안 설계

> 순수 계산 모듈이므로 네트워크/인증 위협은 해당되지 않지만, 입력 검증 및 방어적 프로그래밍 관점에서 다음을 적용한다.

### 7.1 위협 분석 및 대응

| 위협 | 대응 방안 | 구현 위치 |
|---|---|---|
| 비정상 입력 (음수 가격, 소수 수량) | 모든 입력값 타입/범위 검증, 정수 체크 | 각 함수 진입부 |
| 부동소수점 오류 | 통화 계산 시 `Math.round()` 적용으로 정수 기반 처리 | `calculateTax` |
| 대량 데이터 DoS | 배열 길이 상한 검증 (선택적) | `calculateSubtotal` |
| Prototype Pollution | 입력 객체의 프로토타입 체인 접근 방지, 순수 함수로만 구현 | 전체 모듈 |

### 7.2 입력 검증 정책

```typescript
// 방어적 프로그래밍: 모든 외부 입력은 신뢰하지 않는다
function validateCartItem(item: CartItem): void {
  if (typeof item.price !== 'number' || item.price <= 0) {
    throw new InvalidCartItemError('상품 가격은 0보다 커야 합니다');
  }
  if (typeof item.quantity !== 'number' || item.quantity < 1 || !Number.isInteger(item.quantity)) {
    throw new InvalidCartItemError('수량은 1 이상의 정수여야 합니다');
  }
}

function validateCoupon(coupon: Coupon): void {
  if (typeof coupon.discountRate !== 'number' || coupon.discountRate < 1 || coupon.discountRate > 100 || !Number.isInteger(coupon.discountRate)) {
    throw new InvalidCouponError('쿠폰 할인율은 1~100 사이의 정수여야 합니다');
  }
}
```

### 7.3 민감 데이터 처리

- 본 모듈은 민감 개인정보를 직접 다루지 않음
- 에러 메시지에 내부 구현 세부사항 노출 방지
- 쿠폰 코드 등 비즈니스 데이터는 로그에 마스킹 처리 권장

### 7.4 의존성 라이선스 검토

| 패키지 | 용도 | 라이선스 | 허용 여부 |
|---|---|---|---|
| TypeScript | 언어 | Apache-2.0 | 허용 |
| Vitest | 테스트 | MIT | 허용 |

> GPL류 라이선스 패키지는 사용하지 않는다.

---

## 8. 예외 처리 및 에러 전략

### 8.1 에러 분류 체계

| 분류 | 에러 클래스 | 설명 | 복구 가능 여부 |
|---|---|---|---|
| 비즈니스 에러 | `InvalidCartItemError` | 상품 데이터 제약 위반 | 호출자가 입력 수정 후 재시도 |
| 비즈니스 에러 | `InvalidCouponError` | 쿠폰 유효성 위반 | 호출자가 쿠폰 변경 후 재시도 |

### 8.2 에러 클래스 설계

```typescript
// 기본 도메인 에러
class CartError extends Error {
  constructor(message: string) {
    super(message);
    this.name = this.constructor.name;
  }
}

class InvalidCartItemError extends CartError {}
class InvalidCouponError extends CartError {}
```

### 8.3 에러 처리 원칙

- **Fail Fast**: 잘못된 입력은 계산 시작 전에 즉시 거부
- **명확한 메시지**: 사용자/개발자가 원인을 즉시 파악할 수 있는 에러 메시지
- **계층적 에러**: `CartError`를 상속하여 호출자가 `instanceof`로 분기 가능
- **민감정보 미노출**: 에러 메시지에 시스템 내부 정보 포함 금지

### 8.4 로깅 권장사항

- 에러 발생 시 입력 값의 구조만 기록 (실제 값 노출 최소화)
- 쿠폰 코드는 마스킹 처리 (예: `"ABC***"`)

---

## 9. 3줄 요약 및 비유

> **3줄 요약**
> 1. 장바구니에 담긴 상품들의 금액을 합산하고, 쿠폰 할인과 세금, 배송비를 적용하여 최종 결제 금액을 계산합니다.
> 2. 모든 입력값은 철저히 검증하여 잘못된 데이터가 계산에 사용되는 것을 원천 차단합니다.
> 3. 각 계산 단계(합계, 할인, 세금, 배송비)는 독립 함수로 분리되어 테스트와 유지보수가 용이합니다.
>
> **비유로 이해하기**
> 이 모듈은 마트 계산대의 점원과 비슷합니다. 손님(호출자)이 장바구니(CartItem[])를 가져오면,
> 점원은 먼저 각 상품의 가격표가 올바른지 확인합니다(입력 검증).
> 그다음 상품 합계를 계산하고, 할인 쿠폰이 있으면 적용한 뒤(할인 계산),
> 부가세를 추가하고(세금 계산), 일정 금액 이상이면 무료 배송 도장을 찍어줍니다(배송비 결정).
> 최종적으로 영수증(CartSummary)을 출력하여 손님에게 전달합니다.
