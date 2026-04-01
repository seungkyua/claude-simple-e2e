'use client'

import { useState, useEffect } from 'react'
import AuthGuard from '@/components/layout/AuthGuard'
import Sidebar from '@/components/layout/Sidebar'
import Header from '@/components/layout/Header'
import DataTable from '@/components/resource/DataTable'
import type { Server } from '@/types/compute'

const columns = [
  { key: 'name', label: '이름' },
  { key: 'status', label: '상태', render: (s: Server) => (
    <span className={s.status === 'ACTIVE' ? 'text-green-400' : s.status === 'ERROR' ? 'text-red-400' : 'text-yellow-400'}>
      {s.status}
    </span>
  )},
  { key: 'flavor', label: 'Flavor', render: (s: Server) => s.flavor?.name || '-' },
  { key: 'id', label: 'ID', render: (s: Server) => s.id.slice(0, 8) + '...' },
]

export default function ComputePage() {
  const [servers, setServers] = useState<Server[]>([])

  useEffect(() => {
    // TODO: API 호출
  }, [])

  return (
    <AuthGuard>
      <div className="flex h-screen">
        <Sidebar />
        <div className="flex flex-1 flex-col">
          <Header />
          <main className="flex-1 overflow-y-auto p-6">
            <div className="mb-4 flex items-center justify-between">
              <h2 className="text-xl font-bold">Compute 관리</h2>
              <button className="rounded bg-blue-600 px-4 py-2 text-sm hover:bg-blue-700">
                VM 생성
              </button>
            </div>
            <DataTable columns={columns} data={servers} />
          </main>
        </div>
      </div>
    </AuthGuard>
  )
}
