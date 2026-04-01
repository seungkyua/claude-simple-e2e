'use client'

import { useState } from 'react'

interface Column<T> {
  key: string
  label: string
  render?: (item: T) => React.ReactNode
}

interface DataTableProps<T> {
  columns: Column<T>[]
  data: T[]
  total?: number
  page?: number
  pageSize?: number
  onPageChange?: (page: number) => void
  onDelete?: (item: T) => void
  idKey?: string
}

export default function DataTable<T extends Record<string, unknown>>({
  columns,
  data,
  total = 0,
  page = 1,
  pageSize = 20,
  onPageChange,
  onDelete,
  idKey = 'id',
}: DataTableProps<T>) {
  const [search, setSearch] = useState('')
  const totalPages = Math.ceil(total / pageSize)

  return (
    <div className="space-y-4">
      <input
        type="text"
        placeholder="검색..."
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        className="w-full max-w-sm rounded border border-gray-700 bg-gray-800 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
      />
      <div className="overflow-x-auto rounded border border-gray-800">
        <table className="w-full text-left text-sm">
          <thead className="border-b border-gray-800 bg-gray-900 text-gray-400">
            <tr>
              {columns.map((col) => (
                <th key={col.key} className="px-4 py-3 font-medium">{col.label}</th>
              ))}
              {onDelete && <th className="px-4 py-3 font-medium">작업</th>}
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-800">
            {data.map((item, i) => (
              <tr key={String(item[idKey] ?? i)} className="hover:bg-gray-900">
                {columns.map((col) => (
                  <td key={col.key} className="px-4 py-3">
                    {col.render ? col.render(item) : String(item[col.key] ?? '')}
                  </td>
                ))}
                {onDelete && (
                  <td className="px-4 py-3">
                    <button
                      onClick={() => onDelete(item)}
                      className="text-red-400 hover:text-red-300"
                    >
                      삭제
                    </button>
                  </td>
                )}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {totalPages > 1 && (
        <div className="flex items-center gap-2">
          <button
            onClick={() => onPageChange?.(page - 1)}
            disabled={page <= 1}
            className="rounded bg-gray-800 px-3 py-1 text-sm disabled:opacity-50"
          >
            이전
          </button>
          <span className="text-sm text-gray-400">
            {page} / {totalPages} (총 {total}건)
          </span>
          <button
            onClick={() => onPageChange?.(page + 1)}
            disabled={page >= totalPages}
            className="rounded bg-gray-800 px-3 py-1 text-sm disabled:opacity-50"
          >
            다음
          </button>
        </div>
      )}
    </div>
  )
}
