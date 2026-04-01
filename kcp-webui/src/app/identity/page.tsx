'use client'

import { useState, useEffect } from 'react'
import AuthGuard from '@/components/layout/AuthGuard'
import Sidebar from '@/components/layout/Sidebar'
import Header from '@/components/layout/Header'
import DataTable from '@/components/resource/DataTable'
import api from '@/services/api'

interface Project {
  id: string
  name: string
  domain_id: string
  description: string
  enabled: boolean
  [key: string]: unknown
}

const columns = [
  { key: 'id', label: 'ID', render: (p: Project) => p.id.slice(0, 8) + '...' },
  { key: 'name', label: 'Name' },
  { key: 'domain_id', label: 'Domain ID' },
  { key: 'description', label: 'Description' },
  { key: 'enabled', label: 'Enabled', render: (p: Project) => String(p.enabled) },
]

export default function IdentityPage() {
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.get('/identity/projects')
      .then((res) => setProjects(res.data.items || []))
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
            <h2 className="mb-4 text-xl font-bold">Identity 관리</h2>
            {loading ? <p className="text-gray-400">로딩 중...</p> : (
              <DataTable columns={columns} data={projects} total={projects.length} />
            )}
          </main>
        </div>
      </div>
    </AuthGuard>
  )
}
