'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuthStore } from '@/stores/authStore'
import api from '@/services/api'

export default function LoginPage() {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const router = useRouter()
  const { login } = useAuthStore()

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      const res = await api.post('/auth/login', {
        username,
        password,
        authType: 'JWT',
      })
      login(res.data.token, res.data.user.username)
      router.push('/dashboard')
    } catch (err: unknown) {
      const axiosErr = err as { response?: { data?: { error?: { message?: string } } } }
      setError(axiosErr.response?.data?.error?.message || '로그인에 실패했습니다')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center">
      <form onSubmit={handleLogin} className="w-full max-w-sm space-y-4 rounded-lg border border-gray-800 bg-gray-900 p-8">
        <h1 className="text-center text-2xl font-bold">KCP 로그인</h1>
        {error && (
          <div className="rounded border border-red-800 bg-red-900/30 px-4 py-2 text-sm text-red-400">
            {error}
          </div>
        )}
        <input
          type="text"
          placeholder="사용자명"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          className="w-full rounded border border-gray-700 bg-gray-800 px-4 py-2 focus:border-blue-500 focus:outline-none"
        />
        <input
          type="password"
          placeholder="비밀번호"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          className="w-full rounded border border-gray-700 bg-gray-800 px-4 py-2 focus:border-blue-500 focus:outline-none"
        />
        <button
          type="submit"
          disabled={loading}
          className="w-full rounded bg-blue-600 py-2 font-medium hover:bg-blue-700 disabled:opacity-50"
        >
          {loading ? '로그인 중...' : '로그인'}
        </button>
      </form>
    </div>
  )
}
