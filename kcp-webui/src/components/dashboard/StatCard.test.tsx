import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import StatCard from './StatCard'

describe('StatCard', () => {
  it('제목과 값을 표시한다', () => {
    render(<StatCard title="VM 인스턴스" value={5} />)
    expect(screen.getByText('VM 인스턴스')).toBeInTheDocument()
    expect(screen.getByText('5')).toBeInTheDocument()
  })

  it('subtitle이 있으면 표시한다', () => {
    render(<StatCard title="볼륨" value={10} subtitle="100 GB" />)
    expect(screen.getByText('100 GB')).toBeInTheDocument()
  })

  it('subtitle이 없으면 표시하지 않는다', () => {
    render(<StatCard title="네트워크" value={3} />)
    expect(screen.queryByText('GB')).not.toBeInTheDocument()
  })

  it('문자열 값도 표시할 수 있다', () => {
    render(<StatCard title="상태" value="정상" />)
    expect(screen.getByText('정상')).toBeInTheDocument()
  })
})
