import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import AuditPage from './page'
import { useAuthStore } from '@/stores/authStore'

vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: vi.fn(), replace: vi.fn() }),
  usePathname: () => '/audit',
}))

vi.mock('@/services/api', () => ({
  default: { get: vi.fn() },
}))

import api from '@/services/api'

describe('AuditPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    useAuthStore.setState({ token: 'test', username: 'admin', isAuthenticated: true })
  })

  it('감사 로그 데이터를 렌더링한다', async () => {
    vi.mocked(api.get).mockResolvedValue({
      data: {
        items: [
          {
            id: 'log-1',
            user_id: 'user-1',
            username: 'operator',
            action: 'CREATE',
            resource_type: 'VM',
            source: 'CLI',
            status_code: 201,
            ip_address: '127.0.0.1',
            created_at: '2026-04-01T10:00:00Z',
          },
          {
            id: 'log-2',
            user_id: 'user-1',
            username: 'operator',
            action: 'READ',
            resource_type: 'NETWORK',
            source: 'WEBUI',
            status_code: 200,
            ip_address: '127.0.0.1',
            created_at: '2026-04-01T10:01:00Z',
          },
        ],
      },
    })

    render(<AuditPage />)

    await waitFor(() => {
      expect(screen.getByText('CREATE')).toBeInTheDocument()
      expect(screen.getByText('READ')).toBeInTheDocument()
      expect(screen.getByText('VM')).toBeInTheDocument()
      expect(screen.getByText('NETWORK')).toBeInTheDocument()
      expect(screen.getByText('CLI')).toBeInTheDocument()
      expect(screen.getByText('WEBUI')).toBeInTheDocument()
    })
  })

  it('로딩 중 상태를 표시한다', () => {
    vi.mocked(api.get).mockImplementation(() => new Promise(() => {}))
    render(<AuditPage />)
    expect(screen.getByText('로딩 중...')).toBeInTheDocument()
  })

  it('빈 데이터일 때 테이블 헤더만 표시된다', async () => {
    vi.mocked(api.get).mockResolvedValue({ data: { items: [] } })

    render(<AuditPage />)

    await waitFor(() => {
      expect(screen.queryByText('로딩 중...')).not.toBeInTheDocument()
    })

    // "감사 로그"가 Sidebar와 제목 두 곳에 있으므로 getAllByText 사용
    expect(screen.getAllByText('감사 로그').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('Action')).toBeInTheDocument()
    expect(screen.getByText('Resource')).toBeInTheDocument()
  })
})
