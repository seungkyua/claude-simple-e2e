import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import Header from './Header'
import { useAuthStore } from '@/stores/authStore'

// Next.js 라우터 모킹
const mockPush = vi.fn()
vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: mockPush }),
}))

describe('Header', () => {
  beforeEach(() => {
    useAuthStore.setState({ token: 'test', username: 'admin', isAuthenticated: true })
    mockPush.mockClear()
  })

  it('사용자명을 표시한다', () => {
    render(<Header />)
    expect(screen.getByText('admin')).toBeInTheDocument()
  })

  it('로그아웃 버튼이 있다', () => {
    render(<Header />)
    expect(screen.getByText('로그아웃')).toBeInTheDocument()
  })

  it('로그아웃 클릭 시 스토어를 초기화하고 로그인 페이지로 이동한다', () => {
    render(<Header />)
    fireEvent.click(screen.getByText('로그아웃'))
    const state = useAuthStore.getState()
    expect(state.isAuthenticated).toBe(false)
    expect(mockPush).toHaveBeenCalledWith('/login')
  })
})
