'use client'

import { useState } from 'react'
import AuthGuard from '@/components/layout/AuthGuard'
import Sidebar from '@/components/layout/Sidebar'
import Header from '@/components/layout/Header'
import DataTable from '@/components/resource/DataTable'
import type { Project } from '@/types/identity'

const columns = [
  { key: 'name', label: '이름' },
  { key: 'description', label: '설명' },
  { key: 'enabled', label: '활성', render: (p: Project) => p.enabled ? '예' : '아니오' },
  { key: 'id', label: 'ID', render: (p: Project) => p.id.slice(0, 8) + '...' },
]

export default function IdentityPage() {
  const [projects] = useState<Project[]>([])

  return (
    <AuthGuard>
      <div className="flex h-screen">
        <Sidebar />
        <div className="flex flex-1 flex-col">
          <Header />
          <main className="flex-1 overflow-y-auto p-6">
            <div className="mb-4 flex items-center justify-between">
              <h2 className="text-xl font-bold">Identity 관리</h2>
              <button className="rounded bg-blue-600 px-4 py-2 text-sm hover:bg-blue-700">
                프로젝트 생성
              </button>
            </div>
            <DataTable columns={columns} data={projects} />
          </main>
        </div>
      </div>
    </AuthGuard>
  )
}
