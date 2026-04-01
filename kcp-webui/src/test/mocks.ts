import { vi } from 'vitest'

// Next.js 라우터 모킹 공통 헬퍼
export const mockPush = vi.fn()
export const mockReplace = vi.fn()

export function mockNextNavigation() {
  vi.mock('next/navigation', () => ({
    useRouter: () => ({ push: mockPush, replace: mockReplace }),
    usePathname: () => '/dashboard',
  }))
}

// API 응답 생성 헬퍼
export function mockApiResponse(data: unknown) {
  return Promise.resolve({ data })
}

export function mockApiError(status: number, message: string) {
  return Promise.reject({
    response: { status, data: { error: { code: 'ERROR', message, status } } },
  })
}
