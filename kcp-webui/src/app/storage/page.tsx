'use client'

import { useState, useEffect } from 'react'
import AuthGuard from '@/components/layout/AuthGuard'
import Sidebar from '@/components/layout/Sidebar'
import Header from '@/components/layout/Header'
import DataTable from '@/components/resource/DataTable'
import api from '@/services/api'

interface Volume {
  id: string
  name: string
  status: string
  size: number
  volume_type: string
  [key: string]: unknown
}

const columns = [
  { key: 'id', label: 'ID', render: (v: Volume) => v.id.slice(0, 8) + '...' },
  { key: 'name', label: 'Name' },
  { key: 'status', label: 'Status' },
  { key: 'size', label: 'Size (GB)' },
  { key: 'volume_type', label: 'Type' },
]

export default function StoragePage() {
  const [volumes, setVolumes] = useState<Volume[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    api.get('/storage/volumes')
      .then((res) => setVolumes(res.data.items || []))
      .catch((err) => {
        const msg = err.response?.data?.error?.message || 'Storage 서비스를 사용할 수 없습니다'
        setError(msg)
      })
      .finally(() => setLoading(false))
  }, [])

  return (
    <AuthGuard>
      <div className="flex h-screen">
        <Sidebar />
        <div className="flex flex-1 flex-col">
          <Header />
          <main className="flex-1 overflow-y-auto p-6">
            <h2 className="mb-4 text-xl font-bold">Storage 관리</h2>
            {loading ? <p className="text-gray-400">로딩 중...</p> :
             error ? <p className="text-yellow-400">{error}</p> : (
              <DataTable columns={columns} data={volumes} total={volumes.length} />
            )}
          </main>
        </div>
      </div>
    </AuthGuard>
  )
}
