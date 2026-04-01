'use client'

import { useAuthStore } from '@/stores/authStore'
import { useRouter } from 'next/navigation'

export default function Header() {
  const { username, logout } = useAuthStore()
  const router = useRouter()

  const handleLogout = () => {
    logout()
    router.push('/login')
  }

  return (
    <header className="flex h-14 items-center justify-between border-b border-gray-800 bg-gray-950 px-6">
      <div />
      <div className="flex items-center gap-4">
        <span className="text-sm text-gray-400">{username}</span>
        <button
          onClick={handleLogout}
          className="rounded bg-gray-800 px-3 py-1 text-sm text-gray-300 hover:bg-gray-700"
        >
          로그아웃
        </button>
      </div>
    </header>
  )
}
