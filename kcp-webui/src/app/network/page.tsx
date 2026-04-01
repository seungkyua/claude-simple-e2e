'use client'

import { useState, useEffect } from 'react'
import AuthGuard from '@/components/layout/AuthGuard'
import Sidebar from '@/components/layout/Sidebar'
import Header from '@/components/layout/Header'
import DataTable from '@/components/resource/DataTable'
import api from '@/services/api'

interface Network {
  id: string
  name: string
  status: string
  subnets: string[]
  shared: boolean
  'router:external': boolean
  [key: string]: unknown
}

const columns = [
  { key: 'id', label: 'ID', render: (n: Network) => n.id.slice(0, 8) + '...' },
  { key: 'name', label: 'Name' },
  { key: 'status', label: 'Status' },
  { key: 'subnets', label: 'Subnets', render: (n: Network) => (n.subnets || []).length.toString() },
  { key: 'shared', label: 'Shared', render: (n: Network) => String(n.shared) },
  { key: 'router:external', label: 'External', render: (n: Network) => String(n['router:external']) },
]

export default function NetworkPage() {
  const [networks, setNetworks] = useState<Network[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.get('/network/networks')
      .then((res) => setNetworks(res.data.items || []))
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
            <h2 className="mb-4 text-xl font-bold">Network 관리</h2>
            {loading ? <p className="text-gray-400">로딩 중...</p> : (
              <DataTable columns={columns} data={networks} total={networks.length} />
            )}
          </main>
        </div>
      </div>
    </AuthGuard>
  )
}
