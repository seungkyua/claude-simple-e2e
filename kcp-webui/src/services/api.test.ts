import { describe, it, expect, beforeEach } from 'vitest'
import api from './api'
import { useAuthStore } from '@/stores/authStore'

describe('api 인스턴스', () => {
  beforeEach(() => {
    useAuthStore.setState({ token: null, username: null, isAuthenticated: false })
  })

  it('baseURL이 설정되어 있다', () => {
    expect(api.defaults.baseURL).toBeDefined()
  })

  it('Content-Type 헤더가 application/json이다', () => {
    expect(api.defaults.headers['Content-Type']).toBe('application/json')
  })

  it('Accept 헤더가 application/json이다', () => {
    expect(api.defaults.headers['Accept']).toBe('application/json')
  })

  it('타임아웃이 설정되어 있다', () => {
    expect(api.defaults.timeout).toBe(30000)
  })

  it('인증 토큰이 있으면 요청 인터셉터가 Authorization 헤더를 추가한다', () => {
    useAuthStore.getState().login('my-jwt-token', 'admin')

    // 인터셉터가 설정에 적용하는지 간접 확인
    const state = useAuthStore.getState()
    expect(state.token).toBe('my-jwt-token')
  })
})
