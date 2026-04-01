'use client'

import { useState } from 'react'
import AuthGuard from '@/components/layout/AuthGuard'
import Sidebar from '@/components/layout/Sidebar'
import Header from '@/components/layout/Header'
import DataTable from '@/components/resource/DataTable'
import type { Image } from '@/types/image'

const columns = [
  { key: 'name', label: '이름' },
  { key: 'status', label: '상태' },
  { key: 'disk_format', label: '디스크 형식' },
  { key: 'size', label: '크기', render: (img: Image) => img.size ? `${(img.size / 1024 / 1024).toFixed(0)} MB` : '-' },
  { key: 'id', label: 'ID', render: (img: Image) => img.id.slice(0, 8) + '...' },
]

export default function ImagePage() {
  const [images] = useState<Image[]>([])

  return (
    <AuthGuard>
      <div className="flex h-screen">
        <Sidebar />
        <div className="flex flex-1 flex-col">
          <Header />
          <main className="flex-1 overflow-y-auto p-6">
            <div className="mb-4 flex items-center justify-between">
              <h2 className="text-xl font-bold">Image 관리</h2>
              <button className="rounded bg-blue-600 px-4 py-2 text-sm hover:bg-blue-700">
                이미지 업로드
              </button>
            </div>
            <DataTable columns={columns} data={images} />
          </main>
        </div>
      </div>
    </AuthGuard>
  )
}
