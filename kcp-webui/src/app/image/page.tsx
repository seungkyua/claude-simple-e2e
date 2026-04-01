'use client'

import { useState, useEffect } from 'react'
import AuthGuard from '@/components/layout/AuthGuard'
import Sidebar from '@/components/layout/Sidebar'
import Header from '@/components/layout/Header'
import DataTable from '@/components/resource/DataTable'
import api from '@/services/api'

interface Image {
  id: string
  name: string
  status: string
  disk_format: string
  size: number
  visibility: string
  [key: string]: unknown
}

const columns = [
  { key: 'id', label: 'ID', render: (img: Image) => img.id.slice(0, 8) + '...' },
  { key: 'name', label: 'Name' },
  { key: 'status', label: 'Status', render: (img: Image) => (
    <span className={img.status === 'active' ? 'text-green-400' : 'text-yellow-400'}>
      {img.status}
    </span>
  )},
  { key: 'disk_format', label: 'Disk Format' },
  { key: 'size', label: 'Size', render: (img: Image) => img.size ? `${(img.size / 1024 / 1024).toFixed(0)} MB` : '-' },
  { key: 'visibility', label: 'Visibility' },
]

export default function ImagePage() {
  const [images, setImages] = useState<Image[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.get('/image/images')
      .then((res) => setImages(res.data.items || []))
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
            <h2 className="mb-4 text-xl font-bold">Image 관리</h2>
            {loading ? <p className="text-gray-400">로딩 중...</p> : (
              <DataTable columns={columns} data={images} total={images.length} />
            )}
          </main>
        </div>
      </div>
    </AuthGuard>
  )
}
