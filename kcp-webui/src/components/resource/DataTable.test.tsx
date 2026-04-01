import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import DataTable from './DataTable'

interface TestItem {
  id: string
  name: string
  status: string
  [key: string]: unknown
}

const columns = [
  { key: 'id', label: 'ID' },
  { key: 'name', label: 'Name' },
  { key: 'status', label: 'Status' },
]

const testData: TestItem[] = [
  { id: '1', name: 'server-1', status: 'ACTIVE' },
  { id: '2', name: 'server-2', status: 'SHUTOFF' },
]

describe('DataTable', () => {
  it('헤더를 렌더링한다', () => {
    render(<DataTable columns={columns} data={testData} />)
    expect(screen.getByText('ID')).toBeInTheDocument()
    expect(screen.getByText('Name')).toBeInTheDocument()
    expect(screen.getByText('Status')).toBeInTheDocument()
  })

  it('데이터 행을 렌더링한다', () => {
    render(<DataTable columns={columns} data={testData} />)
    expect(screen.getByText('server-1')).toBeInTheDocument()
    expect(screen.getByText('server-2')).toBeInTheDocument()
    expect(screen.getByText('ACTIVE')).toBeInTheDocument()
    expect(screen.getByText('SHUTOFF')).toBeInTheDocument()
  })

  it('빈 데이터일 때 행이 없다', () => {
    render(<DataTable columns={columns} data={[]} />)
    expect(screen.getByText('ID')).toBeInTheDocument()
    expect(screen.queryByText('server-1')).not.toBeInTheDocument()
  })

  it('행 클릭 시 onRowClick이 호출된다', () => {
    const handleRowClick = vi.fn()
    render(<DataTable columns={columns} data={testData} onRowClick={handleRowClick} />)
    fireEvent.click(screen.getByText('server-1'))
    expect(handleRowClick).toHaveBeenCalledWith(testData[0])
  })

  it('삭제 버튼이 onDelete와 함께 렌더링된다', () => {
    const handleDelete = vi.fn()
    render(<DataTable columns={columns} data={testData} onDelete={handleDelete} />)
    const deleteButtons = screen.getAllByText('삭제')
    expect(deleteButtons).toHaveLength(2)
    fireEvent.click(deleteButtons[0])
    expect(handleDelete).toHaveBeenCalledWith(testData[0])
  })

  it('커스텀 렌더러가 적용된다', () => {
    const columnsWithRender = [
      ...columns,
      { key: 'status', label: 'Custom', render: (item: TestItem) => `[${item.status}]` },
    ]
    render(<DataTable columns={columnsWithRender} data={testData} />)
    expect(screen.getByText('[ACTIVE]')).toBeInTheDocument()
  })

  it('검색 입력 필드가 존재한다', () => {
    render(<DataTable columns={columns} data={testData} />)
    expect(screen.getByPlaceholderText('검색...')).toBeInTheDocument()
  })

  it('페이지네이션이 총 페이지 1일 때 표시되지 않는다', () => {
    render(<DataTable columns={columns} data={testData} total={2} pageSize={20} />)
    expect(screen.queryByText('이전')).not.toBeInTheDocument()
  })

  it('페이지네이션이 총 페이지 2 이상일 때 표시된다', () => {
    render(<DataTable columns={columns} data={testData} total={30} pageSize={20} />)
    expect(screen.getByText('이전')).toBeInTheDocument()
    expect(screen.getByText('다음')).toBeInTheDocument()
  })
})
