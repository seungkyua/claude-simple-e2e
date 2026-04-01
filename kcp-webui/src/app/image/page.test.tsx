import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import ImagePage from './page'
import { useAuthStore } from '@/stores/authStore'

vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: vi.fn(), replace: vi.fn() }),
  usePathname: () => '/image',
}))

vi.mock('@/services/api', () => ({
  default: { get: vi.fn() },
}))

import api from '@/services/api'

describe('ImagePage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    useAuthStore.setState({ token: 'test', username: 'admin', isAuthenticated: true })
  })

  it('이미지 목록을 렌더링한다', async () => {
    vi.mocked(api.get).mockResolvedValue({
      data: {
        items: [
          { id: 'img-1', name: 'ubuntu-22', status: 'active', disk_format: 'qcow2', size: 2147483648, visibility: 'public' },
          { id: 'img-2', name: 'cirros', status: 'active', disk_format: 'qcow2', size: 16338944, visibility: 'public' },
        ],
      },
    })

    render(<ImagePage />)

    await waitFor(() => {
      expect(screen.getByText('ubuntu-22')).toBeInTheDocument()
      expect(screen.getByText('cirros')).toBeInTheDocument()
    })
  })

  it('이미지 크기를 MB로 표시한다', async () => {
    vi.mocked(api.get).mockResolvedValue({
      data: {
        items: [
          { id: 'img-1', name: 'test-img', status: 'active', disk_format: 'qcow2', size: 10485760, visibility: 'public' },
        ],
      },
    })

    render(<ImagePage />)

    await waitFor(() => {
      expect(screen.getByText('10 MB')).toBeInTheDocument()
    })
  })
})
