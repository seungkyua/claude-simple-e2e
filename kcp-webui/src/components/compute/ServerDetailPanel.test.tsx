import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import ServerDetailPanel from './ServerDetailPanel'

vi.mock('@/services/api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    delete: vi.fn(),
  },
}))

import api from '@/services/api'

const mockServer = {
  id: 'aaaa-1111-2222-3333',
  name: 'web-server',
  status: 'ACTIVE',
  flavor: { id: '1', name: 'm1.tiny' },
  image: { id: 'img-1', name: 'ubuntu-22' },
  networks: 'private=10.0.0.5',
  tenant_id: 'proj-1',
  user_id: 'user-1',
  key_name: 'mykey',
  'OS-EXT-AZ:availability_zone': 'nova',
  'OS-EXT-STS:power_state': 1,
  'OS-EXT-STS:vm_state': 'active',
  'OS-EXT-SRV-ATTR:host': 'compute-1',
  security_groups: [{ name: 'default' }],
  created: '2026-04-01T10:00:00Z',
  updated: '2026-04-01T10:05:00Z',
}

describe('ServerDetailPanel', () => {
  const onClose = vi.fn()
  const onAction = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('서버 상세 정보를 표시한다', async () => {
    vi.mocked(api.get).mockResolvedValue({ data: mockServer })

    render(<ServerDetailPanel serverId="aaaa-1111-2222-3333" onClose={onClose} onAction={onAction} />)

    await waitFor(() => {
      expect(screen.getByText('서버 상세')).toBeInTheDocument()
      expect(screen.getByText('web-server')).toBeInTheDocument()
      expect(screen.getByText('ACTIVE')).toBeInTheDocument()
      expect(screen.getByText('m1.tiny')).toBeInTheDocument()
      expect(screen.getByText('Running')).toBeInTheDocument()
    })
  })

  it('ACTIVE 상태에서 중지/재부팅 버튼을 표시한다', async () => {
    vi.mocked(api.get).mockResolvedValue({ data: mockServer })

    render(<ServerDetailPanel serverId="aaaa-1111-2222-3333" onClose={onClose} onAction={onAction} />)

    await waitFor(() => {
      expect(screen.getByText('중지')).toBeInTheDocument()
      expect(screen.getByText('재부팅')).toBeInTheDocument()
      expect(screen.queryByText('시작')).not.toBeInTheDocument()
    })
  })

  it('SHUTOFF 상태에서 시작 버튼을 표시한다', async () => {
    vi.mocked(api.get).mockResolvedValue({
      data: { ...mockServer, status: 'SHUTOFF', 'OS-EXT-STS:power_state': 4 },
    })

    render(<ServerDetailPanel serverId="aaaa-1111-2222-3333" onClose={onClose} onAction={onAction} />)

    await waitFor(() => {
      expect(screen.getByText('시작')).toBeInTheDocument()
      expect(screen.queryByText('중지')).not.toBeInTheDocument()
    })
  })

  it('삭제 버튼이 항상 표시된다', async () => {
    vi.mocked(api.get).mockResolvedValue({ data: mockServer })

    render(<ServerDetailPanel serverId="aaaa-1111-2222-3333" onClose={onClose} onAction={onAction} />)

    await waitFor(() => {
      expect(screen.getByText('삭제')).toBeInTheDocument()
    })
  })

  it('닫기 버튼 클릭 시 onClose가 호출된다', async () => {
    vi.mocked(api.get).mockResolvedValue({ data: mockServer })

    render(<ServerDetailPanel serverId="aaaa-1111-2222-3333" onClose={onClose} onAction={onAction} />)

    await waitFor(() => {
      fireEvent.click(screen.getByText('X'))
    })
    expect(onClose).toHaveBeenCalled()
  })

  it('로딩 중 상태를 표시한다', () => {
    vi.mocked(api.get).mockImplementation(() => new Promise(() => {}))

    render(<ServerDetailPanel serverId="aaaa-1111-2222-3333" onClose={onClose} onAction={onAction} />)

    expect(screen.getByText('로딩 중...')).toBeInTheDocument()
  })

  it('API 실패 시 에러 메시지를 표시한다', async () => {
    vi.mocked(api.get).mockRejectedValue(new Error('Network error'))

    render(<ServerDetailPanel serverId="aaaa-1111-2222-3333" onClose={onClose} onAction={onAction} />)

    await waitFor(() => {
      expect(screen.getByText('서버 정보를 불러올 수 없습니다')).toBeInTheDocument()
    })
  })
})
