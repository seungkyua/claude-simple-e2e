/** 페이지네이션 정보 */
export interface Pagination {
  page: number
  size: number
  total: number
}

/** 목록 조회 공통 응답 */
export interface ListResponse<T> {
  items: T[]
  pagination: Pagination
}

/** API 에러 응답 */
export interface ErrorResponse {
  error: {
    code: string
    message: string
    status: number
    detail?: string
    traceId?: string
  }
}

/** 인증 응답 */
export interface AuthResponse {
  token: string
  authType: string
  expiresAt: string
  user: {
    id: string
    username: string
    role: string
  }
}
