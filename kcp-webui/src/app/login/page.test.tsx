import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import LoginPage from './page'
import { useAuthStore } from '@/stores/authStore'

const mockPush = vi.fn()
vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: mockPush }),
}))

vi.mock('@/services/api', () => ({
  default: {
    post: vi.fn(),
  },
}))

import api from '@/services/api'

describe('LoginPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    useAuthStore.setState({ token: null, username: null, isAuthenticated: false })
  })

  it('로그인 폼을 렌더링한다', () => {
    render(<LoginPage />)
    expect(screen.getByText('KCP 로그인')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('사용자명')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('비밀번호')).toBeInTheDocument()
    expect(screen.getByText('로그인')).toBeInTheDocument()
  })

  it('로그인 성공 시 대시보드로 이동한다', async () => {
    vi.mocked(api.post).mockResolvedValue({
      data: { token: 'jwt-token', user: { username: 'admin' } },
    })

    render(<LoginPage />)
    fireEvent.change(screen.getByPlaceholderText('사용자명'), { target: { value: 'admin' } })
    fireEvent.change(screen.getByPlaceholderText('비밀번호'), { target: { value: 'pass' } })
    fireEvent.click(screen.getByText('로그인'))

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith('/auth/login', {
        username: 'admin',
        password: 'pass',
        authType: 'JWT',
      })
      expect(mockPush).toHaveBeenCalledWith('/dashboard')
    })

    const state = useAuthStore.getState()
    expect(state.isAuthenticated).toBe(true)
    expect(state.token).toBe('jwt-token')
  })

  it('로그인 실패 시 에러 메시지를 표시한다', async () => {
    vi.mocked(api.post).mockRejectedValue({
      response: { data: { error: { message: '자격 증명이 올바르지 않습니다' } } },
    })

    render(<LoginPage />)
    fireEvent.change(screen.getByPlaceholderText('사용자명'), { target: { value: 'admin' } })
    fireEvent.change(screen.getByPlaceholderText('비밀번호'), { target: { value: 'wrong' } })
    fireEvent.click(screen.getByText('로그인'))

    await waitFor(() => {
      expect(screen.getByText('자격 증명이 올바르지 않습니다')).toBeInTheDocument()
    })
  })

  it('로그인 중 버튼이 비활성화된다', async () => {
    vi.mocked(api.post).mockImplementation(() => new Promise(() => {})) // 영원히 pending

    render(<LoginPage />)
    fireEvent.change(screen.getByPlaceholderText('사용자명'), { target: { value: 'admin' } })
    fireEvent.change(screen.getByPlaceholderText('비밀번호'), { target: { value: 'pass' } })
    fireEvent.click(screen.getByText('로그인'))

    await waitFor(() => {
      expect(screen.getByText('로그인 중...')).toBeInTheDocument()
    })
  })
})
