import { describe, it, expect, beforeEach } from 'vitest'
import { useAuthStore } from './authStore'

describe('authStore', () => {
  beforeEach(() => {
    // 각 테스트 전 스토어 초기화
    useAuthStore.setState({ token: null, username: null, isAuthenticated: false })
  })

  it('초기 상태는 미인증이다', () => {
    const state = useAuthStore.getState()
    expect(state.isAuthenticated).toBe(false)
    expect(state.token).toBeNull()
    expect(state.username).toBeNull()
  })

  it('login 호출 시 인증 상태로 변경된다', () => {
    useAuthStore.getState().login('test-token', 'admin')
    const state = useAuthStore.getState()
    expect(state.isAuthenticated).toBe(true)
    expect(state.token).toBe('test-token')
    expect(state.username).toBe('admin')
  })

  it('logout 호출 시 미인증 상태로 변경된다', () => {
    useAuthStore.getState().login('test-token', 'admin')
    useAuthStore.getState().logout()
    const state = useAuthStore.getState()
    expect(state.isAuthenticated).toBe(false)
    expect(state.token).toBeNull()
    expect(state.username).toBeNull()
  })
})
