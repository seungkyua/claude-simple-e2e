'use client'

import { useState } from 'react'
import AuthGuard from '@/components/layout/AuthGuard'
import Sidebar from '@/components/layout/Sidebar'
import Header from '@/components/layout/Header'
import DataTable from '@/components/resource/DataTable'
import type { AuditLog } from '@/types/audit'

const columns = [
  { key: 'createdAt', label: '시각' },
  { key: 'user', label: '사용자', render: (log: AuditLog) => log.user?.username || '-' },
  { key: 'action', label: '작업' },
  { key: 'resourceType', label: '리소스' },
  { key: 'source', label: '출처' },
  { key: 'statusCode', label: '상태', render: (log: AuditLog) => (
    <span className={log.statusCode < 400 ? 'text-green-400' : 'text-red-400'}>
      {log.statusCode}
    </span>
  )},
]

export default function AuditPage() {
  const [logs] = useState<AuditLog[]>([])

  return (
    <AuthGuard>
      <div className="flex h-screen">
        <Sidebar />
        <div className="flex flex-1 flex-col">
          <Header />
          <main className="flex-1 overflow-y-auto p-6">
            <h2 className="mb-4 text-xl font-bold">감사 로그</h2>
            <div className="mb-4 flex gap-2">
              <select className="rounded border border-gray-700 bg-gray-800 px-3 py-2 text-sm">
                <option value="">작업 유형</option>
                <option value="CREATE">CREATE</option>
                <option value="READ">READ</option>
                <option value="UPDATE">UPDATE</option>
                <option value="DELETE">DELETE</option>
              </select>
              <select className="rounded border border-gray-700 bg-gray-800 px-3 py-2 text-sm">
                <option value="">리소스 유형</option>
                <option value="VM">VM</option>
                <option value="NETWORK">NETWORK</option>
                <option value="VOLUME">VOLUME</option>
              </select>
            </div>
            <DataTable columns={columns} data={logs} />
          </main>
        </div>
      </div>
    </AuthGuard>
  )
}
