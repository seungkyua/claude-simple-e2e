import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import AuthGuard from './AuthGuard'
import { useAuthStore } from '@/stores/authStore'

const mockReplace = vi.fn()
vi.mock('next/navigation', () => ({
  useRouter: () => ({ replace: mockReplace }),
}))

describe('AuthGuard', () => {
  beforeEach(() => {
    mockReplace.mockClear()
  })

  it('인증된 상태면 자식 컴포넌트를 렌더링한다', () => {
    useAuthStore.setState({ token: 'test', username: 'admin', isAuthenticated: true })
    render(
      <AuthGuard>
        <div>보호된 콘텐츠</div>
      </AuthGuard>
    )
    expect(screen.getByText('보호된 콘텐츠')).toBeInTheDocument()
  })

  it('미인증 상태면 자식 컴포넌트를 렌더링하지 않고 리다이렉트한다', () => {
    useAuthStore.setState({ token: null, username: null, isAuthenticated: false })
    render(
      <AuthGuard>
        <div>보호된 콘텐츠</div>
      </AuthGuard>
    )
    expect(screen.queryByText('보호된 콘텐츠')).not.toBeInTheDocument()
    expect(mockReplace).toHaveBeenCalledWith('/login')
  })
})
