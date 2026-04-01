import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import NetworkPage from './page'
import { useAuthStore } from '@/stores/authStore'

vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: vi.fn(), replace: vi.fn() }),
  usePathname: () => '/network',
}))

vi.mock('@/services/api', () => ({
  default: { get: vi.fn() },
}))

import api from '@/services/api'

describe('NetworkPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    useAuthStore.setState({ token: 'test', username: 'admin', isAuthenticated: true })
  })

  it('네트워크 목록을 렌더링한다', async () => {
    vi.mocked(api.get).mockResolvedValue({
      data: {
        items: [
          { id: 'net-1', name: 'private', status: 'ACTIVE', subnets: ['sub-1', 'sub-2'], shared: false, 'router:external': false },
          { id: 'net-2', name: 'public', status: 'ACTIVE', subnets: ['sub-3'], shared: false, 'router:external': true },
        ],
      },
    })

    render(<NetworkPage />)

    await waitFor(() => {
      expect(screen.getByText('private')).toBeInTheDocument()
      expect(screen.getByText('public')).toBeInTheDocument()
    })
  })

  it('로딩 중 상태를 표시한다', () => {
    vi.mocked(api.get).mockImplementation(() => new Promise(() => {}))
    render(<NetworkPage />)
    expect(screen.getByText('로딩 중...')).toBeInTheDocument()
  })
})
