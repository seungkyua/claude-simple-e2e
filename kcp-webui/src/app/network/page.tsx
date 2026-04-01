'use client'

import { useState } from 'react'
import AuthGuard from '@/components/layout/AuthGuard'
import Sidebar from '@/components/layout/Sidebar'
import Header from '@/components/layout/Header'
import DataTable from '@/components/resource/DataTable'
import type { Network } from '@/types/network'

const columns = [
  { key: 'name', label: '이름' },
  { key: 'status', label: '상태' },
  { key: 'shared', label: '공유', render: (n: Network) => n.shared ? '예' : '아니오' },
  { key: 'id', label: 'ID', render: (n: Network) => n.id.slice(0, 8) + '...' },
]

export default function NetworkPage() {
  const [networks] = useState<Network[]>([])

  return (
    <AuthGuard>
      <div className="flex h-screen">
        <Sidebar />
        <div className="flex flex-1 flex-col">
          <Header />
          <main className="flex-1 overflow-y-auto p-6">
            <div className="mb-4 flex items-center justify-between">
              <h2 className="text-xl font-bold">Network 관리</h2>
              <button className="rounded bg-blue-600 px-4 py-2 text-sm hover:bg-blue-700">
                네트워크 생성
              </button>
            </div>
            <DataTable columns={columns} data={networks} />
          </main>
        </div>
      </div>
    </AuthGuard>
  )
}
