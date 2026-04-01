'use client'

import { useState, useEffect } from 'react'
import AuthGuard from '@/components/layout/AuthGuard'
import Sidebar from '@/components/layout/Sidebar'
import Header from '@/components/layout/Header'
import DataTable from '@/components/resource/DataTable'
import api from '@/services/api'

interface Server {
  id: string
  name: string
  status: string
  flavor: { id: string; name?: string }
  image: { id: string; name?: string }
  networks?: string
  [key: string]: unknown
}

const columns = [
  { key: 'id', label: 'ID', render: (s: Server) => s.id.slice(0, 8) + '...' },
  { key: 'name', label: 'Name' },
  { key: 'status', label: 'Status', render: (s: Server) => (
    <span className={s.status === 'ACTIVE' ? 'text-green-400' : s.status === 'ERROR' ? 'text-red-400' : 'text-yellow-400'}>
      {s.status}
    </span>
  )},
  { key: 'networks', label: 'Networks', render: (s: Server) => s.networks || '-' },
  { key: 'image', label: 'Image', render: (s: Server) => s.image?.name || s.image?.id || '-' },
  { key: 'flavor', label: 'Flavor', render: (s: Server) => s.flavor?.name || s.flavor?.id || '-' },
]

export default function ComputePage() {
  const [servers, setServers] = useState<Server[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.get('/compute/servers')
      .then((res) => setServers(res.data.items || []))
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
            <div className="mb-4 flex items-center justify-between">
              <h2 className="text-xl font-bold">Compute 관리</h2>
            </div>
            {loading ? (
              <p className="text-gray-400">로딩 중...</p>
            ) : (
              <DataTable columns={columns} data={servers} total={servers.length} />
            )}
          </main>
        </div>
      </div>
    </AuthGuard>
  )
}
