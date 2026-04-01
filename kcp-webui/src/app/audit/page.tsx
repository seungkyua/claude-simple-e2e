'use client'

import { useState, useEffect } from 'react'
import AuthGuard from '@/components/layout/AuthGuard'
import Sidebar from '@/components/layout/Sidebar'
import Header from '@/components/layout/Header'
import DataTable from '@/components/resource/DataTable'
import api from '@/services/api'

interface AuditLog {
  id: string
  user_id: string
  username: string
  action: string
  resource_type: string
  resource_id: string
  source: string
  status_code: number
  ip_address: string
  created_at: string
  [key: string]: unknown
}

const columns = [
  { key: 'created_at', label: 'Time', render: (log: AuditLog) => {
    if (!log.created_at) return '-'
    return new Date(log.created_at).toLocaleString('ko-KR')
  }},
  { key: 'username', label: 'User', render: (log: AuditLog) => log.username || log.user_id?.slice(0, 8) || '-' },
  { key: 'action', label: 'Action' },
  { key: 'resource_type', label: 'Resource' },
  { key: 'source', label: 'Source' },
  { key: 'status_code', label: 'Status', render: (log: AuditLog) => (
    <span className={log.status_code < 400 ? 'text-green-400' : 'text-red-400'}>
      {log.status_code}
    </span>
  )},
  { key: 'ip_address', label: 'IP' },
]

export default function AuditPage() {
  const [logs, setLogs] = useState<AuditLog[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.get('/audit/logs')
      .then((res) => setLogs(res.data.items || []))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  return (
    <AuthGuard>
      <div className="flex h-screen">
        <Sidebar />
        <div className="flex flex-1 flex-col">
          <Header />
          <main className="flex-1 overflow-y-auto p-6">
            <h2 className="mb-4 text-xl font-bold">감사 로그</h2>
            {loading ? <p className="text-gray-400">로딩 중...</p> : (
              <DataTable columns={columns} data={logs} total={logs.length} />
            )}
          </main>
        </div>
      </div>
    </AuthGuard>
  )
}
