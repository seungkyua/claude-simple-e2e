'use client'

import { useState, useEffect, useCallback, useRef } from 'react'
import AuthGuard from '@/components/layout/AuthGuard'
import Sidebar from '@/components/layout/Sidebar'
import Header from '@/components/layout/Header'
import DataTable from '@/components/resource/DataTable'
import ServerCreateModal from '@/components/compute/ServerCreateModal'
import ServerDetailPanel from '@/components/compute/ServerDetailPanel'
import api from '@/services/api'

/** 자동 갱신 기본 주기 (밀리초) */
const DEFAULT_REFRESH_INTERVAL = 5000

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

const REFRESH_OPTIONS = [
  { label: '끄기', value: 0 },
  { label: '3초', value: 3000 },
  { label: '5초', value: 5000 },
  { label: '10초', value: 10000 },
  { label: '30초', value: 30000 },
  { label: '60초', value: 60000 },
]

export default function ComputePage() {
  const [servers, setServers] = useState<Server[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreate, setShowCreate] = useState(false)
  const [selectedServerId, setSelectedServerId] = useState<string | null>(null)
  const [refreshInterval, setRefreshInterval] = useState(DEFAULT_REFRESH_INTERVAL)
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const fetchServers = useCallback(() => {
    api.get('/compute/servers')
      .then((res) => setServers(res.data.items || []))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  // 초기 로드
  useEffect(() => {
    setLoading(true)
    fetchServers()
  }, [fetchServers])

  // 자동 갱신 타이머
  useEffect(() => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current)
      intervalRef.current = null
    }
    if (refreshInterval > 0) {
      intervalRef.current = setInterval(fetchServers, refreshInterval)
    }
    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
      }
    }
  }, [refreshInterval, fetchServers])

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
              <div className="flex items-center gap-3">
                <label className="flex items-center gap-2 text-sm text-gray-400">
                  자동 갱신
                  <select
                    value={refreshInterval}
                    onChange={(e) => setRefreshInterval(Number(e.target.value))}
                    className="rounded border border-gray-700 bg-gray-800 px-2 py-1 text-sm focus:border-blue-500 focus:outline-none"
                  >
                    {REFRESH_OPTIONS.map((opt) => (
                      <option key={opt.value} value={opt.value}>{opt.label}</option>
                    ))}
                  </select>
                </label>
                <button
                  onClick={() => setShowCreate(true)}
                  className="rounded bg-blue-600 px-4 py-2 text-sm font-medium hover:bg-blue-700"
                >
                  서버 생성
                </button>
              </div>
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
