import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import ComputePage from './page'
import { useAuthStore } from '@/stores/authStore'

vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: vi.fn(), replace: vi.fn() }),
  usePathname: () => '/compute',
}))

vi.mock('@/services/api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    delete: vi.fn(),
  },
}))

import api from '@/services/api'

const mockServers = [
  {
    id: 'aaaa-1111',
    name: 'web-server',
    status: 'ACTIVE',
    flavor: { id: '1', name: 'm1.small' },
    image: { id: 'img-1', name: 'ubuntu-22' },
    networks: 'private=10.0.0.5',
  },
  {
    id: 'bbbb-2222',
    name: 'db-server',
    status: 'SHUTOFF',
    flavor: { id: '2', name: 'm1.medium' },
    image: { id: 'img-2', name: 'centos-9' },
    networks: 'private=10.0.0.6',
  },
]

describe('ComputePage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    useAuthStore.setState({ token: 'test', username: 'admin', isAuthenticated: true })
  })

  it('서버 목록을 렌더링한다', async () => {
    vi.mocked(api.get).mockResolvedValue({
      data: { items: mockServers, pagination: { page: 1, size: 2, total: 2 } },
    })

    render(<ComputePage />)

    await waitFor(() => {
      expect(screen.getByText('web-server')).toBeInTheDocument()
      expect(screen.getByText('db-server')).toBeInTheDocument()
      expect(screen.getByText('ACTIVE')).toBeInTheDocument()
      expect(screen.getByText('SHUTOFF')).toBeInTheDocument()
    })
  })

  it('서버 생성 버튼이 있다', async () => {
    vi.mocked(api.get).mockResolvedValue({ data: { items: [] } })

    render(<ComputePage />)

    await waitFor(() => {
      expect(screen.getByText('서버 생성')).toBeInTheDocument()
    })
  })

  it('서버 생성 버튼 클릭 시 모달이 열린다', async () => {
    vi.mocked(api.get).mockResolvedValue({
      data: { items: [] },
    })

    render(<ComputePage />)

    // 로딩 완료 대기
    await waitFor(() => {
      expect(screen.queryByText('로딩 중...')).not.toBeInTheDocument()
    })

    fireEvent.click(screen.getByText('서버 생성'))

    await waitFor(() => {
      expect(screen.getByText('서버 이름')).toBeInTheDocument()
    }, { timeout: 3000 })
  })

  it('로딩 중 상태를 표시한다', () => {
    vi.mocked(api.get).mockImplementation(() => new Promise(() => {}))
    render(<ComputePage />)
    expect(screen.getByText('로딩 중...')).toBeInTheDocument()
  })
})
