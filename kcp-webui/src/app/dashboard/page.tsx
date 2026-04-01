'use client'

import { useState, useEffect } from 'react'
import AuthGuard from '@/components/layout/AuthGuard'
import Sidebar from '@/components/layout/Sidebar'
import Header from '@/components/layout/Header'
import StatCard from '@/components/dashboard/StatCard'
import api from '@/services/api'

interface DashboardStats {
  stats: Record<string, number>
  warnings?: string[]
}

export default function DashboardPage() {
  const [stats, setStats] = useState<Record<string, number>>({})
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.get<DashboardStats>('/stats/dashboard')
      .then((res) => setStats(res.data.stats || {}))
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
            <h2 className="mb-6 text-xl font-bold">대시보드</h2>
            {loading ? (
              <p className="text-gray-400">로딩 중...</p>
            ) : (
              <>
                <div className="grid grid-cols-4 gap-4">
                  <StatCard title="VM 인스턴스" value={stats.servers ?? 0} color="blue" />
                  <StatCard title="네트워크" value={stats.networks ?? 0} color="green" />
                  <StatCard title="서브넷" value={stats.subnets ?? 0} color="green" />
                  <StatCard title="이미지" value={stats.images ?? 0} color="blue" />
                </div>
                <div className="mt-4 grid grid-cols-4 gap-4">
                  <StatCard title="라우터" value={stats.routers ?? 0} color="yellow" />
                  <StatCard title="보안그룹" value={stats.security_groups ?? 0} color="yellow" />
                  <StatCard title="프로젝트" value={stats.projects ?? 0} color="green" />
                  <StatCard title="사용자" value={stats.users ?? 0} color="blue" />
                </div>
              </>
            )}
          </main>
        </div>
      </div>
    </AuthGuard>
  )
}
