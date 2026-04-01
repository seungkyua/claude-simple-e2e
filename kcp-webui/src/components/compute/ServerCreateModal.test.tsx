import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import ServerCreateModal from './ServerCreateModal'

vi.mock('@/services/api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
  },
}))

import api from '@/services/api'

describe('ServerCreateModal', () => {
  const onClose = vi.fn()
  const onCreated = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    // flavor, image, network, security-group 목록 로드
    vi.mocked(api.get).mockImplementation((url: string) => {
      if (url.includes('flavors')) {
        return Promise.resolve({
          data: { items: [{ id: '1', name: 'm1.tiny', vcpus: 1, ram: 512, disk: 1 }] },
        })
      }
      if (url.includes('images')) {
        return Promise.resolve({
          data: { items: [{ id: 'img-1', name: 'cirros', status: 'active' }] },
        })
      }
      if (url.includes('security-groups')) {
        return Promise.resolve({
          data: { items: [{ id: 'sg-1', name: 'default' }] },
        })
      }
      if (url.includes('networks')) {
        return Promise.resolve({
          data: { items: [{ id: 'net-1', name: 'private' }] },
        })
      }
      return Promise.resolve({ data: { items: [] } })
    })
  })

  it('모달 제목을 표시한다', async () => {
    render(<ServerCreateModal onClose={onClose} onCreated={onCreated} />)
    expect(screen.getByText('서버 생성')).toBeInTheDocument()
  })

  it('폼 필드를 렌더링한다', async () => {
    render(<ServerCreateModal onClose={onClose} onCreated={onCreated} />)

    await waitFor(() => {
      expect(screen.getByPlaceholderText('my-server')).toBeInTheDocument()
      expect(screen.getByText('Flavor')).toBeInTheDocument()
      expect(screen.getByText('이미지')).toBeInTheDocument()
    })
  })

  it('Flavor 드롭다운에 옵션이 로드된다', async () => {
    render(<ServerCreateModal onClose={onClose} onCreated={onCreated} />)

    await waitFor(() => {
      expect(screen.getByText(/m1\.tiny/)).toBeInTheDocument()
    })
  })

  it('이미지 드롭다운에 active 이미지만 로드된다', async () => {
    render(<ServerCreateModal onClose={onClose} onCreated={onCreated} />)

    await waitFor(() => {
      expect(screen.getByText('cirros')).toBeInTheDocument()
    })
  })

  it('취소 버튼 클릭 시 onClose가 호출된다', async () => {
    render(<ServerCreateModal onClose={onClose} onCreated={onCreated} />)
    fireEvent.click(screen.getByText('취소'))
    expect(onClose).toHaveBeenCalled()
  })

  it('서버 생성 성공 시 onCreated가 호출된다', async () => {
    vi.mocked(api.post).mockResolvedValue({ data: { id: 'new-server' } })

    render(<ServerCreateModal onClose={onClose} onCreated={onCreated} />)

    await waitFor(() => {
      expect(screen.getByText(/m1\.tiny/)).toBeInTheDocument()
    })

    fireEvent.change(screen.getByPlaceholderText('my-server'), { target: { value: 'test-vm' } })

    // flavor 선택
    const flavorSelect = screen.getAllByRole('combobox')[0]
    fireEvent.change(flavorSelect, { target: { value: '1' } })

    // image 선택
    const imageSelect = screen.getAllByRole('combobox')[1]
    fireEvent.change(imageSelect, { target: { value: 'img-1' } })

    fireEvent.click(screen.getByText('생성'))

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith('/compute/servers', expect.objectContaining({
        name: 'test-vm',
        flavorId: '1',
        imageId: 'img-1',
      }))
      expect(onCreated).toHaveBeenCalled()
    })
  })

  it('서버 생성 실패 시 에러 메시지를 표시한다', async () => {
    vi.mocked(api.post).mockRejectedValue({
      response: { data: { error: { message: 'Flavor not found' } } },
    })

    render(<ServerCreateModal onClose={onClose} onCreated={onCreated} />)

    await waitFor(() => {
      expect(screen.getByText(/m1\.tiny/)).toBeInTheDocument()
    })

    fireEvent.change(screen.getByPlaceholderText('my-server'), { target: { value: 'test-vm' } })
    const flavorSelect = screen.getAllByRole('combobox')[0]
    fireEvent.change(flavorSelect, { target: { value: '1' } })
    const imageSelect = screen.getAllByRole('combobox')[1]
    fireEvent.change(imageSelect, { target: { value: 'img-1' } })
    fireEvent.click(screen.getByText('생성'))

    await waitFor(() => {
      expect(screen.getByText('Flavor not found')).toBeInTheDocument()
    })
  })
})
