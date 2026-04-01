'use client'

import AuthGuard from '@/components/layout/AuthGuard'
import Sidebar from '@/components/layout/Sidebar'
import Header from '@/components/layout/Header'
import StatCard from '@/components/dashboard/StatCard'

export default function DashboardPage() {
  return (
    <AuthGuard>
      <div className="flex h-screen">
        <Sidebar />
        <div className="flex flex-1 flex-col">
          <Header />
          <main className="flex-1 overflow-y-auto p-6">
            <h2 className="mb-6 text-xl font-bold">대시보드</h2>
            <div className="grid grid-cols-4 gap-4">
              <StatCard title="VM 인스턴스" value={0} subtitle="Active" color="blue" />
              <StatCard title="네트워크" value={0} color="green" />
              <StatCard title="볼륨" value={0} subtitle="0 GB" color="yellow" />
              <StatCard title="이미지" value={0} color="blue" />
            </div>
            <div className="mt-6 grid grid-cols-2 gap-4">
              <StatCard title="프로젝트" value={0} color="green" />
              <StatCard title="사용자" value={0} color="blue" />
            </div>
          </main>
        </div>
      </div>
    </AuthGuard>
  )
}
