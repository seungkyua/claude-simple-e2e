'use client'

import { useState, useEffect, useCallback } from 'react'
import AuthGuard from '@/components/layout/AuthGuard'
import Sidebar from '@/components/layout/Sidebar'
import Header from '@/components/layout/Header'
import DataTable from '@/components/resource/DataTable'
import ServerCreateModal from '@/components/compute/ServerCreateModal'
import ServerDetailPanel from '@/components/compute/ServerDetailPanel'
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
  const [showCreate, setShowCreate] = useState(false)
  const [selectedServerId, setSelectedServerId] = useState<string | null>(null)

  const fetchServers = useCallback(() => {
    setLoading(true)
    api.get('/compute/servers')
      .then((res) => setServers(res.data.items || []))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    fetchServers()
  }, [fetchServers])

  const handleCreated = () => {
    setShowCreate(false)
    fetchServers()
  }

  const handleRowClick = (server: Server) => {
    setSelectedServerId(server.id)
  }

  const handleDetailAction = () => {
    fetchServers()
  }

  return (
    <AuthGuard>
      <div className="flex h-screen">
        <Sidebar />
        <div className="flex flex-1 flex-col">
          <Header />
          <main className="flex-1 overflow-y-auto p-6">
            <div className="mb-4 flex items-center justify-between">
              <h2 className="text-xl font-bold">Compute 관리</h2>
              <button
                onClick={() => setShowCreate(true)}
                className="rounded bg-blue-600 px-4 py-2 text-sm font-medium hover:bg-blue-700"
              >
                서버 생성
              </button>
            </div>
            {loading ? (
              <p className="text-gray-400">로딩 중...</p>
            ) : (
              <DataTable
                columns={columns}
                data={servers}
                total={servers.length}
                onRowClick={handleRowClick}
              />
            )}
          </main>
        </div>
      </div>

      {showCreate && (
        <ServerCreateModal
          onClose={() => setShowCreate(false)}
          onCreated={handleCreated}
        />
      )}

      {selectedServerId && (
        <ServerDetailPanel
          serverId={selectedServerId}
          onClose={() => setSelectedServerId(null)}
          onAction={handleDetailAction}
        />
      )}
    </AuthGuard>
  )
}
