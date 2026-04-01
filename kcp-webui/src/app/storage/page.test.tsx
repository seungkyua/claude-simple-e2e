import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import StoragePage from './page'
import { useAuthStore } from '@/stores/authStore'

vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: vi.fn(), replace: vi.fn() }),
  usePathname: () => '/storage',
}))

vi.mock('@/services/api', () => ({
  default: { get: vi.fn() },
}))

import api from '@/services/api'

describe('StoragePage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    useAuthStore.setState({ token: 'test', username: 'admin', isAuthenticated: true })
  })

  it('볼륨 목록을 렌더링한다', async () => {
    vi.mocked(api.get).mockResolvedValue({
      data: {
        items: [
          { id: 'vol-1', name: 'data-vol', status: 'available', size: 100, volume_type: 'ssd' },
        ],
      },
    })

    render(<StoragePage />)

    await waitFor(() => {
      expect(screen.getByText('data-vol')).toBeInTheDocument()
      expect(screen.getByText('available')).toBeInTheDocument()
    })
  })

  it('Storage 서비스 미설치 시 에러 메시지를 표시한다', async () => {
    vi.mocked(api.get).mockRejectedValue({
      response: { data: { error: { message: "서비스 'volumev3'의 엔드포인트를 찾을 수 없습니다" } } },
    })

    render(<StoragePage />)

    await waitFor(() => {
      expect(screen.getByText("서비스 'volumev3'의 엔드포인트를 찾을 수 없습니다")).toBeInTheDocument()
    })
  })
})
