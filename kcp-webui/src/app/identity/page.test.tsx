import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import IdentityPage from './page'
import { useAuthStore } from '@/stores/authStore'

vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: vi.fn(), replace: vi.fn() }),
  usePathname: () => '/identity',
}))

vi.mock('@/services/api', () => ({
  default: { get: vi.fn() },
}))

import api from '@/services/api'

describe('IdentityPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    useAuthStore.setState({ token: 'test', username: 'admin', isAuthenticated: true })
  })

  it('프로젝트 목록을 렌더링한다', async () => {
    vi.mocked(api.get).mockResolvedValue({
      data: {
        items: [
          { id: 'proj-1', name: 'myproject', domain_id: 'default', description: 'Main project', enabled: true },
          { id: 'proj-2', name: 'demo', domain_id: 'default', description: 'Demo project', enabled: true },
        ],
      },
    })

    render(<IdentityPage />)

    await waitFor(() => {
      expect(screen.getByText('myproject')).toBeInTheDocument()
      expect(screen.getByText('demo')).toBeInTheDocument()
      expect(screen.getByText('Main project')).toBeInTheDocument()
    })
  })
})
