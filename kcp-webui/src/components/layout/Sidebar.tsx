'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'

const navItems = [
  { href: '/dashboard', label: '대시보드', icon: '📊' },
  { href: '/compute', label: 'Compute', icon: '🖥️' },
  { href: '/network', label: 'Network', icon: '🌐' },
  { href: '/storage', label: 'Storage', icon: '💾' },
  { href: '/identity', label: 'Identity', icon: '👤' },
  { href: '/image', label: 'Image', icon: '📀' },
  { href: '/audit', label: '감사 로그', icon: '📋' },
]

export default function Sidebar() {
  const pathname = usePathname()

  return (
    <aside className="flex h-screen w-56 flex-col border-r border-gray-800 bg-gray-950">
      <div className="border-b border-gray-800 px-4 py-4">
        <h1 className="text-lg font-bold text-white">KCP</h1>
        <p className="text-xs text-gray-500">OpenStack 관리 콘솔</p>
      </div>
      <nav className="flex-1 space-y-1 p-2">
        {navItems.map((item) => (
          <Link
            key={item.href}
            href={item.href}
            className={`flex items-center gap-3 rounded px-3 py-2 text-sm transition-colors ${
              pathname === item.href
                ? 'bg-gray-800 text-white'
                : 'text-gray-400 hover:bg-gray-900 hover:text-white'
            }`}
          >
            <span>{item.icon}</span>
            <span>{item.label}</span>
          </Link>
        ))}
      </nav>
    </aside>
  )
}
