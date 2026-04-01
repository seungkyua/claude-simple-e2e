import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import Sidebar from './Sidebar'

// Next.js 라우터 모킹
vi.mock('next/navigation', () => ({
  usePathname: () => '/dashboard',
}))

describe('Sidebar', () => {
  it('모든 네비게이션 항목을 렌더링한다', () => {
    render(<Sidebar />)
    expect(screen.getByText('대시보드')).toBeInTheDocument()
    expect(screen.getByText('Compute')).toBeInTheDocument()
    expect(screen.getByText('Network')).toBeInTheDocument()
    expect(screen.getByText('Storage')).toBeInTheDocument()
    expect(screen.getByText('Identity')).toBeInTheDocument()
    expect(screen.getByText('Image')).toBeInTheDocument()
    expect(screen.getByText('감사 로그')).toBeInTheDocument()
  })

  it('KCP 타이틀을 표시한다', () => {
    render(<Sidebar />)
    expect(screen.getByText('KCP')).toBeInTheDocument()
  })
})
