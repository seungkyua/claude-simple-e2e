import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, act } from '@testing-library/react'
import DashboardPage from './page'
import { useAuthStore } from '@/stores/authStore'

vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: vi.fn(), replace: vi.fn() }),
  usePathname: () => '/dashboard',
}))

vi.mock('@/services/api', () => ({
  default: {
    get: vi.fn(),
  },
}))

import api from '@/services/api'

describe('DashboardPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    useAuthStore.setState({ token: 'test', username: 'admin', isAuthenticated: true })
  })

  it('로딩 중 상태를 표시한다', () => {
    vi.mocked(api.get).mockImplementation(() => new Promise(() => {}))
    render(<DashboardPage />)
    expect(screen.getByText('로딩 중...')).toBeInTheDocument()
  })

  it('통계 데이터를 표시한다', async () => {
    vi.mocked(api.get).mockResolvedValue({
      data: {
        stats: {
          servers: 5,
          networks: 7,
          subnets: 8,
          images: 4,
          routers: 2,
          security_groups: 6,
          projects: 3,
          users: 10,
        },
      },
    })

    render(<DashboardPage />)

    await waitFor(() => {
      expect(screen.getByText('5')).toBeInTheDocument()   // servers
      expect(screen.getByText('VM 인스턴스')).toBeInTheDocument()
      expect(screen.getByText('7')).toBeInTheDocument()   // networks
      expect(screen.getByText('10')).toBeInTheDocument()  // users
    })
  })

  it('API 실패 시에도 0 값으로 렌더링된다', async () => {
    vi.mocked(api.get).mockRejectedValue(new Error('Network error'))

    await act(async () => {
      render(<DashboardPage />)
    })

    // 에러 후 loading=false → 0 값 표시
    await waitFor(() => {
      expect(screen.getByText('VM 인스턴스')).toBeInTheDocument()
    })
  })
})
